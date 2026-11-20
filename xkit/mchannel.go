package xkit

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// MChannel 是一个缓冲区通道，支持安全关闭。
// 写端持读锁、End 持写锁关闭底层 channel，保证不会出现 send 与 close 并发（无数据竞争）。
type MChannel[T any] struct {
	c         chan T
	done      chan struct{} // End 时关闭，唤醒阻塞中的写端
	mu        sync.RWMutex
	ended     atomic.Bool
	closeOnce sync.Once
}

// NewDefaultMChannel 创建一个通道，其缓冲区容量为50000。
func NewDefaultMChannel[T any]() *MChannel[T] {
	return NewMChannel[T](50000)
}

// NewMChannel 创建一个通道，其缓冲区容量为给定值。
func NewMChannel[T any](capacity int) *MChannel[T] {
	return &MChannel[T]{
		c:    make(chan T, capacity),
		done: make(chan struct{}),
	}
}

// Write 发送单个项；通道已结束（含阻塞期间结束）时返回 false。
func (ch *MChannel[T]) Write(row T) bool {
	return ch.WriteCtx(context.Background(), row)
}

// WriteCtx 发送单个项。支持上下文取消。
func (ch *MChannel[T]) WriteCtx(ctx context.Context, row T) bool {
	ch.mu.RLock()
	defer ch.mu.RUnlock()
	if ch.ended.Load() {
		return false
	}
	select {
	case ch.c <- row:
		return true
	case <-ch.done:
		return false
	case <-ctx.Done():
		return false
	}
}

// WriteAll 发送多个项。返回成功发送的项数。
func (ch *MChannel[T]) WriteAll(rows []T) (n int) {
	for i := range rows {
		if !ch.Write(rows[i]) {
			return n
		}
		n++
	}
	return n
}

// Read 从通道读取单个项。
func (ch *MChannel[T]) Read() (row T, done bool) {
	row, ok := <-ch.c
	return row, !ok
}

// ReadCtx 从通道读取单个项。支持上下文取消。
func (ch *MChannel[T]) ReadCtx(ctx context.Context) (row T, done bool) {
	select {
	case <-ctx.Done():
		return row, false
	case row, ok := <-ch.c:
		return row, !ok
	}
}

// End 标记通道为已结束并关闭它。并发安全；会等待阻塞中的写端退出后再关闭底层 channel。
func (ch *MChannel[T]) End() {
	ch.closeOnce.Do(func() {
		ch.ended.Store(true)
		close(ch.done)
		ch.mu.Lock()
		close(ch.c)
		ch.mu.Unlock()
	})
}

// IsEnded 报告通道是否已结束。
func (ch *MChannel[T]) IsEnded() bool {
	return ch.ended.Load()
}

// MultiRead 在2s内收集最多项。返回收集项以及是否已完成。
func (ch *MChannel[T]) MultiRead(count int) ([]T, bool) {
	return ch.MultiReadCtx(context.Background(), count, 2*time.Second)
}

// MultiReadCtx 在给定的持续时间内收集最多项。返回收集项以及是否已完成。
func (ch *MChannel[T]) MultiReadCtx(ctx context.Context, count int, duration time.Duration) ([]T, bool) {
	timeout := time.After(duration)
	var rows []T

	for i := 0; i < count; i++ {
		select {
		case <-ctx.Done():
			return rows, ch.IsEnded()
		case row, ok := <-ch.c:
			if !ok {
				return rows, true
			}
			rows = append(rows, row)
		case <-timeout:
			return rows, false
		}
	}
	return rows, false
}
