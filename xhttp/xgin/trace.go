package xgin

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xhttp"
	"github.com/raymray/xinfra/xlog"
	"go.uber.org/zap"
)

const DefaultRequestIDHeader = "X-Request-ID"

type traceOptions struct {
	header string
	gen    func() string
}

type TraceOption func(*traceOptions)

func WithRequestIDHeader(h string) TraceOption {
	return func(o *traceOptions) { o.header = h }
}

func WithIDGenerator(fn func() string) TraceOption {
	return func(o *traceOptions) { o.gen = fn }
}

// Trace 透传或生成 request id，写入响应头与请求 ctx，
// 并把带 request_id 字段的 base logger 放入 ctx（xlog.FromContext 可取）。
func Trace(base *zap.Logger, opts ...TraceOption) gin.HandlerFunc {
	o := traceOptions{header: DefaultRequestIDHeader, gen: newRequestID}
	for _, opt := range opts {
		opt(&o)
	}
	if base == nil {
		base = zap.NewNop()
	}
	return func(c *gin.Context) {
		id := c.GetHeader(o.header)
		if id == "" {
			id = o.gen()
		}
		c.Header(o.header, id)
		ctx := xhttp.WithRequestID(c.Request.Context(), id)
		ctx = xlog.WithLogger(ctx, base.With(zap.String("request_id", id)))
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}

func newRequestID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}
