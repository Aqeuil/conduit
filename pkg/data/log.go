package data

import (
	"fmt"
	"io"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	XesLogTraceIDKey    = "x_trace_id"
	XesLogServerIPKey   = "x_server_ip"
	XesLogTimestampKey  = "x_timestamp"
	XesLogTimeKey       = "x_logdate"
	XesLogLevelKey      = "x_level"
	XesLogNameKey       = "x_logger"
	XesLogMessageKey    = ""
	XesLogCallerKey     = "x_source"
	XesLogStacktraceKey = "x_stack"

	DefaultLevel      = "debug"
	DefaultFileName   = "./logs/log.log"
	DefaultMaxSize    = 100
	DefaultMaxAge     = 30
	DefaultMaxBackups = 10
	DefaultCompress   = false
)

type Option func(*config)

type XLogger struct {
	log        *zap.Logger
	Sync       func() error
	messageKey string
}

type config struct {
	Level      string
	FileName   string
	MaxSize    int
	MaxAge     int
	MaxBackups int
	Compress   bool
	MessageKey string
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

func WithMessageKey(messageKey string) Option {
	return func(opts *config) {
		opts.MessageKey = messageKey
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
		MessageKey: XesLogMessageKey,
	}

	var zapOption []zap.Option
	zapOption = append(zapOption,
		zap.AddStacktrace(zap.NewAtomicLevelAt(zapcore.ErrorLevel)),
		zap.AddCaller(),
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

	return &XLogger{log: zapLogger, Sync: zapLogger.Sync, messageKey: logConf.MessageKey}
}

// Log Implementation of kratos logger interface
func (l *XLogger) Log(level zapcore.Level, keyvals ...interface{}) error {
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
		zap.String(XesLogServerIPKey, hostname),
		zap.Int64(XesLogTimestampKey, time.Now().Unix()),
	)

	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case zap.DebugLevel:
		l.log.Debug("", data...)
	case zap.InfoLevel:
		l.log.Info("", data...)
	case zap.WarnLevel:
		l.log.Warn("", data...)
	case zap.ErrorLevel:
		l.log.Error("", data...)
	}
	return nil
}

func (l *XLogger) Infof(format string, keyvals ...interface{}) {
	_ = l.Log(zap.InfoLevel, l.messageKey, fmt.Sprintf(format, keyvals...))
}
func (l *XLogger) Infow(keyvals ...interface{}) {
	_ = l.Log(zap.InfoLevel, keyvals...)
}

func (l *XLogger) Warnf(format string, keyvals ...interface{}) {
	_ = l.Log(zap.WarnLevel, l.messageKey, fmt.Sprintf(format, keyvals...))
}

func (l *XLogger) Warnw(keyvals ...interface{}) {
	_ = l.Log(zap.WarnLevel, keyvals)
}

func (l *XLogger) Errorf(format string, keyvals ...interface{}) {
	_ = l.Log(zap.ErrorLevel, l.messageKey, fmt.Sprintf(format, keyvals...))
}

func (l *XLogger) Errorw(keyvals ...interface{}) {
	_ = l.Log(zap.ErrorLevel, keyvals...)
}

func (l *XLogger) Debugf(format string, keyvals ...interface{}) {
	_ = l.Log(zap.DebugLevel, l.messageKey, fmt.Sprintf(format, keyvals...))
}

func (l *XLogger) Debugw(keyvals ...interface{}) {
	_ = l.Log(zap.DebugLevel, keyvals...)
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
