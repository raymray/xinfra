package xlog

import "go.uber.org/zap/zapcore"

type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

type Policy string

const (
	PolicyStdout Policy = "stdout"
	PolicyFile   Policy = "file"
)

type Config struct {
	Level          Level  `yaml:"level" mapstructure:"level"`                       // error/warn/info/debug, {default: info}
	Policy         Policy `yaml:"policy" mapstructure:"policy"`                     // file|stdout, {default: stdout}
	Path           string `yaml:"path" mapstructure:"path"`                         // policy=file 时必填
	RotationTime   uint   `yaml:"rotation_time" mapstructure:"rotation_time"`       // 单位: day, {default: 1}
	RotationMaxAge uint   `yaml:"rotation_max_age" mapstructure:"rotation_max_age"` // 单位: day, {default: 7}
	RotationCount  uint   `yaml:"rotation_count" mapstructure:"rotation_count"`     // {default: 10}
	RotationSize   uint   `yaml:"rotation_size" mapstructure:"rotation_size"`       // 单位: MB, {default: 200}
	Compress       bool   `yaml:"compress" mapstructure:"compress"`                 // {default: false}
}

// ApplyDefaults 填充未设置的字段，已设置的值保持不变。
func ApplyDefaults(c *Config) {
	if c.Level == "" {
		c.Level = LevelInfo
	}
	if c.Policy == "" {
		c.Policy = PolicyStdout
	}
	if c.RotationTime == 0 {
		c.RotationTime = 1
	}
	if c.RotationMaxAge == 0 {
		c.RotationMaxAge = 7
	}
	if c.RotationCount == 0 {
		c.RotationCount = 10
	}
	if c.RotationSize == 0 {
		c.RotationSize = 200
	}
}

func parseLevel(l Level) zapcore.Level {
	switch l {
	case LevelDebug:
		return zapcore.DebugLevel
	case LevelWarn:
		return zapcore.WarnLevel
	case LevelError:
		return zapcore.ErrorLevel
	default:
		return zapcore.InfoLevel
	}
}
