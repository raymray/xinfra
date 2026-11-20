package xrotate

import (
	"time"
)

type Option interface {
	apply(*option)
}

type optionFunc func(*option)

func (f optionFunc) apply(opt *option) {
	f(opt)
}

type option struct {
	// LinkName 是指向当前日志文件的符号链接的名称
	LinkName string
	// RotationTime 是日志文件轮转的时间间隔
	RotationTime time.Duration
	// MaxAge 是日志文件最大保留时间
	MaxAge time.Duration
	// RotationCount 是最大保留的日志文件数量
	RotationCount uint
	// RotationSize 是日志文件轮转的大小阈值
	RotationSize int64
	// Compress 是否压缩备份的日志文件
	Compress bool
	// LocalTime 是否使用本地时间
	LocalTime bool
}

func WithLinkName(linkName string) Option {
	return optionFunc(func(opt *option) {
		opt.LinkName = linkName
	})
}

func WithRotationTime(d time.Duration) Option {
	return optionFunc(func(opt *option) {
		opt.RotationTime = d
	})
}

func WithMaxAge(d time.Duration) Option {
	return optionFunc(func(opt *option) {
		opt.MaxAge = d
	})
}

func WithRotationCount(n uint) Option {
	return optionFunc(func(opt *option) {
		opt.RotationCount = n
	})
}

func WithRotationSize(size int64) Option {
	return optionFunc(func(opt *option) {
		opt.RotationSize = size
	})
}

func WithCompress(compress bool) Option {
	return optionFunc(func(opt *option) {
		opt.Compress = compress
	})
}

func WithLocalTime(localTime bool) Option {
	return optionFunc(func(opt *option) {
		opt.LocalTime = localTime
	})
}

func loadOptions(opts ...Option) *option {
	options := &option{
		RotationTime:  24 * time.Hour,      // 24 hours
		MaxAge:        30 * 24 * time.Hour, // 30 days
		RotationCount: 0,
		RotationSize:  100 * 1024 * 1024, // 100 MB
		Compress:      false,
		LocalTime:     true,
	}
	for _, opt := range opts {
		opt.apply(options)
	}
	return options
}
