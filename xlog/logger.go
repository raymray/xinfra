package xlog

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/raymray/xinfra/xerror"
	"github.com/raymray/xinfra/xlog/xrotate"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger 持有 zap.Logger 及其底层输出，由调用方构造并负责 Close。
type Logger struct {
	log       *zap.Logger
	closer    io.Closer
	closeOnce sync.Once
	closeErr  error
}

func New(cfg Config) (*Logger, error) {
	ApplyDefaults(&cfg)

	var (
		out     io.Writer = os.Stdout
		closer  io.Closer
		encoder = zapcore.NewConsoleEncoder
	)
	if cfg.Policy == PolicyFile {
		if cfg.Path == "" {
			return nil, xerror.Newf("log policy is file but path is empty")
		}
		lj := xrotate.NewLumberjack(cfg.Path,
			xrotate.WithRotationTime(time.Duration(cfg.RotationTime)*24*time.Hour),
			xrotate.WithMaxAge(time.Duration(cfg.RotationMaxAge)*24*time.Hour),
			xrotate.WithRotationCount(cfg.RotationCount),
			xrotate.WithRotationSize(int64(cfg.RotationSize)*1024*1024),
			xrotate.WithCompress(cfg.Compress),
		)
		out, closer, encoder = lj, lj, zapcore.NewJSONEncoder
	}

	encCfg := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	core := zapcore.NewCore(encoder(encCfg), zapcore.AddSync(out), zap.NewAtomicLevelAt(parseLevel(cfg.Level)))
	return &Logger{log: zap.New(core, zap.AddCaller()), closer: closer}, nil
}

func (l *Logger) Zap() *zap.Logger { return l.log }

// Close 刷新并释放底层输出，可重复调用。
func (l *Logger) Close() error {
	l.closeOnce.Do(func() {
		_ = l.log.Sync()
		if l.closer != nil {
			l.closeErr = l.closer.Close()
		}
	})
	return l.closeErr
}
