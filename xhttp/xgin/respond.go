// Package xgin 是 xhttp 契约的 gin 适配：响应函数与通用中间件。
package xgin

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xhttp"
)

const mapperKey = "xhttp.mapper"

var defaultMapper = xhttp.NewMapper()

// UseMapper 把服务的错误码表放入请求上下文，供 Fail 使用。
func UseMapper(m *xhttp.Mapper) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(mapperKey, m)
		c.Next()
	}
}

func mapperOf(c *gin.Context) *xhttp.Mapper {
	if v, ok := c.Get(mapperKey); ok {
		if m, ok := v.(*xhttp.Mapper); ok && m != nil {
			return m
		}
	}
	return defaultMapper
}

// OK 写成功信封，HTTP 200。
func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, xhttp.NewOK(data, xhttp.RequestIDFrom(c.Request.Context())))
}

// Fail 按错误码表写失败信封并中止后续 handler；原始错误文本不进入响应。
func Fail(c *gin.Context, err error) {
	status, code, msg := mapperOf(c).Resolve(err)
	if err != nil {
		_ = c.Error(err) // 留给访问日志记录原始错误
	}
	c.AbortWithStatusJSON(status, xhttp.NewFail(code, msg, xhttp.RequestIDFrom(c.Request.Context())))
}
