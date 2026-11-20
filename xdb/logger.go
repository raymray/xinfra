package xdb

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

func NewGormLogger(logger *zap.Logger, config gormlogger.Config) gormlogger.Interface {
	return &gormLogger{
		Logger: logger,
		Config: config,
	}
}

type gormLogger struct {
	*zap.Logger
	gormlogger.Config
}

func (gl *gormLogger) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	ngl := *gl
	ngl.LogLevel = level
	return &ngl
}

func (gl *gormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	gl.Logger.Sugar().Infof(msg, data...)
}

func (gl *gormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	gl.Logger.Sugar().Warnf(msg, data...)
}

func (gl *gormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	gl.Logger.Sugar().Errorf(msg, data...)
}

func (gl *gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	elapsedMs := float64(elapsed) / float64(time.Millisecond)
	slowMs := float64(gl.SlowThreshold) / float64(time.Millisecond)

	switch {
	case err != nil && gl.LogLevel >= gormlogger.Error && (!errors.Is(err, gorm.ErrRecordNotFound) || !gl.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			gl.Logger.Error(fmt.Sprintf("[%.3fms] [rows:%v] %s", elapsedMs, "-", sql), zap.Error(err))
		} else {
			gl.Logger.Error(fmt.Sprintf("[%.3fms] [rows:%v] %s", elapsedMs, rows, sql), zap.Error(err))
		}
	case elapsed > gl.SlowThreshold && gl.SlowThreshold != 0 && gl.LogLevel >= gormlogger.Warn:
		sql, rows := fc()
		if rows == -1 {
			gl.Logger.Warn(fmt.Sprintf("[SLOW SQL >= %.3fms] [%.3fms] [rows:%v] %s",
				slowMs, elapsedMs, "-", sql))
		} else {
			gl.Logger.Warn(fmt.Sprintf("[SLOW SQL >= %.3fms] [%.3fms] [rows:%v] %s",
				slowMs, elapsedMs, rows, sql))
		}
	case gl.LogLevel == gormlogger.Info:
		sql, rows := fc()
		if rows == -1 {
			gl.Logger.Info(fmt.Sprintf("[%.3fms] [rows:%v] %s", elapsedMs, "-", sql))
		} else {
			gl.Logger.Info(fmt.Sprintf("[%.3fms] [rows:%v] %s", elapsedMs, rows, sql))
		}
	}
}
