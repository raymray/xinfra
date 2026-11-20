package xlog

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"
)

func TestApplyDefaults(t *testing.T) {
	c := Config{}
	ApplyDefaults(&c)
	if c.Level != LevelInfo || c.Policy != PolicyStdout {
		t.Fatalf("level/policy = %q/%q", c.Level, c.Policy)
	}
	if c.RotationTime != 1 || c.RotationMaxAge != 7 || c.RotationCount != 10 || c.RotationSize != 200 {
		t.Fatalf("rotation defaults wrong: %+v", c)
	}
}

func TestApplyDefaultsPreservesSet(t *testing.T) {
	c := Config{Level: LevelDebug, RotationCount: 3}
	ApplyDefaults(&c)
	if c.Level != LevelDebug || c.RotationCount != 3 {
		t.Fatalf("set values overwritten: %+v", c)
	}
}

func TestFromContextWithoutLoggerReturnsNop(t *testing.T) {
	l := FromContext(context.Background())
	if l == nil {
		t.Fatal("FromContext returned nil")
	}
	l.Info("must not panic")
}

func TestWithLoggerRoundTrip(t *testing.T) {
	want := zap.NewExample()
	ctx := WithLogger(context.Background(), want)
	if got := FromContext(ctx); got != want {
		t.Fatal("FromContext did not return the logger set by WithLogger")
	}
}

func TestNewFilePolicyWritesFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.log")
	l, err := New(Config{Policy: PolicyFile, Path: path})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	l.Zap().Info("hello")
	_ = l.Zap().Sync()
	if err := l.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	b, err := os.ReadFile(path)
	if err != nil || len(b) == 0 {
		t.Fatalf("log file not written: err=%v len=%d", err, len(b))
	}
}

func TestNewFilePolicyRequiresPath(t *testing.T) {
	if _, err := New(Config{Policy: PolicyFile}); err == nil {
		t.Fatal("expected error when file policy has empty path")
	}
}

func TestCloseIdempotent(t *testing.T) {
	l, err := New(Config{})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("first Close: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Fatalf("second Close: %v", err)
	}
}
