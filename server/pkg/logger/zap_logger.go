package logger

import (
	"fmt"
	"os"
	"path"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ log.Logger = (*ZapLogger)(nil)

var stdLogger *zap.Logger

func Init(logPath string) {
	loggerPath := path.Join(logPath, "api"+".log")
	w := zapcore.AddSync(&lumberjack.Logger{
		Filename:   loggerPath,
		MaxSize:    128, //MB
		MaxBackups: 3,
		MaxAge:     14, //days
		LocalTime:  true,
		Compress:   true,
	})

	config := zap.NewProductionEncoderConfig()
	config.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05")
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(config),
		// zapcore.AddSync(os.Stdout),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout), w),
		zapcore.Level(zap.InfoLevel),
	)
	stdLogger = zap.New(core,
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zap.FatalLevel),
	)
}

func SugaredLogger() *zap.SugaredLogger {
	return stdLogger.Sugar()
}

func StdLogger() *zap.Logger {
	return stdLogger
}

// ZapLogger is a logger impl.
type ZapLogger struct {
	Zlog *zap.Logger
	Sync func() error
}

// NewZapLogger return a zap logger.
func NewZapLogger(zapLogger *zap.Logger) *ZapLogger {
	return &ZapLogger{Zlog: zapLogger, Sync: zapLogger.Sync}
}

// Log Implementation of logger interface.
func (l *ZapLogger) Log(level log.Level, keyvals ...interface{}) error {
	if len(keyvals) == 0 || len(keyvals)%2 != 0 {
		l.Zlog.Warn(fmt.Sprint("Keyvalues must appear in pairs: ", keyvals))
		return nil
	}
	// Zap.Field is used when keyvals pairs appear
	var data []zap.Field
	for i := 0; i < len(keyvals); i += 2 {
		data = append(data, zap.Any(fmt.Sprint(keyvals[i]), fmt.Sprint(keyvals[i+1])))
	}
	switch level {
	case log.LevelDebug:
		l.Zlog.Debug("", data...)
	case log.LevelInfo:
		l.Zlog.Info("", data...)
	case log.LevelWarn:
		l.Zlog.Warn("", data...)
	case log.LevelError:
		l.Zlog.Error("", data...)
	case log.LevelFatal:
		l.Zlog.Fatal("", data...)
	}
	return nil
}
