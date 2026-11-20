package xlifecycle

import (
	"context"
	"errors"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type recorder struct {
	mu    sync.Mutex
	calls []string
}

func (r *recorder) closer(name string) func(context.Context) error {
	return func(context.Context) error {
		r.mu.Lock()
		defer r.mu.Unlock()
		r.calls = append(r.calls, name)
		return nil
	}
}

func (r *recorder) got() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return strings.Join(r.calls, ",")
}

func TestStagesRunInAddOrder(t *testing.T) {
	r := &recorder{}
	c := New(time.Second)
	c.Add(Stage{Name: "http", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("http")}})
	c.Add(Stage{Name: "stores", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("stores")}})
	if err := c.Close(context.Background()); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if got := r.got(); got != "http,stores" {
		t.Fatalf("order = %s", got)
	}
}

func TestClosersRunInReverseWithinStage(t *testing.T) {
	r := &recorder{}
	c := New(time.Second)
	c.Add(Stage{Name: "stores", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("db"), r.closer("redis")}})
	_ = c.Close(context.Background())
	if got := r.got(); got != "redis,db" {
		t.Fatalf("order = %s, want redis,db", got)
	}
}

func TestStageTimeoutContinuesToNextStage(t *testing.T) {
	r := &recorder{}
	block := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	c := New(time.Second)
	c.Add(Stage{Name: "http", Timeout: 50 * time.Millisecond, Closers: []func(context.Context) error{block}})
	c.Add(Stage{Name: "stores", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("stores")}})
	err := c.Close(context.Background())
	if r.got() != "stores" {
		t.Fatalf("next stage not run, calls=%q", r.got())
	}
	if err == nil || !strings.Contains(err.Error(), "http") {
		t.Fatalf("want error naming timed-out stage http, got %v", err)
	}
}

func TestStageTimeoutWhenCloserIgnoresContext(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	stuck := func(context.Context) error { <-release; return nil }
	r := &recorder{}
	c := New(time.Second)
	c.Add(Stage{Name: "http", Timeout: 50 * time.Millisecond, Closers: []func(context.Context) error{stuck}})
	c.Add(Stage{Name: "stores", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("stores")}})
	start := time.Now()
	_ = c.Close(context.Background())
	if d := time.Since(start); d > 500*time.Millisecond {
		t.Fatalf("Close blocked %v on a closer ignoring ctx", d)
	}
	if r.got() != "stores" {
		t.Fatalf("next stage not run, calls=%q", r.got())
	}
}

func TestCloserErrorDoesNotBlockOthers(t *testing.T) {
	r := &recorder{}
	fail := func(msg string) func(context.Context) error {
		return func(context.Context) error { return errors.New(msg) }
	}
	c := New(time.Second)
	c.Add(Stage{Name: "stores", Timeout: time.Second, Closers: []func(context.Context) error{fail("db boom"), r.closer("mid"), fail("redis boom")}})
	err := c.Close(context.Background())
	if r.got() != "mid" {
		t.Fatalf("middle closer not run, calls=%q", r.got())
	}
	if err == nil || !strings.Contains(err.Error(), "db boom") || !strings.Contains(err.Error(), "redis boom") {
		t.Fatalf("want both errors, got %v", err)
	}
}

func TestTotalBudgetExhaustedReturnsError(t *testing.T) {
	r := &recorder{}
	slow := func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
	c := New(60 * time.Millisecond)
	c.Add(Stage{Name: "http", Timeout: time.Second, Closers: []func(context.Context) error{slow}})
	c.Add(Stage{Name: "stores", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("stores")}})
	err := c.Close(context.Background())
	if err == nil || !strings.Contains(err.Error(), "budget") || !strings.Contains(err.Error(), "stores") {
		t.Fatalf("want budget error naming skipped stage stores, got %v", err)
	}
	if r.got() != "" {
		t.Fatalf("stage after budget exhausted should be skipped, calls=%q", r.got())
	}
}

func TestCloseIdempotent(t *testing.T) {
	var n atomic.Int32
	c := New(time.Second)
	c.Add(Stage{Name: "s", Timeout: time.Second, Closers: []func(context.Context) error{func(context.Context) error {
		n.Add(1)
		return errors.New("once")
	}}})
	first := c.Close(context.Background())
	second := c.Close(context.Background())
	if n.Load() != 1 {
		t.Fatalf("closer ran %d times", n.Load())
	}
	if first == nil || second == nil || first.Error() != second.Error() {
		t.Fatalf("second Close should return first result: %v / %v", first, second)
	}
}

func TestWaitSignalInvokesFnOnce(t *testing.T) {
	ch := make(chan os.Signal, 1)
	var n atomic.Int32
	done := make(chan struct{})
	go func() { waitOn(context.Background(), ch, func() { n.Add(1) }); close(done) }()
	ch <- os.Interrupt
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("waitOn did not return after signal")
	}
	if n.Load() != 1 {
		t.Fatalf("fn called %d times", n.Load())
	}
}

func TestWaitSignalReturnsOnContextCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	var n atomic.Int32
	done := make(chan struct{})
	go func() { waitOn(ctx, make(chan os.Signal, 1), func() { n.Add(1) }); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("waitOn did not return after ctx cancel")
	}
	if n.Load() != 0 {
		t.Fatal("fn must not run on ctx cancel")
	}
}

func TestHungCloserDoesNotSkipRemainingClosersInStage(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	hung := func(context.Context) error { <-release; return nil }
	r := &recorder{}
	c := New(time.Second)
	// 逆序执行：hung 先跑，db 必须仍被调用
	c.Add(Stage{Name: "stores", Timeout: 50 * time.Millisecond, Closers: []func(context.Context) error{r.closer("db"), hung}})
	err := c.Close(context.Background())
	if r.got() != "db" {
		t.Fatalf("remaining closer not run, calls=%q", r.got())
	}
	if err == nil || !strings.Contains(err.Error(), "closer 2/2") {
		t.Fatalf("want error naming unfinished closer 2/2, got %v", err)
	}
}

func TestZeroStageTimeoutMeansNoStageLimit(t *testing.T) {
	r := &recorder{}
	c := New(time.Second)
	c.Add(Stage{Name: "s", Closers: []func(context.Context) error{func(ctx context.Context) error {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return r.closer("ok")(ctx)
	}}})
	if err := c.Close(context.Background()); err != nil || r.got() != "ok" {
		t.Fatalf("err=%v calls=%q", err, r.got())
	}
}

func TestZeroBudgetRunsAllStages(t *testing.T) {
	r := &recorder{}
	c := New(0)
	c.Add(Stage{Name: "a", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("a")}})
	c.Add(Stage{Name: "b", Timeout: time.Second, Closers: []func(context.Context) error{r.closer("b")}})
	if err := c.Close(context.Background()); err != nil || r.got() != "a,b" {
		t.Fatalf("err=%v calls=%q", err, r.got())
	}
}
