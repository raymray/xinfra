package xdb

import (
	"net"
	"strings"
	"testing"
	"time"

	"go.uber.org/zap"
)

const okDSN = "root:pw@tcp(127.0.0.1:3306)/app"

func TestApplyDefaultsFillsAll(t *testing.T) {
	c := Config{Driver: "mysql", DSN: okDSN}
	if err := ApplyDefaults(&c); err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	want := Config{DialTimeout: 5, ReadTimeout: 5, WriteTimeout: 3, MaxLifeTime: 30, MaxIdleTime: 10, MaxOpenConns: 256, MaxIdleConns: 10, SlowThreshold: 200}
	if c.DialTimeout != want.DialTimeout || c.ReadTimeout != want.ReadTimeout || c.WriteTimeout != want.WriteTimeout ||
		c.MaxLifeTime != want.MaxLifeTime || c.MaxIdleTime != want.MaxIdleTime || c.MaxOpenConns != want.MaxOpenConns ||
		c.MaxIdleConns != want.MaxIdleConns || c.SlowThreshold != want.SlowThreshold {
		t.Fatalf("defaults wrong: %+v", c)
	}
}

func TestApplyDefaultsDriverDefaultsToMySQL(t *testing.T) {
	c := Config{DSN: okDSN}
	if err := ApplyDefaults(&c); err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	if c.Driver != "mysql" {
		t.Fatalf("Driver = %q, want mysql", c.Driver)
	}
}

func TestApplyDefaultsRejectsUnknownDriver(t *testing.T) {
	c := Config{Driver: "dameng", DSN: "x"}
	err := ApplyDefaults(&c)
	if err == nil || !strings.Contains(err.Error(), "unsupported") || !strings.Contains(err.Error(), "dameng") {
		t.Fatalf("want unsupported driver error naming dameng, got %v", err)
	}
}

func TestApplyDefaultsRejectsEmptyDSN(t *testing.T) {
	c := Config{Driver: "mysql"}
	if err := ApplyDefaults(&c); err == nil {
		t.Fatal("want error for empty dsn")
	}
}

func TestApplyDefaultsPreservesSet(t *testing.T) {
	c := Config{Driver: "mysql", DSN: okDSN, DialTimeout: 9, MaxOpenConns: 64}
	if err := ApplyDefaults(&c); err != nil {
		t.Fatalf("ApplyDefaults: %v", err)
	}
	if c.DialTimeout != 9 || c.MaxOpenConns != 64 {
		t.Fatalf("set values overwritten: %+v", c)
	}
}

func TestApplyDefaultsIdempotent(t *testing.T) {
	c := Config{Driver: "mysql", DSN: okDSN}
	_ = ApplyDefaults(&c)
	first := c
	_ = ApplyDefaults(&c)
	if c != first {
		t.Fatalf("second ApplyDefaults changed config:\n%+v\n%+v", first, c)
	}
}

func TestNewManagerAppliesDefaults(t *testing.T) {
	m := NewManager(map[string]Config{"master": {DSN: okDSN}}, zap.NewNop())
	if got := m.configs["master"].MaxOpenConns; got != 256 {
		t.Fatalf("MaxOpenConns after NewManager = %d, want 256", got)
	}
}

func TestInstanceNamesSorted(t *testing.T) {
	m := NewManager(map[string]Config{"slave": {DSN: okDSN}, "master": {DSN: okDSN}}, zap.NewNop())
	got := strings.Join(m.InstanceNames(), ",")
	if got != "master,slave" {
		t.Fatalf("InstanceNames = %s", got)
	}
}

func TestClientUnknownNameListsAvailable(t *testing.T) {
	m := NewManager(map[string]Config{"master": {DSN: okDSN}, "report": {DSN: okDSN}}, zap.NewNop())
	_, err := m.Client("mastr")
	if err == nil {
		t.Fatal("want error for unknown instance")
	}
	for _, s := range []string{"mastr", "master", "report"} {
		if !strings.Contains(err.Error(), s) {
			t.Fatalf("error %q should mention %q", err, s)
		}
	}
}

func TestClientInvalidConfigReturnsError(t *testing.T) {
	m := NewManager(map[string]Config{"bad": {Driver: "dameng", DSN: "x"}}, zap.NewNop())
	_, err := m.Client("bad")
	if err == nil || !strings.Contains(err.Error(), "bad") {
		t.Fatalf("want error naming instance bad, got %v", err)
	}
}

func TestClientUnreachableDSNFails(t *testing.T) {
	m := NewManager(map[string]Config{"master": {DSN: "root:pw@tcp(127.0.0.1:1)/app", DialTimeout: 1}}, zap.NewNop())
	start := time.Now()
	_, err := m.Client("master")
	if err == nil {
		t.Fatal("want error for unreachable dsn")
	}
	if !strings.Contains(err.Error(), "master") {
		t.Fatalf("error should name instance: %v", err)
	}
	if d := time.Since(start); d > 3*time.Second {
		t.Fatalf("took %v, want < 3s", d)
	}
}

func TestCloseIdempotent(t *testing.T) {
	m := NewManager(map[string]Config{"master": {DSN: okDSN}}, zap.NewNop())
	if err := m.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := m.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
	if _, err := m.Client("master"); err == nil {
		t.Fatal("Client after Close should fail")
	}
}

func TestSlowInstanceDoesNotBlockOthers(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer ln.Close()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			defer c.Close()
		}
	}()
	m := NewManager(map[string]Config{
		"slow": {DSN: "root:pw@tcp(" + ln.Addr().String() + ")/app", DialTimeout: 2, ReadTimeout: 2},
		"bad":  {Driver: "dameng", DSN: "x"},
	}, zap.NewNop())
	defer m.Close()
	go m.Client("slow")
	time.Sleep(100 * time.Millisecond)
	start := time.Now()
	_, _ = m.Client("bad") // 不需要建连，应立即返回
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Fatalf("other instance blocked %v behind slow dial", d)
	}
}
