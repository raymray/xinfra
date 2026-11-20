package xlifecycle

import (
	"context"
	"os"
	"os/signal"
)

// WaitSignal 阻塞直到收到退出信号（调用一次 fn）或 ctx 取消（不调用 fn）。
func WaitSignal(ctx context.Context, fn func()) {
	ch := make(chan os.Signal, 1) // 带缓冲，避免 handler 就绪前到达的信号被丢弃
	signal.Notify(ch, exitSignals...)
	defer signal.Stop(ch)
	waitOn(ctx, ch, fn)
}

func waitOn(ctx context.Context, ch <-chan os.Signal, fn func()) {
	select {
	case <-ctx.Done():
	case <-ch:
		fn()
	}
}
