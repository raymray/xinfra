package xlifecycle

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/raymray/xinfra/xerror"
)

// Stage 是一个关闭阶段，阶段内的 Closers 按注册的逆序执行。
type Stage struct {
	Name    string
	Timeout time.Duration
	Closers []func(context.Context) error
}

// Closer 按 Add 顺序分阶段关闭资源，整体受总预算约束。
type Closer struct {
	budget time.Duration

	mu     sync.Mutex
	stages []Stage

	once sync.Once
	err  error
}

func New(totalBudget time.Duration) *Closer {
	return &Closer{budget: totalBudget}
}

func (c *Closer) Add(stage Stage) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.stages = append(c.stages, stage)
}

// Close 执行全部阶段：某阶段超时或出错不阻断后续阶段；总预算耗尽时跳过剩余阶段。
// 可重复调用，之后的调用直接返回第一次的结果。
func (c *Closer) Close(ctx context.Context) error {
	c.once.Do(func() { c.err = c.run(ctx) })
	return c.err
}

func (c *Closer) run(ctx context.Context) error {
	c.mu.Lock()
	stages := append([]Stage(nil), c.stages...)
	c.mu.Unlock()

	total, cancel := withOptionalTimeout(ctx, c.budget)
	defer cancel()

	var errs []error
	for i, s := range stages {
		if total.Err() != nil {
			skipped := make([]string, 0, len(stages)-i)
			for _, rest := range stages[i:] {
				skipped = append(skipped, rest.Name)
			}
			errs = append(errs, xerror.Newf("shutdown budget %v exhausted, skipped stages: %v", c.budget, skipped))
			break
		}
		errs = append(errs, runStage(total, s)...)
	}
	if len(errs) > 0 {
		return xerror.Joinf("shutdown", errs)
	}
	return nil
}

// lateGrace 是阶段超时后，剩余 closer 各自最多等待的时间。尊重 ctx 的 closer 会立即返回，
// 不依赖 ctx 的快速 closer（如关闭连接池）仍能完成。
const lateGrace = 100 * time.Millisecond

// withOptionalTimeout 在 d<=0 时不设时限。
func withOptionalTimeout(parent context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(parent)
	}
	return context.WithTimeout(parent, d)
}

// runStage 按逆序逐个执行 closer。每个 closer 在独立 goroutine 中运行，超时即放弃等待；
// 阶段超时后剩余的 closer 仍会被调用（ctx 已取消），未完成的逐个记录。
func runStage(parent context.Context, s Stage) []error {
	ctx, cancel := withOptionalTimeout(parent, s.Timeout)
	defer cancel()

	var errs []error
	n := len(s.Closers)
	for i := n - 1; i >= 0; i-- {
		label := fmt.Sprintf("stage %s closer %d/%d", s.Name, i+1, n)
		done := make(chan error, 1)
		go func(fn func(context.Context) error) { done <- fn(ctx) }(s.Closers[i])

		var wait <-chan time.Time
		if ctx.Err() != nil {
			wait = time.After(lateGrace)
		}
		select {
		case err := <-done:
			if err != nil {
				errs = append(errs, xerror.Wrapf(err, "%s", label))
			}
			continue
		case <-ctx.Done():
		case <-wait:
		}
		// 超时：再给一次宽限，仍未返回则记为未完成
		select {
		case err := <-done:
			if err != nil {
				errs = append(errs, xerror.Wrapf(err, "%s", label))
			}
		case <-time.After(lateGrace):
			errs = append(errs, xerror.Newf("%s not finished: stage timed out after %v", label, s.Timeout))
		}
	}
	return errs
}
