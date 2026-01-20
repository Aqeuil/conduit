package data

import (
	"context"
	json2 "encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/google/uuid"
	"github.com/spf13/cast"
)

const (
	MaxReplay       = 200
	MaxTimeDuration = time.Second * 5
)

type LoggingConn struct {
	Pool   *redis.Pool
	logger *XLogger
}

type logAttr struct {
	method    string
	command   string
	args      []interface{}
	reply     interface{}
	startTime time.Time

	err    error
	shrink time.Duration
}

type Config struct {
	Addr               string
	Auth               string
	SelectDb           int
	MaxIdleConns       int
	IdleTimeout        time.Duration
	DialConnectTimeout time.Duration
	DialReadTimeout    time.Duration
	DialWriteTimeout   time.Duration
}

// NewLoggingPool returns a logging wrapper around a connection.
func NewLoggingPool(config *Config, logger *XLogger) (*LoggingConn, error) {
	redisPool := &redis.Pool{
		MaxIdle:     config.MaxIdleConns,
		IdleTimeout: config.IdleTimeout,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config.Addr,
				redis.DialConnectTimeout(config.DialConnectTimeout),
				redis.DialReadTimeout(config.DialReadTimeout),
				redis.DialWriteTimeout(config.DialWriteTimeout),
			)
			if err != nil {
				return nil, err
			}

			if len(config.Auth) != 0 {
				if _, err := c.Do("AUTH", config.Auth); err != nil {
					c.Close()
					return nil, err
				}
			}

			if _, err := c.Do("SELECT", config.SelectDb); err != nil {
				c.Close()
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if t.Add(config.IdleTimeout).After(time.Now()) {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	if _, err := redisPool.Get().Do("PING"); err != nil {
		return nil, err
	}

	return &LoggingConn{
		Pool:   redisPool,
		logger: logger,
	}, nil
}

func stringifyReply(reply any) string {
	str, _ := json2.Marshal(reply)
	return string(str)
}

func (c *LoggingConn) log(ctx context.Context, attr logAttr) {
	var (
		x_action   string
		x_params   string
		x_shrink   float64
		x_response string
		x_error    string
	)

	x_action = strings.Join([]string{attr.method, attr.command}, ".")
	x_shrink = attr.shrink.Seconds()

	replyStr := cast.ToString(attr.reply)
	if len(replyStr) > MaxReplay {
		x_response = replyStr[:MaxReplay] + "..."
	} else {
		x_response = replyStr
	}
	if attr.err != nil {
		x_error = attr.err.Error()
	}
	if len(attr.args) > 0 {
		strList := append([]string{attr.command}, cast.ToStringSlice(attr.args)...)
		x_params = strings.Join(strList, " ")
	}

	c.logger.Debugw(
		"x_action", x_action,
		"x_param", x_params,
		"x_response", x_response,
		"x_shrink", x_shrink,
		"x_error", x_error,
		"x_duration", time.Since(attr.startTime).Seconds(),
	)
}

func (c *LoggingConn) do(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	startTime := time.Now()
	reply, err := conn.Do(commandName, args...)
	c.log(ctx, logAttr{
		method:    "do",
		startTime: startTime,
		command:   commandName,
		args:      args,
		reply:     reply,
		err:       err,
	})
	return reply, err
}

func (c *LoggingConn) doContext(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	startTime := time.Now()
	reply, err := redis.DoContext(conn, ctx, commandName, args...)
	c.log(ctx, logAttr{
		method:    "doContext",
		startTime: startTime,
		command:   commandName,
		args:      args,
		reply:     reply,
		err:       err,
	})
	return reply, err
}

func (c *LoggingConn) DoWithTimeout(ctx context.Context, commandName string, args ...interface{}) (interface{}, error) {
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		if err == redis.ErrPoolExhausted {
			c.logger.Errorf("DoWithTimeout.%s failed, %s", commandName, err.Error())
		}
		return nil, err
	}
	defer conn.Close()

	startTime := time.Now()
	reply, err := redis.DoWithTimeout(conn, MaxTimeDuration, commandName, args...)
	c.log(ctx, logAttr{
		method:    "DoWithTimeout",
		startTime: startTime,
		command:   commandName,
		args:      args,

		reply:  reply,
		err:    err,
		shrink: MaxTimeDuration,
	})
	return reply, err
}

/*************************业务使用*******************/

func (c *LoggingConn) IsErrNil(err error) bool {
	return err == redis.ErrNil
}

func (c *LoggingConn) TTl(ctx context.Context, key string) int {
	times, _ := redis.Int(c.DoWithTimeout(ctx, "TTL", key))
	return times
}

func (c *LoggingConn) Expire(ctx context.Context, key string, time time.Duration) error {
	seconds := int(time.Seconds())
	var err error
	if seconds > 0 {
		_, err = redis.String(c.DoWithTimeout(ctx, "EXPIRE", key, int(time.Seconds())))
	}

	return err
}

func (c *LoggingConn) ExpireIfMatch(ctx context.Context, key string, value string, duration time.Duration) (bool, error) {
	seconds := int(duration.Seconds())

	luaScript := `
        if redis.call("GET", KEYS[1]) == ARGV[1] then
            return redis.call("EXPIRE", KEYS[1], ARGV[2])
        else
            return 0
        end
    `

	result, err := redis.Int(c.DoWithTimeout(ctx, "EVAL", luaScript, 1, key, value, seconds))
	if err != nil {
		return false, err
	}

	return result == 1, nil
}

func (c *LoggingConn) PipelineExpire(ctx context.Context, keys []string, duration time.Duration) error {
	seconds := int(duration.Seconds())
	if seconds <= 0 {
		return nil
	}
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return err
	}

	var args []any
	conn.Send("MULTI")
	for _, key := range keys {
		if len(key) == 0 {
			continue
		}
		args = append(args, key)
	}
	_, err = conn.Do("EXEC")
	conn.Close()

	c.log(ctx, logAttr{
		method:    "PipelineExpire",
		startTime: startTime,
		command:   "MULTI EXPIRE",
		args:      args,
		reply:     "",
		err:       err,
	})
	return err
}

