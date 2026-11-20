package xdb

import (
	"sort"
	"sync"
	"time"

	"github.com/raymray/xinfra/xerror"
	"go.uber.org/zap"
	gormlogger "gorm.io/gorm/logger"
)

// Manager 管理多个数据库实例，首次 Client(name) 时建连。由调用方构造并 Close。
type Manager struct {
	configs map[string]Config
	errs    map[string]error // ApplyDefaults 失败的实例，在 Client 时返回
	dialMu  map[string]*sync.Mutex
	logger  *zap.Logger

	mu      sync.RWMutex
	clients map[string]Client
	closed  bool
}

// NewManager 对每个实例填充默认值；配置错误延迟到 Client(name) 返回。
func NewManager(configs map[string]Config, logger *zap.Logger) *Manager {
	if logger == nil {
		logger = zap.NewNop()
	}
	m := &Manager{
		configs: make(map[string]Config, len(configs)),
		errs:    make(map[string]error),
		dialMu:  make(map[string]*sync.Mutex, len(configs)),
		logger:  logger,
		clients: make(map[string]Client),
	}
	for name, cfg := range configs {
		m.dialMu[name] = &sync.Mutex{}
		if err := ApplyDefaults(&cfg); err != nil {
			m.errs[name] = err
		}
		m.configs[name] = cfg
	}
	return m
}

// InstanceNames 返回已配置的实例名（已排序），装配层据此逐个预连接。
func (m *Manager) InstanceNames() []string {
	names := make([]string, 0, len(m.configs))
	for name := range m.configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (m *Manager) Client(name string) (Client, error) {
	m.mu.RLock()
	c, ok := m.clients[name]
	closed := m.closed
	m.mu.RUnlock()
	if closed {
		return nil, xerror.Newf("db manager closed")
	}
	if ok {
		return c, nil
	}

	cfg, ok := m.configs[name]
	if !ok {
		return nil, xerror.Newf("db instance %q not configured (available: %v)", name, m.InstanceNames())
	}
	if err := m.errs[name]; err != nil {
		return nil, xerror.Wrapf(err, "db instance %q config", name)
	}

	// 每个实例独立的拨号锁：慢实例只阻塞同名调用，不阻塞其他实例
	dl := m.dialMu[name]
	dl.Lock()
	defer dl.Unlock()
	m.mu.RLock()
	c, ok = m.clients[name]
	m.mu.RUnlock()
	if ok {
		return c, nil
	}

	nc, err := newClient(cfg, m.gormLogger(cfg))
	if err != nil {
		return nil, xerror.Wrapf(err, "db instance %q", name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		_ = nc.Close()
		return nil, xerror.Newf("db manager closed")
	}
	m.clients[name] = nc
	return nc, nil
}

func (m *Manager) gormLogger(cfg Config) gormlogger.Interface {
	level := gormlogger.Warn
	if cfg.Debug {
		level = gormlogger.Info
	}
	return NewGormLogger(m.logger, gormlogger.Config{
		SlowThreshold:             time.Duration(cfg.SlowThreshold) * time.Millisecond,
		IgnoreRecordNotFoundError: true,
		LogLevel:                  level,
	})
}

// Close 关闭所有已建立的连接，可重复调用；关闭后 Client 返回错误。
func (m *Manager) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return nil
	}
	m.closed = true
	var errs []error
	for name, c := range m.clients {
		if err := c.Close(); err != nil {
			errs = append(errs, xerror.Wrapf(err, "close db instance %q", name))
		}
	}
	m.clients = nil
	if len(errs) > 0 {
		return xerror.Joinf("close db instances", errs)
	}
	return nil
}
