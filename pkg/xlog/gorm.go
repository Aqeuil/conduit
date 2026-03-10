package xlog

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/pkg/errors"
	gormLog "gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

type Gorm struct {
	gormLog.Config

	logGrom *log.Helper
}

func NewGorm(logger log.Logger) *Gorm {
	return &Gorm{
		logGrom: log.NewHelper(log.With(logger, "x_module", "pkg/gorm")),
		Config: gormLog.Config{
			SlowThreshold: time.Millisecond * 500,
			LogLevel:      gormLog.Info,
			Colorful:      false,
		},
	}
}

// LogMode log mode
func (l *Gorm) LogMode(level gormLog.LogLevel) gormLog.Interface {
	newlogger := *l
	newlogger.LogLevel = level
	return &newlogger
}

// Info print info
func (l *Gorm) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLog.Info {
		l.logGrom.WithContext(ctx).Info(append([]interface{}{msg, utils.FileWithLineNum()}, data...)...)
	}
}

// Warn print warn messages
func (l *Gorm) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLog.Warn {
		l.logGrom.WithContext(ctx).Warn(append([]interface{}{msg, utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l *Gorm) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= gormLog.Error {
		l.logGrom.WithContext(ctx).Error(append([]interface{}{msg, utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l *Gorm) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.LogLevel <= gormLog.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= gormLog.Error && !errors.Is(err, gormLog.ErrRecordNotFound):
		sql, rows := fc()
		if rows == -1 {
			l.logGrom.WithContext(ctx).Errorw("x_file", utils.FileWithLineNum(), "x_error", err, "x_duration", elapsed.Seconds(), "x_rows", "-", "x_action", sql)
		} else {
			l.logGrom.WithContext(ctx).Errorw("x_file", utils.FileWithLineNum(), "x_error", err, "x_duration", elapsed.Seconds(), "x_rows", rows, "x_action", sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= gormLog.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL: %v", l.SlowThreshold)
		if rows == -1 {
			l.logGrom.WithContext(ctx).Errorw("x_file", utils.FileWithLineNum(), "x_error", slowLog, "x_duration", elapsed.Seconds(), "x_rows", "-", "x_action", sql)
		} else {
			l.logGrom.WithContext(ctx).Errorw("x_file", utils.FileWithLineNum(), "x_error", slowLog, "x_duration", elapsed.Seconds(), "x_rows", rows, "x_action", sql)
		}
	case l.LogLevel == gormLog.Info:
		sql, rows := fc()
		if rows == -1 {
			l.logGrom.WithContext(ctx).Infow("x_file", utils.FileWithLineNum(), "x_duration", elapsed.Seconds(), "x_rows", "-", "x_action", sql)
		} else {
			l.logGrom.WithContext(ctx).Infow("x_file", utils.FileWithLineNum(), "x_duration", elapsed.Seconds(), "x_rows", rows, "x_action", sql)
		}
	}
}