func (c *LoggingConn) GetString(ctx context.Context, key string) (string, error) {
	res, err := redis.String(c.DoWithTimeout(ctx, "GET", key))
	return res, err
}

func (c *LoggingConn) GetInt(ctx context.Context, key string) (int, error) {
	res, err := redis.Int(c.DoWithTimeout(ctx, "GET", key))
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (c *LoggingConn) Incr(ctx context.Context, key string) (int, error) {
	res, err := redis.Int(c.DoWithTimeout(ctx, "INCR", key))
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (c *LoggingConn) Hincr(ctx context.Context, key string, field any) (int, error) {
	res, err := redis.Int(c.DoWithTimeout(ctx, "HINCRBY", key, field, 1))
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (c *LoggingConn) MGetInt(ctx context.Context, keys ...string) ([]int, error) {
	return redis.Ints(c.DoWithTimeout(ctx, "MGET", keys))
}

func (c *LoggingConn) MGetString(ctx context.Context, keys ...string) ([]string, error) {
	return redis.Strings(c.DoWithTimeout(ctx, "MGET", keys))
}

func (c *LoggingConn) Set(ctx context.Context, key string, val interface{}) error {
	_, err := c.DoWithTimeout(ctx, "SET", key, val)
	return err
}

func (c *LoggingConn) SetEX(ctx context.Context, key string, val interface{}, duration time.Duration) error {
	var err error
	seconds := int(duration.Seconds())
	if seconds > 0 {
		_, err = c.DoWithTimeout(ctx, "SET", key, val, "EX", seconds)
	} else {
		_, err = c.DoWithTimeout(ctx, "SET", key, val)
	}
	return err
}

func (c *LoggingConn) SetBit(ctx context.Context, key string, offset int, val int) error {
	_, err := c.DoWithTimeout(ctx, "SETBIT", key, offset, val)
	return err
}

func (c *LoggingConn) SetBitEX(ctx context.Context, key string, offset int, val int, duration time.Duration) error {
	var err error
	seconds := int(duration.Seconds())
	if seconds > 0 {
		_, err = c.DoWithTimeout(ctx, "SETBIT", key, offset, val, "EX", seconds)
	} else {
		_, err = c.DoWithTimeout(ctx, "SETBIT", key, offset, val)
	}
	return err
}

func (c *LoggingConn) GetBit(ctx context.Context, key string, offset int) (int, error) {
	res, err := redis.Int(c.DoWithTimeout(ctx, "GETBIT", key, offset))
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (c *LoggingConn) BitCount(ctx context.Context, key string) (int, error) {
	res, err := redis.Int(c.DoWithTimeout(ctx, "BITCOUNT", key))
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (c *LoggingConn) MSet(ctx context.Context, args ...interface{}) error {
	_, err := c.DoWithTimeout(ctx, "MSET", args)
	return err
}

func (c *LoggingConn) MSetEx(ctx context.Context, duration time.Duration, values map[string]any) error {
	seconds := int(duration.Seconds())
	var args []any
	var keys []string
	for k, v := range values {
		keys = append(keys, k)
		args = append(args, k, v)
	}

	err := c.MSet(ctx, args)
	if err == nil {
		if seconds > 0 {
			err = c.PipelineExpire(ctx, keys, duration)
		}
	}
	return err
}

func (c *LoggingConn) SetNXWithExpire(ctx context.Context, key, value string, duration time.Duration) (bool, error) {
	seconds := int(duration.Seconds())
	var err error
	var ok interface{}
	if seconds > 0 {
		ok, err = c.DoWithTimeout(ctx, "SET", key, value, "EX", seconds, "NX")
	} else {
		ok, err = c.DoWithTimeout(ctx, "SET", key, value)
	}

	if ok == nil {
		return false, err
	}

	return true, err
}

func (c *LoggingConn) HSet(ctx context.Context, key string, field, value interface{}) error {
	_, err := c.DoWithTimeout(ctx, "HSET", key, field, value)
	return err
}

func (c *LoggingConn) HMSet(ctx context.Context, key string, value interface{}) error {
	_, err := c.DoWithTimeout(ctx, "HMSET", redis.Args{}.Add(key).AddFlat(value)...)
	return err
}

func (c *LoggingConn) HMSetWithDuration(ctx context.Context, key string, value interface{}, duration time.Duration) error {
	_, err := c.DoWithTimeout(ctx, "HMSET", redis.Args{}.Add(key).AddFlat(value)...)
	if err == nil {
		c.Expire(ctx, key, duration)
	}
	return err
}

func (c *LoggingConn) HDel(ctx context.Context, key string, field interface{}) error {
	_, err := c.DoWithTimeout(ctx, "HDEL", key, field)
	return err
}

func (c *LoggingConn) HGetString(ctx context.Context, key string, field interface{}) (string, error) {
	res, err := redis.String(c.DoWithTimeout(ctx, "HGET", key, field))
	return res, err
}

func (c *LoggingConn) HMGETString(ctx context.Context, key string, field interface{}) ([]string, error) {
	reply, err := c.DoWithTimeout(ctx, "HMGET", redis.Args{}.Add(key).AddFlat(field)...)
	return redis.Strings(reply, err)
}

func (c *LoggingConn) HMGET(ctx context.Context, key string, field interface{}) (interface{}, error) {
	reply, err := c.DoWithTimeout(ctx, "HMGET", redis.Args{}.Add(key).AddFlat(field)...)
	return reply, err
}

func (c *LoggingConn) HLen(ctx context.Context, key string) (int, error) {
	reply, err := redis.Int(c.DoWithTimeout(ctx, "HLen", key))
	return reply, err
}

func (c *LoggingConn) EXISTS(ctx context.Context, key string) (int, error) {
	reply, err := redis.Int(c.DoWithTimeout(ctx, "EXISTS", key))
	return reply, err
}

func (c *LoggingConn) HGetAll(ctx context.Context, key string, dest interface{}) error {
	reply, err := c.DoWithTimeout(ctx, "HGETALL", key)
	values, err := redis.Values(reply, err)
	if err != nil {
		return err
	}
	if len(values) == 0 {
		return redis.ErrNil
	}
	return redis.ScanStruct(values, dest)
}

func (c *LoggingConn) HGetAllMap(ctx context.Context, key string) (map[string]string, error) {
	return redis.StringMap(c.DoWithTimeout(ctx, "HGETALL", key))
}

func (c *LoggingConn) SAdd(ctx context.Context, key string, value []interface{}) error {
	_, err := c.DoWithTimeout(ctx, "SADD", redis.Args{}.Add(key).AddFlat(value)...)
	return err
}

func (c *LoggingConn) SREM(ctx context.Context, key string, value []interface{}) error {
	_, err := c.DoWithTimeout(ctx, "SREM", redis.Args{}.Add(key).AddFlat(value)...)
	return err
}

func (c *LoggingConn) Sismember(ctx context.Context, key string, value string) (bool, error) {
	b, err := redis.Bool(c.DoWithTimeout(ctx, "SISMEMBER", key, value))
	return b, err
}

func (c *LoggingConn) SCard(ctx context.Context, key string) (int64, error) {
	return redis.Int64(c.DoWithTimeout(ctx, "SCard", key))
}

func (c *LoggingConn) SCardInt(ctx context.Context, key string) ([]int, error) {
	return redis.Ints(c.DoWithTimeout(ctx, "SCard", key))
}

func (c *LoggingConn) SCardString(ctx context.Context, key string) ([]string, error) {
	return redis.Strings(c.DoWithTimeout(ctx, "SCard", key))
}

func (c *LoggingConn) ZAdd(ctx context.Context, key string, score int64, value string) error {
	_, err := c.DoWithTimeout(ctx, "ZADD", key, key, score, value)
	return err
}

func (c *LoggingConn) ZRank(ctx context.Context, key string, value string) (int, error) {
	return redis.Int(c.DoWithTimeout(ctx, "ZRANK", key, value))
}

func (c *LoggingConn) Del(ctx context.Context, key string) error {
	_, err := c.DoWithTimeout(ctx, "DEL", key)
	return err
}

func (c *LoggingConn) MDel(ctx context.Context, keys []string) (err error) {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	conn.Send("MULTI")
	var args []any
	for _, key := range keys {
		if len(key) == 0 {
			continue
		}
		args = append(args, key)
		conn.Send("DEL", strings.Trim(key, " "))
	}
	_, err = conn.Do("EXEC")

	c.log(ctx, logAttr{
		method:    "MDel",
		startTime: startTime,
		command:   "MULTI DEL",
		args:      args,
		reply:     "",
		err:       err,
	})
	return
}

func (c *LoggingConn) ZRevRange(ctx context.Context, key string, scoreStart, scoreEnd int) (map[string]string, error) {
	byteList, err := redis.Strings(c.DoWithTimeout(ctx, "ZREVRANGE", key, scoreStart, scoreEnd, "WITHSCORES"))
	if err != nil {
		return nil, err
	}

	list := make(map[string]string, len(byteList)/2)
	for k, v := range byteList {
		if k%2 != 0 {
			continue
		}
		list[v] = byteList[k+1]
	}
	return list, nil
}

func (c *LoggingConn) PipelineZRange(ctx context.Context, scoreStart, scoreEnd interface{}, keys ...string) (map[string][]string, error) {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var args []any
	for _, key := range keys {
		args = append(args, key)
		_ = conn.Send("ZRANGEBYSCORE", key, scoreStart, scoreEnd)
	}

	err = conn.Flush()

	container := make(map[string][]string, len(keys))
	if err == nil {
		for _, key := range keys {
			reply, err := redis.Strings(redis.ReceiveContext(conn, ctx))
			if err != nil {
				continue
			}
			container[key] = reply
		}
	}

	c.log(ctx, logAttr{
		method:    "PipelineZRange",
		startTime: startTime,
		command:   "ZRANGEBYSCORE",
		args:      args,
		reply:     stringifyReply(container),
		err:       err,
	})
	return container, nil
}

// PipelineHGetAll
// //key与 对应空结构一一对应
//
//	reply, err := redisConn.PipelineHgetAll(context.Background(), []string{"promotionv3_info_17577", "promotionv3_info_17578"}, map[string]interface{}{
//		"promotionv3_info_17577": &user{},
//		"promotionv3_info_17578": &user{},
//		"promotionv3_info_17579": &user{},
//	})
func (c *LoggingConn) PipelineHGetAll(ctx context.Context, keys []string, keyMapContainer map[string]interface{}) (map[string]interface{}, error) {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		if err == redis.ErrPoolExhausted {
			c.logger.Errorf("PipelineHGetAll failed, %s", err.Error())
		}
		return nil, err
	}
	defer conn.Close()
	for _, key := range keys {
		_ = conn.Send("HGETALL", key)
	}
	if err := conn.Flush(); err != nil {
		return nil, err
	}

	var args []any
	containers := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		args = append(args, key)
		values, err := redis.Values(redis.ReceiveContext(conn, ctx))
		if err != nil {
			c.logger.Errorf("Redis PipelineHGetAll error %v", err)
			continue
		}

		if len(values) == 0 {
			c.logger.Infof("Redis PipelineHGetAll value empty, key: %v", key)
			continue
		}

		if _, ok := keyMapContainer[key].(map[string]string); ok {
			ret, errMap := redis.StringMap(values, err)
			if errMap != nil {
				c.logger.Errorf("PipelineHGetAll.StringMap failed, %s", err)
				continue
			}
			containers[key] = ret
		} else {
			errStruct := redis.ScanStruct(values, keyMapContainer[key])
			if errStruct != nil {
				c.logger.Errorf("PipelineHGetAll.ScanStruct failed, %s", err)
				continue
			}
			containers[key] = keyMapContainer[key]
		}
	}

	c.log(ctx, logAttr{
		method:    "PipelineHGetAll",
		startTime: startTime,
		command:   "HGETALL",
		args:      args,
		reply:     stringifyReply(containers),
		err:       err,
	})
	return containers, nil
}

func (c *LoggingConn) PipelineHMSet(ctx context.Context, setData map[string]interface{}) error {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var args []any
	for key, val := range setData {
		args = append(args, key)
		_ = conn.Send("HMSET", redis.Args{}.Add(key).AddFlat(val)...)
	}

	err = conn.Flush()
	c.log(ctx, logAttr{
		method:    "PipelineHMSet",
		startTime: startTime,
		command:   "Pipeline",
		args:      args,
		reply:     "",
		err:       err,
	})
	return err
}

func (c *LoggingConn) PipelineHMSetWithDuration(ctx context.Context, setData map[string]interface{}, duration time.Duration) error {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var args []any
	seconds := int(duration.Seconds())
	for key, val := range setData {
		_ = conn.Send("HMSET", redis.Args{}.Add(key).AddFlat(val)...)
		if seconds > 0 {
			_ = conn.Send("EXPIRE", redis.Args{}.Add(key).Add(seconds)...)
		}

		if seconds > 0 {
			args = append(args, fmt.Sprintf("HMSET %s EXPIRE %d", key, seconds))
		} else {
			args = append(args, fmt.Sprintf("HMSET %s", key))
		}
	}

	err = conn.Flush()
	c.log(ctx, logAttr{
		method:    "PipelineHMSetWithDuration",
		startTime: startTime,
		command:   "Pipeline",
		args:      args,
		reply:     "",
		err:       err,
	})
	return err
}

func (c *LoggingConn) PipelineHMSetWithDurations(ctx context.Context, setData map[string]interface{}, durations map[string]time.Duration) error {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	var args []any
	for key, val := range setData {
		_ = conn.Send("HMSET", redis.Args{}.Add(key).AddFlat(val)...)
		seconds := 0
		if duration, ok := durations[key]; ok {
			seconds = int(duration.Seconds())
			if seconds > 0 {
				_ = conn.Send("EXPIRE", redis.Args{}.Add(key).Add(seconds)...)
			}
		}

		if seconds > 0 {
			args = append(args, fmt.Sprintf("HMSET %s EXPIRE %d", key, seconds))
		} else {
			args = append(args, fmt.Sprintf("HMSET %s", key))
		}
	}

	err = conn.Flush()
	c.log(ctx, logAttr{
		method:    "PipelineHMSetWithDurations",
		startTime: startTime,
		command:   "Pipeline",
		args:      args,
		reply:     "",
		err:       err,
	})
	return err
}

func (c *LoggingConn) PipelineHGetField(ctx context.Context, keyList []string, field string) map[string]interface{} {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		if err == redis.ErrPoolExhausted {
			c.logger.Errorf("PipelineHGetAll failed, %s", err.Error())
		}
		return nil
	}
	defer conn.Close()
	for _, key := range keyList {
		_ = conn.Send("HGet", key, field)
	}
	if err := conn.Flush(); err != nil {
		return nil
	}

	var args []any
	containers := make(map[string]interface{}, len(keyList))
	for _, key := range keyList {
		args = append(args, key)
		val, err := conn.Receive()
		if err != nil {
			c.logger.Errorf("Redis PipelineHGetAll error %v", err)
			continue
		}

		if val != nil {
			containers[key] = string(val.([]byte))
		}
	}

	c.log(ctx, logAttr{
		method:    "PipelineHGetField",
		startTime: startTime,
		command:   "HGETField",
		args:      args,
		reply:     stringifyReply(containers),
		err:       err,
	})
	return containers
}

func (c *LoggingConn) PipelineHGetAllMap(ctx context.Context, keys []string) (map[string]interface{}, error) {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		if err == redis.ErrPoolExhausted {
			c.logger.Errorf("PipelineHGetAllMap failed, %s", err.Error())
		}
		return nil, err
	}
	defer conn.Close()
	for _, key := range keys {
		_ = conn.Send("HGETALL", key)
	}
	if err := conn.Flush(); err != nil {
		return nil, err
	}

	var args []any
	containers := make(map[string]interface{}, len(keys))
	for _, key := range keys {
		args = append(args, key)
		values, err := redis.Values(redis.ReceiveContext(conn, ctx))
		if err != nil {
			c.logger.Errorf("Redis PipelineHGetAllMap error %v", err)
			continue
		}

		if len(values) == 0 {
			c.logger.Infof("Redis PipelineHGetAllMap value empty, key: %v", key)
			continue
		}
		stringMap, _ := redis.StringMap(values, err)
		containers[key] = stringMap
	}

	c.log(ctx, logAttr{
		method:    "PipelineHGetAllMap",
		startTime: startTime,
		command:   "HGETALL",
		args:      args,
		reply:     stringifyReply(containers),
		err:       err,
	})
	return containers, nil
}

func (c *LoggingConn) PipelineBitCount(ctx context.Context, keyList []string) map[string]int {
	startTime := time.Now()
	conn, err := c.Pool.GetContext(ctx)
	if err != nil {
		if err == redis.ErrPoolExhausted {
			c.logger.Errorf("PipelineBitCount failed, %s", err.Error())
		}
		return nil
	}
	defer conn.Close()
	for _, key := range keyList {
		_ = conn.Send("BITCOUNT", key)
	}
	if err := conn.Flush(); err != nil {
		return nil
	}

	var args []any
	containers := make(map[string]int, len(keyList))
	for _, key := range keyList {
		args = append(args, key)
		val, err := conn.Receive()
		if err != nil {
			c.logger.Errorf("Redis PipelineBitCount error %v", err)
			continue
		}

		if val != nil {
			containers[key] = cast.ToInt(val.(int64))
		}
	}

	c.log(ctx, logAttr{
		method:    "PipelineBitCount",
		startTime: startTime,
		command:   "BITCOUNT",
		args:      args,
		reply:     stringifyReply(containers),
		err:       err,
	})
	return containers
}

func (c *LoggingConn) LLen(ctx context.Context, key string) (int, error) {
	reply, err := redis.Int(c.DoWithTimeout(ctx, "LLEN", key))
	return reply, err
}

func (c *LoggingConn) LPop(ctx context.Context, key string) (string, error) {
	reply, err := redis.String(c.DoWithTimeout(ctx, "LPOP", key))
	return reply, err
}

func (c *LoggingConn) LPush(ctx context.Context, key string, value []interface{}) error {
	_, err := c.DoWithTimeout(ctx, "LPUSH", redis.Args{}.Add(key).AddFlat(value)...)
	return err
}

func (c *LoggingConn) SetNxLock(ctx context.Context, key, value string, duration time.Duration) (bool, error) {
	seconds := int(duration.Seconds())
	if seconds <= 0 {
		seconds = 3
	}
	var (
		str string
		err error
	)
	for i := 0; i < 20; i++ {
		result, seterr := c.DoWithTimeout(ctx, "SET", key, value, "EX", seconds, "NX")
		str, seterr = redis.String(result, seterr)
		if strings.ToUpper(str) == "OK" {
			break
		}
		time.Sleep(time.Millisecond * 100)
	}
	return strings.ToUpper(str) == "OK", err
}

// DelNxLock returns (是否正确删除, 是否出现错误)
func (c *LoggingConn) DelNxLock(ctx context.Context, key, value string) (bool, error) {
	res, err := redis.String(c.DoWithTimeout(ctx, "GET", key))
	if err != nil {
		return false, err
	}
	if res == value {
		_, err = c.DoWithTimeout(ctx, "DEL", key)
		return true, err
	}
	return false, nil
}

// RedLock 加锁, 定期续期
func (c *LoggingConn) RedLock(ctx context.Context, key string, duration time.Duration) (bool, func() error, error) {
	value := uuid.New().String()
	ok, err := c.SetNXWithExpire(ctx, key, value, duration)
	if err != nil {
		return false, nil, err
	}
	if !ok {
		return false, nil, nil
	}

	renewCtx, cancelRenew := context.WithCancel(context.Background())
	go c.renewLock(renewCtx, key, value, duration)

	return true, func() error {
		cancelRenew()

		_, err1 := c.UnLock(ctx, key, value)
		return err1
	}, nil
}

// renewLock 定期续期锁
func (c *LoggingConn) renewLock(ctx context.Context, key, value string, duration time.Duration) {
	ticker := time.NewTicker(duration / 2)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// 只有value匹配时才续期
			ok, err := c.ExpireIfMatch(ctx, key, value, duration)
			if err != nil {
				return
			}
			if !ok {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}

// UnLock 解锁
func (c *LoggingConn) UnLock(ctx context.Context, key, value string) (bool, error) {
	return c.DelNxLock(ctx, key, value)
}
