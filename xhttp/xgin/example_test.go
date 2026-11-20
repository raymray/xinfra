package xgin_test

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xerror"
	"github.com/raymray/xinfra/xhttp"
	"github.com/raymray/xinfra/xhttp/xgin"
	"go.uber.org/zap"
)

var CodeAssetNotFound = xerror.NewCode(40401, http.StatusNotFound)

// 推荐的中间件顺序：Recovery → Trace → AccessLog → UseMapper。
func Example() {
	logger := zap.NewNop()

	mapper := xhttp.NewMapper()
	mapper.Register(CodeAssetNotFound, "资产不存在")

	r := gin.New()
	r.Use(
		xgin.Recovery(logger),
		xgin.Trace(logger),
		xgin.AccessLog(logger, xgin.WithSkipPaths("/healthz")),
		xgin.UseMapper(mapper),
	)
	r.GET("/assets/:id", func(c *gin.Context) {
		if c.Param("id") == "0" {
			xgin.Fail(c, xerror.NewCodeErr(CodeAssetNotFound))
			return
		}
		xgin.OK(c, gin.H{"id": c.Param("id")})
	})
	_ = r
	// Output:
}
