package xdb

import (
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/raymray/xinfra/xerror"
)

const (
	DriverMySQL    = "mysql"
	DriverPostgres = "postgres"
)

type Config struct {
	Driver        string `yaml:"driver" mapstructure:"driver"`                 // mysql | postgres, {default: mysql}
	DSN           string `yaml:"dsn" mapstructure:"dsn"`                       // 必填
	DialTimeout   int    `yaml:"dial_timeout" mapstructure:"dial_timeout"`     // 单位: second, {default: 5}
	ReadTimeout   int    `yaml:"read_timeout" mapstructure:"read_timeout"`     // 单位: second, {default: 5}
	WriteTimeout  int    `yaml:"write_timeout" mapstructure:"write_timeout"`   // 单位: second, {default: 3}
	MaxLifeTime   int    `yaml:"max_life_time" mapstructure:"max_life_time"`   // 单位: second, {default: 30}
	MaxIdleTime   int    `yaml:"max_idle_time" mapstructure:"max_idle_time"`   // 单位: second, {default: 10}
	MaxOpenConns  int    `yaml:"max_open_conns" mapstructure:"max_open_conns"` // {default: 256}
	MaxIdleConns  int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns"` // {default: 10}
	SlowThreshold int    `yaml:"slow_threshold" mapstructure:"slow_threshold"` // 单位: ms, {default: 200}
	Debug         bool   `yaml:"debug" mapstructure:"debug"`                   // 打印全部 SQL, {default: false}
}

// ApplyDefaults 填充默认值并校验 driver 与 DSN；可重复调用，结果不变。
func ApplyDefaults(c *Config) error {
	if c.Driver == "" {
		c.Driver = DriverMySQL
	}
	if c.DialTimeout <= 0 {
		c.DialTimeout = 5
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 5
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 3
	}
	if c.MaxLifeTime <= 0 {
		c.MaxLifeTime = 30
	}
	if c.MaxIdleTime <= 0 {
		c.MaxIdleTime = 10
	}
	if c.MaxOpenConns <= 0 {
		c.MaxOpenConns = 256
	}
	if c.MaxIdleConns <= 0 {
		c.MaxIdleConns = 10
	}
	if c.SlowThreshold <= 0 {
		c.SlowThreshold = 200
	}
	if !hasDialector(c.Driver) {
		return xerror.Newf("unsupported db driver: %s", c.Driver)
	}
	if c.DSN == "" {
		return xerror.Newf("db dsn is empty")
	}
	return normalizeDSN(c)
}

// normalizeDSN 为 mysql DSN 补上未显式设置的超时参数。
func normalizeDSN(c *Config) error {
	if c.Driver != DriverMySQL {
		return nil
	}
	dsn, err := mysql.ParseDSN(c.DSN)
	if err != nil {
		return xerror.Wrapf(err, "parse mysql dsn")
	}
	if dsn.Timeout <= 0 {
		dsn.Timeout = time.Duration(c.DialTimeout) * time.Second
	}
	if dsn.ReadTimeout <= 0 {
		dsn.ReadTimeout = time.Duration(c.ReadTimeout) * time.Second
	}
	if dsn.WriteTimeout <= 0 {
		dsn.WriteTimeout = time.Duration(c.WriteTimeout) * time.Second
	}
	c.DSN = dsn.FormatDSN()
	return nil
}
