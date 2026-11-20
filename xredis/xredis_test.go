package xredis

import (
	"context"
	"net"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestApplyDefaults(t *testing.T) {
	c := Config{Single: Single{Enable: true, Address: "127.0.0.1:6379"}}
	if err := ApplyDefaults(&c); err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	if c.DialTimeout != 5 || c.ReadTimeout != 3 || c.WriteTimeout != 3 || c.ConnMaxIdleTime != 30 {
		t.Fatalf("defaults wrong: %+v", c)
	}
}

func TestApplyDefaultsRequiresExactlyOneMode(t *testing.T) {
	cases := map[string]Config{
		"none": {},
		"two": {
			Single:   Single{Enable: true, Address: "a:1"},
			Sentinel: Sentinel{Enable: true, MasterName: "m", Address: []string{"b:1"}},
		},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			err := ApplyDefaults(&c)
			if err == nil {
				t.Fatal("want error")
			}
			for _, s := range []string{"single", "sentinel", "cluster"} {
				if !strings.Contains(err.Error(), s) {
					t.Fatalf("error %q should mention %s", err, s)
				}
			}
		})
	}
}

func TestApplyDefaultsRequiresAddress(t *testing.T) {
	cases := map[string]Config{
		"single":   {Single: Single{Enable: true}},
		"sentinel": {Sentinel: Sentinel{Enable: true, MasterName: "m"}},
		"cluster":  {Cluster: Cluster{Enable: true}},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			if err := ApplyDefaults(&c); err == nil {
				t.Fatal("want error for missing address")
			}
		})
	}
}

func TestSingleModeWithMiniredis(t *testing.T) {
	s := miniredis.RunT(t)
	m := NewManager(map[string]Config{"default": {Single: Single{Enable: true, Address: s.Addr()}}})
	defer m.Close()
	c, err := m.Client("default")
	if err != nil {
		t.Fatalf("Client: %v", err)
	}
	ctx := context.Background()
	if err := c.Set(ctx, "k", "v", 0).Err(); err != nil {
		t.Fatalf("Set: %v", err)
	}
	if got, _ := c.Get(ctx, "k").Result(); got != "v" {
		t.Fatalf("Get = %q", got)
	}
	again, _ := m.Client("default")
	if again != c {
		t.Fatal("Client should return cached instance")
	}
}

func TestClusterModeUsesClusterClient(t *testing.T) {
	cfg := Config{Cluster: Cluster{Enable: true, Address: []string{"127.0.0.1:1"}}}
	if err := ApplyDefaults(&cfg); err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	c := newUniversal(cfg)
	defer c.Close()
	if _, ok := c.(*redis.ClusterClient); !ok {
		t.Fatalf("got %T, want *redis.ClusterClient", c)
	}
}

func TestClusterUnreachableReturnsError(t *testing.T) {
	m := NewManager(map[string]Config{"c": {Cluster: Cluster{Enable: true, Address: []string{"127.0.0.1:1"}}, DialTimeout: 1}})
	defer m.Close()
	if _, err := m.Client("c"); err == nil || !strings.Contains(err.Error(), "\"c\"") {
		t.Fatalf("want error naming instance c, got %v", err)
	}
}

func TestUnknownInstanceListsAvailable(t *testing.T) {
	m := NewManager(map[string]Config{"default": {}, "session": {}})
	_, err := m.Client("sesion")
	if err == nil {
		t.Fatal("want error")
	}
	for _, s := range []string{"sesion", "default", "session"} {
		if !strings.Contains(err.Error(), s) {
			t.Fatalf("error %q should mention %q", err, s)
		}
	}
}

func TestCloseIdempotent(t *testing.T) {
	s := miniredis.RunT(t)
	m := NewManager(map[string]Config{"default": {Single: Single{Enable: true, Address: s.Addr()}}})
	if _, err := m.Client("default"); err != nil {
		t.Fatalf("Client: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := m.Client("default"); err == nil {
		t.Fatal("Client after Close should fail")
	}
}

// blackhole 接受连接但从不响应，模拟卡住的实例。
func blackhole(t *testing.T) string {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			t.Cleanup(func() { c.Close() })
		}
	}()
	return ln.Addr().String()
}

func TestSlowInstanceDoesNotBlockOthers(t *testing.T) {
	s := miniredis.RunT(t)
	m := NewManager(map[string]Config{
		"slow": {Single: Single{Enable: true, Address: blackhole(t)}, ReadTimeout: 2, DialTimeout: 2},
		"fast": {Single: Single{Enable: true, Address: s.Addr()}},
	})
	defer m.Close()
	go m.Client("slow")
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	if _, err := m.Client("fast"); err != nil {
		t.Fatalf("fast: %v", err)
	}
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Fatalf("fast instance blocked %v behind slow dial", d)
	}
}
