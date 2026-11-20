package xgin

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xhttp"
	"go.uber.org/zap"
)

type accessLogOptions struct {
	skip   map[string]struct{}
	fields func(*gin.Context) []zap.Field
}

type AccessLogOption func(*accessLogOptions)

func WithSkipPaths(paths ...string) AccessLogOption {
	return func(o *accessLogOptions) {
		for _, p := range paths {
			o.skip[p] = struct{}{}
		}
	}
}

// WithLogFields 为每条访问日志追加字段，如租户、用户。
func WithLogFields(fn func(*gin.Context) []zap.Field) AccessLogOption {
	return func(o *accessLogOptions) { o.fields = fn }
}

// AccessLog 在请求结束后记录一条访问日志，标准字段：
// method、path、status、latency、client_ip、request_id，有错误时附 errors。
func AccessLog(logger *zap.Logger, opts ...AccessLogOption) gin.HandlerFunc {
	o := accessLogOptions{skip: map[string]struct{}{}}
	for _, opt := range opts {
		opt(&o)
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		if _, ok := o.skip[path]; ok {
			c.Next()
			return
		}
		start := time.Now()
		c.Next()

		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", time.Since(start)),
			zap.String("client_ip", c.ClientIP()),
			zap.String("request_id", xhttp.RequestIDFrom(c.Request.Context())),
		}
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("errors", c.Errors.String()))
		}
		if o.fields != nil {
			fields = append(fields, o.fields(c)...)
		}
		logger.Info("access", fields...)
	}
}
