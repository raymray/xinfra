package xredis

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/raymray/xinfra/xerror"
	"github.com/redis/go-redis/v9"
)

const pingTimeout = 5 * time.Second

// Manager 管理多个 Redis 实例，首次 Client(name) 时建连并 Ping。由调用方构造并 Close。
type Manager struct {
	configs map[string]Config
	errs    map[string]error
	dialMu  map[string]*sync.Mutex

	mu      sync.RWMutex
	clients map[string]redis.UniversalClient
	closed  bool
}

func NewManager(configs map[string]Config) *Manager {
	m := &Manager{
		configs: make(map[string]Config, len(configs)),
		errs:    make(map[string]error),
		dialMu:  make(map[string]*sync.Mutex, len(configs)),
		clients: make(map[string]redis.UniversalClient),
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

func (m *Manager) InstanceNames() []string {
	names := make([]string, 0, len(m.configs))
	for name := range m.configs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (m *Manager) Client(name string) (redis.UniversalClient, error) {
	m.mu.RLock()
	c, ok := m.clients[name]
	closed := m.closed
	m.mu.RUnlock()
	if closed {
		return nil, xerror.Newf("redis manager closed")
	}
	if ok {
		return c, nil
	}

	cfg, ok := m.configs[name]
	if !ok {
		return nil, xerror.Newf("redis instance %q not configured (available: %v)", name, m.InstanceNames())
	}
	if err := m.errs[name]; err != nil {
		return nil, xerror.Wrapf(err, "redis instance %q config", name)
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

	nc := newUniversal(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), pingTimeout)
	defer cancel()
	if err := nc.Ping(ctx).Err(); err != nil {
		_ = nc.Close()
		return nil, xerror.Wrapf(err, "redis instance %q ping", name)
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		_ = nc.Close()
		return nil, xerror.Newf("redis manager closed")
	}
	m.clients[name] = nc
	return nc, nil
}

// newUniversal 按启用的模式构造客户端，cfg 须已通过 ApplyDefaults。
func newUniversal(cfg Config) redis.UniversalClient {
	dial := time.Duration(cfg.DialTimeout) * time.Second
	read := time.Duration(cfg.ReadTimeout) * time.Second
	write := time.Duration(cfg.WriteTimeout) * time.Second
	idle := time.Duration(cfg.ConnMaxIdleTime) * time.Minute
	life := time.Duration(cfg.ConnMaxLifetime) * time.Minute

	switch {
	case cfg.Cluster.Enable:
		return redis.NewClusterClient(&redis.ClusterOptions{
			Addrs: cfg.Cluster.Address, Password: cfg.Password,
			DialTimeout: dial, ReadTimeout: read, WriteTimeout: write,
			MinIdleConns: cfg.MinIdleConns, MaxIdleConns: cfg.MaxIdleConns,
			ConnMaxIdleTime: idle, ConnMaxLifetime: life,
		})
	case cfg.Sentinel.Enable:
		return redis.NewFailoverClient(&redis.FailoverOptions{
			MasterName: cfg.Sentinel.MasterName, SentinelAddrs: cfg.Sentinel.Address,
			Password: cfg.Password, DB: cfg.DB,
			DialTimeout: dial, ReadTimeout: read, WriteTimeout: write,
			MinIdleConns: cfg.MinIdleConns, MaxIdleConns: cfg.MaxIdleConns,
			ConnMaxIdleTime: idle, ConnMaxLifetime: life,
		})
	default:
		return redis.NewClient(&redis.Options{
			Addr: cfg.Single.Address, Password: cfg.Password, DB: cfg.DB,
			DialTimeout: dial, ReadTimeout: read, WriteTimeout: write,
			MinIdleConns: cfg.MinIdleConns, MaxIdleConns: cfg.MaxIdleConns,
			ConnMaxIdleTime: idle, ConnMaxLifetime: life,
		})
	}
}

// Close 关闭所有已建立的客户端，可重复调用；关闭后 Client 返回错误。
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
			errs = append(errs, xerror.Wrapf(err, "close redis instance %q", name))
		}
	}
	m.clients = nil
	if len(errs) > 0 {
		return xerror.Joinf("close redis instances", errs)
	}
	return nil
}
