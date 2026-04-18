package xlog

import (
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/go-kratos/kratos/v2/log"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	XesLogTraceIDKey    = "x_trace_id"
	XesLogRpcIDKey      = "x_rpc_id"
	XesLogDeviceIDKey   = "x_devid"
	XesLogServerIPKey   = "x_server_ip"
	XesLogDateKey       = "x_date"
	XesLogTimestampKey  = "x_timestamp"
	XesLogTimeKey       = "x_logdate"
	XesLogLevelKey      = "x_level"
	XesLogNameKey       = "x_logger"
	XesLogMessageKey    = "x_zap_msg"
	XesLogCallerKey     = "x_source"
	XesLogStacktraceKey = "x_stack"
	XesLogMsgKey        = "x_msg"
	XesLogOrgId         = "x_org_id"
	XesLogDepartmentKey = "x_department"
	XesLogVersionKey    = "x_version"

	DefaultLevel         = "debug"
	DefaultFileName      = "./logs/log.log"
	DefaultMaxSize       = 100
	DefaultMaxAge        = 30
	DefaultMaxBackups    = 10
	DefaultCompress      = false
	DefaultXesDepartment = "Xueersibook"
	DefaultXesVersion    = "0.1"
)

var _ log.Logger = (*XLogger)(nil)

type Option func(*config)

type XLogger struct {
	log  *zap.Logger
	Sync func() error
}

type config struct {
	Level      string
	FileName   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool

	XesDepartment string
	XesVersion    string
}

func WithLevel(level string) Option {
	return func(opts *config) {
		opts.Level = level
	}
}

func WithFileName(fileName string) Option {
	return func(opts *config) {
		opts.FileName = fileName
	}
}

func WithMaxSize(maxSize int) Option {
	return func(opts *config) {
		opts.MaxSize = maxSize
	}
}

func WithMaxAge(maxAge int) Option {
	return func(opts *config) {
		opts.MaxAge = maxAge
	}
}

func WithMaxBackups(maxBackups int) Option {
	return func(opts *config) {
		opts.MaxBackups = maxBackups
	}
}

func WithCompress(compress bool) Option {
	return func(opts *config) {
		opts.Compress = compress
	}
}

func NewXLogger(opts ...interface{}) *XLogger {
	logConf := &config{
		Level:      DefaultLevel,
		FileName:   DefaultFileName,
		MaxSize:    DefaultMaxSize,
		MaxAge:     DefaultMaxAge,
		MaxBackups: DefaultMaxBackups,
		Compress:   DefaultCompress,

		XesDepartment: DefaultXesDepartment,
		XesVersion:    DefaultXesVersion,
	}

	var zapOption []zap.Option
	zapOption = append(zapOption,
		zap.AddStacktrace(zap.NewAtomicLevelAt(zapcore.ErrorLevel)),
		zap.AddCaller(),
		zap.Fields(
			zap.String(XesLogDepartmentKey, logConf.XesDepartment),
			zap.String(XesLogVersionKey, logConf.XesVersion),
		),
	)
	for _, opt := range opts {
		if opt == nil {
			continue
		}
		switch opt.(type) {
		case Option:
			opt.(Option)(logConf)
		case zap.Option:
			zapOption = append(zapOption, opt.(zap.Option))
		default:
			panic("unknown option type")
		}
	}

	level := parseLogLevel(logConf)
	writeSyncer := zapcore.NewMultiWriteSyncer(zapcore.AddSync(getLumberJackLogWriter(logConf)))
	if level.Level() == zap.DebugLevel {
		writeSyncer = zapcore.NewMultiWriteSyncer(writeSyncer, zapcore.AddSync(os.Stdout))
	}

	core := zapcore.NewCore(zapcore.NewJSONEncoder(NewXesLogEncoder()), writeSyncer, level)
	zapLogger := zap.New(core, zapOption...)

	return &XLogger{log: zapLogger, Sync: zapLogger.Sync}
}

// NewHelper return log.Helper with message key is x_msg
func NewHelper(logger log.Logger) *log.Helper {
	return log.NewHelper(logger, log.WithMessageKey(XesLogMsgKey))
}

// Log Implementation of kratos logger interface
func (l *XLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.log.Warn(fmt.Sprint("Key values must appear in pairs: ", keyvals))
		return nil
	}

	// Zap.Field is used when keyvals pairs appear
	var data []zap.Field

	// 适配xes log 增加一些字段
	// https://cloud.tal.com/docs/logcenter/standard/golang.html#%E6%97%A5%E5%BF%97%E6%A0%BC%E5%BC%8F
	hostname, _ := os.Hostname()
	data = append(data,
		zap.String(XesLogDeviceIDKey, "-"),
		zap.String(XesLogServerIPKey, hostname),
		zap.String(XesLogDateKey, time.Now().Format("2006/01/02 15:04:05 MST")),
		zap.Int64(XesLogTimestampKey, time.Now().Unix()),
	)

	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.log.Debug("", data...)
	case log.LevelInfo:
		l.log.Info("", data...)
	case log.LevelWarn:
		l.log.Warn("", data...)
	case log.LevelError:
		l.log.Error("", data...)
	}
	return nil
}

// NewXesLogEncoder return zapcore.EncoderConfig with xes log format
func NewXesLogEncoder() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		TimeKey:        XesLogTimeKey,
		LevelKey:       XesLogLevelKey,
		NameKey:        XesLogNameKey,
		CallerKey:      XesLogCallerKey,
		MessageKey:     XesLogMessageKey,
		StacktraceKey:  XesLogStacktraceKey,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}

// getLumberJackLogWriter return io.Writer
func getLumberJackLogWriter(c *config) io.Writer {
	return &lumberjack.Logger{
		Filename:   c.FileName,
		MaxSize:    c.MaxSize,
		MaxBackups: c.MaxBackups,
		MaxAge:     c.MaxAge,
		Compress:   c.Compress,
	}
}

func parseLogLevel(c *config) (level zap.AtomicLevel) {
	switch c.Level {
	case "info":
		level = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "warn":
		level = zap.NewAtomicLevelAt(zap.WarnLevel)
	case "error":
		level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	default:
		level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	return
}

func OrgID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil || ctx.Value("org_id") == nil {
			return ""
		}
		return ctx.Value("org_id")
	}
}

func TraceID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil || ctx.Value(XesLogTraceIDKey) == nil {
			return ""
		}

		return ctx.Value(XesLogTraceIDKey)
	}
}

func RpcID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if ctx == nil || ctx.Value(XesLogRpcIDKey) == nil {
			return ""
		}
		return ctx.Value(XesLogRpcIDKey)
	}
}
