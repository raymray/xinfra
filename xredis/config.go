package xredis

import "github.com/raymray/xinfra/xerror"

type Config struct {
	Single   Single   `yaml:"single" mapstructure:"single"`
	Sentinel Sentinel `yaml:"sentinel" mapstructure:"sentinel"`
	Cluster  Cluster  `yaml:"cluster" mapstructure:"cluster"`

	DB              int    `yaml:"db" mapstructure:"db"`                                 // {default: 0}，cluster 模式忽略
	Password        string `yaml:"password" mapstructure:"password"`                     // {default: ""}
	DialTimeout     int    `yaml:"dial_timeout" mapstructure:"dial_timeout"`             // 单位: second, {default: 5}
	ReadTimeout     int    `yaml:"read_timeout" mapstructure:"read_timeout"`             // 单位: second, {default: 3}
	WriteTimeout    int    `yaml:"write_timeout" mapstructure:"write_timeout"`           // 单位: second, {default: 3}
	MinIdleConns    int    `yaml:"min_idle_conns" mapstructure:"min_idle_conns"`         // {default: 0}
	MaxIdleConns    int    `yaml:"max_idle_conns" mapstructure:"max_idle_conns"`         // {default: 0}
	ConnMaxIdleTime int    `yaml:"conn_max_idle_time" mapstructure:"conn_max_idle_time"` // 单位: minute, {default: 30}
	ConnMaxLifetime int    `yaml:"conn_max_lifetime" mapstructure:"conn_max_lifetime"`   // 单位: minute, {default: 0 不限制}
}

type Single struct {
	Enable  bool   `yaml:"enable" mapstructure:"enable"`
	Address string `yaml:"address" mapstructure:"address"` // host:port
}

type Sentinel struct {
	Enable     bool     `yaml:"enable" mapstructure:"enable"`
	MasterName string   `yaml:"master_name" mapstructure:"master_name"`
	Address    []string `yaml:"address" mapstructure:"address"` // [host:port]
}

type Cluster struct {
	Enable  bool     `yaml:"enable" mapstructure:"enable"`
	Address []string `yaml:"address" mapstructure:"address"` // [host:port]
}

// ApplyDefaults 填充默认值，并要求 single/sentinel/cluster 恰有一个启用；可重复调用。
func ApplyDefaults(c *Config) error {
	if c.DialTimeout <= 0 {
		c.DialTimeout = 5
	}
	if c.ReadTimeout <= 0 {
		c.ReadTimeout = 3
	}
	if c.WriteTimeout <= 0 {
		c.WriteTimeout = 3
	}
	if c.ConnMaxIdleTime <= 0 {
		c.ConnMaxIdleTime = 30
	}

	enabled := 0
	for _, on := range []bool{c.Single.Enable, c.Sentinel.Enable, c.Cluster.Enable} {
		if on {
			enabled++
		}
	}
	if enabled != 1 {
		return xerror.Newf("redis: exactly one of single/sentinel/cluster must be enabled, got %d", enabled)
	}
	switch {
	case c.Single.Enable && c.Single.Address == "":
		return xerror.Newf("redis single: address is empty")
	case c.Sentinel.Enable && (c.Sentinel.MasterName == "" || len(c.Sentinel.Address) == 0):
		return xerror.Newf("redis sentinel: master_name and address are required")
	case c.Cluster.Enable && len(c.Cluster.Address) == 0:
		return xerror.Newf("redis cluster: address is empty")
	}
	return nil
}
