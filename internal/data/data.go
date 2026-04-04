package data

import (
	"conduit/internal/biz/matcher"
	"conduit/internal/conf"
	"conduit/pkg/data"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-pg/pg/v10"
	"github.com/google/wire"
	"github.com/spf13/cast"
)

// ProviderSet is data providers.
var ProviderSet = wire.NewSet(
	NewData,
	NewEtcdClient,
	NewUnitWatcher,
	wire.Bind(new(matcher.Watcher), new(*UnitWatcher)),
)

// Data .
type Data struct {
	Data *pg.DB
}

// NewData .
func NewData(
	c *conf.Data,
	logger log.Logger,
) (*Data, func(), error) {
	l := log.NewHelper(logger)

	db := NewPgSql(c.Data)

	d := &Data{
		Data: db,
	}

	cleanup := func() {
		l.Info("closing the data resources")
		if err := d.Data.Close(); err != nil {
			l.Error(err)
		}
	}

	return d, cleanup, nil
}

func NewPgSql(p *conf.PgSql) *pg.DB {
	if p == nil {
		panic("pgsql configuration is nil")
	}

	opts, err := pg.ParseURL(p.Source)
	if err != nil {
		panic("failed to parse postgres DSN: " + err.Error())
	}

	opts.MaxConnAge = p.ConnMaxLifetime.AsDuration()
	opts.IdleTimeout = p.ConnMaxIdleTime.AsDuration()
	opts.PoolSize = int(p.MaxOpenConns)
	opts.MinIdleConns = int(p.MaxIdleConns)

	db := pg.Connect(opts)
	if _, err = db.Exec("SELECT 1"); err != nil {
		panic("failed to ping postgres: " + err.Error())
	}
	return db
}

func NewRedis(c *conf.Redis, logger log.Logger) *data.LoggingConn {
	logDb := log.NewHelper(log.With(logger, "x_module", "repository/redis"))

	pool, err := data.NewLoggingPool(
		&data.Config{
			Addr:               c.Addr,
			Auth:               c.Auth,
			SelectDb:           cast.ToInt(c.SelectDb),
			MaxIdleConns:       cast.ToInt(c.MaxIdleConns),
			IdleTimeout:        c.IdleTimeout.AsDuration(),
			DialConnectTimeout: c.DialConnectTimeout.AsDuration(),
			DialReadTimeout:    c.DialReadTimeout.AsDuration(),
			DialWriteTimeout:   c.DialWriteTimeout.AsDuration(),
		}, logger)
	if err != nil {
		logDb.Fatalf("failed dbrw use to redis: %v", err)
		panic(err)
	}

	logDb.Info("init redis")
	return pool
}
