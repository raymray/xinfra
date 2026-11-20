package xgin

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xerror"
	"github.com/raymray/xinfra/xhttp"
)

func init() { gin.SetMode(gin.TestMode) }

func serve(t *testing.T, r *gin.Engine, path string) (*httptest.ResponseRecorder, map[string]any) {
	t.Helper()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("body not json: %s", w.Body.String())
	}
	return w, body
}

func withRequestID(id string) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = c.Request.WithContext(xhttp.WithRequestID(c.Request.Context(), id))
		c.Next()
	}
}

func TestOKWritesEnvelope(t *testing.T) {
	r := gin.New()
	r.GET("/", func(c *gin.Context) { OK(c, map[string]int{"id": 7}) })
	w, body := serve(t, r, "/")
	if w.Code != 200 || body["code"].(float64) != 0 || body["message"] != "ok" {
		t.Fatalf("got %d %v", w.Code, body)
	}
	if _, ok := body["request_id"]; !ok {
		t.Fatal("request_id key missing")
	}
	if body["data"].(map[string]any)["id"].(float64) != 7 {
		t.Fatalf("data wrong: %v", body["data"])
	}
}

func TestFailMappedCode(t *testing.T) {
	code := xerror.NewCode(40001, http.StatusNotFound)
	m := xhttp.NewMapper()
	m.Register(code, "资源不存在")
	r := gin.New()
	r.Use(UseMapper(m))
	r.GET("/", func(c *gin.Context) { Fail(c, xerror.NewCodeErr(code, "row missing")) })
	w, body := serve(t, r, "/")
	if w.Code != 404 || body["code"].(float64) != 40001 || body["message"] != "资源不存在" {
		t.Fatalf("got %d %v", w.Code, body)
	}
	if _, ok := body["data"]; ok {
		t.Fatal("failure must not carry data")
	}
}

func TestFailPlainErrorIs500NoLeak(t *testing.T) {
	r := gin.New()
	r.GET("/", func(c *gin.Context) { Fail(c, errors.New("dial tcp 10.0.0.1: password=xyz")) })
	w, _ := serve(t, r, "/")
	if w.Code != 500 || strings.Contains(w.Body.String(), "password") {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestFailIncludesRequestID(t *testing.T) {
	r := gin.New()
	r.Use(withRequestID("rid-1"))
	r.GET("/", func(c *gin.Context) { Fail(c, errors.New("x")) })
	_, body := serve(t, r, "/")
	if body["request_id"] != "rid-1" {
		t.Fatalf("request_id = %v", body["request_id"])
	}
}

func TestFailAbortsChain(t *testing.T) {
	ran := false
	r := gin.New()
	r.GET("/", func(c *gin.Context) { Fail(c, errors.New("x")) }, func(c *gin.Context) { ran = true })
	serve(t, r, "/")
	if ran {
		t.Fatal("handlers after Fail must not run")
	}
}
