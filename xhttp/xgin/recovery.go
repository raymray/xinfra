package xgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xhttp"
	"go.uber.org/zap"
)

// Recovery 捕获 panic，记录带堆栈的错误日志，返回 500 信封。
// 放在 Trace 之前也能拿到 request id：优先取 ctx，其次请求头，都没有则生成。
func Recovery(logger *zap.Logger) gin.HandlerFunc {
	if logger == nil {
		logger = zap.NewNop()
	}
	return func(c *gin.Context) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// net/http 约定用 ErrAbortHandler 中止连接，交还给 http.Server 处理
			if rec == http.ErrAbortHandler {
				panic(rec)
			}
			id := xhttp.RequestIDFrom(c.Request.Context())
			if id == "" {
				id = c.GetHeader(DefaultRequestIDHeader)
			}
			if id == "" {
				id = newRequestID()
			}
			c.Header(DefaultRequestIDHeader, id)
			logger.Error("panic recovered",
				zap.Any("panic", rec),
				zap.String("request_id", id),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Stack("stack"),
			)
			// 响应已开始写出时无法再改状态码，追加信封只会让响应体损坏
			if c.Writer.Written() {
				c.Abort()
				return
			}
			c.AbortWithStatusJSON(http.StatusInternalServerError,
				xhttp.NewFail(xhttp.CodeInternal, "internal error", id))
		}()
		c.Next()
	}
}
