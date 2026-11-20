package xgin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/raymray/xinfra/xhttp"
	"github.com/raymray/xinfra/xlog"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

func observed() (*zap.Logger, *observer.ObservedLogs) {
	core, logs := observer.New(zap.DebugLevel)
	return zap.New(core), logs
}

func do(r *gin.Engine, path string, header map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	for k, v := range header {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func fieldOf(e observer.LoggedEntry, key string) (any, bool) {
	m := e.ContextMap()
	v, ok := m[key]
	return v, ok
}

func TestTraceGeneratesIDWhenAbsent(t *testing.T) {
	var seen string
	r := gin.New()
	r.Use(Trace(zap.NewNop()))
	r.GET("/", func(c *gin.Context) { seen = xhttp.RequestIDFrom(c.Request.Context()) })
	w := do(r, "/", nil)
	if seen == "" || w.Header().Get("X-Request-ID") != seen {
		t.Fatalf("ctx id %q, header %q", seen, w.Header().Get("X-Request-ID"))
	}
}

func TestTracePassesThroughInboundID(t *testing.T) {
	var seen string
	r := gin.New()
	r.Use(Trace(zap.NewNop()))
	r.GET("/", func(c *gin.Context) { seen = xhttp.RequestIDFrom(c.Request.Context()) })
	w := do(r, "/", map[string]string{"X-Request-ID": "in-1"})
	if seen != "in-1" || w.Header().Get("X-Request-ID") != "in-1" {
		t.Fatalf("ctx %q header %q", seen, w.Header().Get("X-Request-ID"))
	}
}

func TestTraceCustomHeaderAndGenerator(t *testing.T) {
	r := gin.New()
	r.Use(Trace(zap.NewNop(), WithRequestIDHeader("X-Trace"), WithIDGenerator(func() string { return "gen" })))
	r.GET("/", func(c *gin.Context) {})
	if got := do(r, "/", nil).Header().Get("X-Trace"); got != "gen" {
		t.Fatalf("X-Trace = %q", got)
	}
}

func TestTraceInjectsLoggerWithRequestID(t *testing.T) {
	base, logs := observed()
	r := gin.New()
	r.Use(Trace(base))
	r.GET("/", func(c *gin.Context) { xlog.FromContext(c.Request.Context()).Info("biz") })
	do(r, "/", map[string]string{"X-Request-ID": "rid-9"})
	entries := logs.FilterMessage("biz").All()
	if len(entries) != 1 {
		t.Fatalf("want 1 biz log, got %d", len(entries))
	}
	if v, _ := fieldOf(entries[0], "request_id"); v != "rid-9" {
		t.Fatalf("request_id field = %v", v)
	}
}

func TestRecoveryReturnsEnvelope500WithRequestID(t *testing.T) {
	r := gin.New()
	r.Use(Recovery(zap.NewNop()), Trace(zap.NewNop()))
	r.GET("/", func(c *gin.Context) { panic("boom") })
	w := do(r, "/", map[string]string{"X-Request-ID": "rid-p"})
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if w.Code != 500 || body["code"].(float64) != xhttp.CodeInternal || body["request_id"] != "rid-p" {
		t.Fatalf("got %d %s", w.Code, w.Body.String())
	}
}

func TestRecoveryWithoutTraceStillHasRequestID(t *testing.T) {
	r := gin.New()
	r.Use(Recovery(zap.NewNop()))
	r.GET("/", func(c *gin.Context) { panic("boom") })
	w := do(r, "/", nil)
	var body map[string]any
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	id, _ := body["request_id"].(string)
	if w.Code != 500 || id == "" || w.Header().Get("X-Request-ID") != id {
		t.Fatalf("got %d body %s header %q", w.Code, w.Body.String(), w.Header().Get("X-Request-ID"))
	}
}

func TestRecoveryLogsStack(t *testing.T) {
	l, logs := observed()
	r := gin.New()
	r.Use(Recovery(l))
	r.GET("/", func(c *gin.Context) { panic("boom") })
	do(r, "/", nil)
	all := logs.All()
	if len(all) != 1 || all[0].Level != zap.ErrorLevel {
		t.Fatalf("want one error log, got %+v", all)
	}
	if s, _ := all[0].ContextMap()["stack"].(string); !strings.Contains(s, "Recovery") {
		t.Fatalf("want one error log with stack, got %+v", all)
	}
}

func TestAccessLogFields(t *testing.T) {
	l, logs := observed()
	r := gin.New()
	r.Use(Trace(zap.NewNop()), AccessLog(l))
	r.GET("/x", func(c *gin.Context) { c.Status(204) })
	do(r, "/x", map[string]string{"X-Request-ID": "rid-a"})
	all := logs.All()
	if len(all) != 1 {
		t.Fatalf("want 1 access log, got %d", len(all))
	}
	m := all[0].ContextMap()
	if m["method"] != "GET" || m["path"] != "/x" || m["status"] != int64(204) || m["request_id"] != "rid-a" {
		t.Fatalf("fields = %v", m)
	}
	if _, ok := m["latency"]; !ok {
		t.Fatal("latency missing")
	}
}

func TestAccessLogSkipsPaths(t *testing.T) {
	l, logs := observed()
	r := gin.New()
	r.Use(AccessLog(l, WithSkipPaths("/healthz")))
	r.GET("/healthz", func(c *gin.Context) {})
	do(r, "/healthz", nil)
	if logs.Len() != 0 {
		t.Fatalf("skipped path logged %d entries", logs.Len())
	}
}

func TestAccessLogExtraFields(t *testing.T) {
	l, logs := observed()
	r := gin.New()
	r.Use(AccessLog(l, WithLogFields(func(*gin.Context) []zap.Field { return []zap.Field{zap.String("tenant", "t1")} })))
	r.GET("/", func(c *gin.Context) {})
	do(r, "/", nil)
	if logs.Len() != 1 || logs.All()[0].ContextMap()["tenant"] != "t1" {
		t.Fatalf("extra field missing: %+v", logs.All())
	}
}

func TestRecoveryAfterPartialWriteDoesNotAppendEnvelope(t *testing.T) {
	r := gin.New()
	r.Use(Recovery(zap.NewNop()))
	r.GET("/", func(c *gin.Context) {
		c.String(200, "partial")
		c.Writer.Flush()
		panic("boom")
	})
	w := do(r, "/", nil)
	if w.Body.String() != "partial" {
		t.Fatalf("body corrupted: %q", w.Body.String())
	}
}

func TestRecoveryRepanicsErrAbortHandler(t *testing.T) {
	r := gin.New()
	r.Use(Recovery(zap.NewNop()))
	r.GET("/", func(c *gin.Context) { panic(http.ErrAbortHandler) })
	defer func() {
		if rec := recover(); rec != http.ErrAbortHandler {
			t.Fatalf("want ErrAbortHandler re-panicked, got %v", rec)
		}
	}()
	do(r, "/", nil)
}
