package xlog

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

var nop = zap.NewNop()

// WithLogger 把 logger 放入 ctx，通常由 trace 中间件写入带 request_id 的 logger。
func WithLogger(ctx context.Context, l *zap.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// FromContext 取出 ctx 中的 logger；没有时返回不输出的 logger，不会 panic。
func FromContext(ctx context.Context) *zap.Logger {
	if ctx != nil {
		if l, ok := ctx.Value(ctxKey{}).(*zap.Logger); ok && l != nil {
			return l
		}
	}
	return nop
}
