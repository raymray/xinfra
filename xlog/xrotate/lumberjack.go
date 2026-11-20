package xrotate

import (
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// NewLumberjack 按大小轮转。fileName 为空时由调用方提前校验。
func NewLumberjack(fileName string, opts ...Option) *lumberjack.Logger {
	options := loadOptions(opts...)
	return &lumberjack.Logger{
		Filename:   fileName,
		MaxSize:    max(1, int(options.RotationSize/(1024*1024))), // MB
		MaxAge:     max(1, int(options.MaxAge/(24*time.Hour))),    // day
		MaxBackups: int(options.RotationCount),
		LocalTime:  options.LocalTime,
		Compress:   options.Compress,
	}
}
