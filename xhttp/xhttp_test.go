package xhttp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/raymray/xinfra/xerror"
)

func golden(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(b))
}

func TestEnvelopeGoldenOK(t *testing.T) {
	b, _ := json.Marshal(NewOK(map[string]int{"id": 1}, "r1"))
	if got, want := string(b), golden(t, "ok.golden.json"); got != want {
		t.Fatalf("\n got %s\nwant %s", got, want)
	}
}

func TestEnvelopeGoldenFail(t *testing.T) {
	b, _ := json.Marshal(NewFail(40001, "not found", "r1"))
	if got, want := string(b), golden(t, "fail.golden.json"); got != want {
		t.Fatalf("\n got %s\nwant %s", got, want)
	}
}

func TestEnvelopeOKNilDataIsEmptyObject(t *testing.T) {
	b, _ := json.Marshal(NewOK(nil, "r1"))
	if !strings.Contains(string(b), `"data":{}`) {
		t.Fatalf("got %s, want data {}", b)
	}
}

var codeNotFound = xerror.NewCode(40001, http.StatusNotFound)

func TestResolveRegisteredCode(t *testing.T) {
	m := NewMapper()
	m.Register(codeNotFound, "资源不存在")
	status, biz, msg := m.Resolve(xerror.WrapCode(codeNotFound, errors.New("row missing")))
	if status != 404 || biz != 40001 || msg != "资源不存在" {
		t.Fatalf("got %d %d %q", status, biz, msg)
	}
}

func TestResolveUnregisteredCodeUsesCodeButHidesDetail(t *testing.T) {
	m := NewMapper()
	status, biz, msg := m.Resolve(xerror.WrapCode(codeNotFound, errors.New("secret detail")))
	if status != 404 || biz != 40001 {
		t.Fatalf("got %d %d", status, biz)
	}
	if strings.Contains(msg, "secret") {
		t.Fatalf("message leaks detail: %q", msg)
	}
}

func TestResolveUnregisteredErrorHidesDetail(t *testing.T) {
	status, biz, msg := NewMapper().Resolve(errors.New("db password leaked"))
	if status != 500 || biz != 50000 || msg != "internal error" {
		t.Fatalf("got %d %d %q", status, biz, msg)
	}
}

func TestResolveNilErrorReturnsOK(t *testing.T) {
	status, biz, msg := NewMapper().Resolve(nil)
	if status != 200 || biz != 0 || msg != "ok" {
		t.Fatalf("got %d %d %q", status, biz, msg)
	}
}

func TestRequestIDRoundTrip(t *testing.T) {
	if got := RequestIDFrom(WithRequestID(context.Background(), "abc")); got != "abc" {
		t.Fatalf("got %q", got)
	}
}

func TestRequestIDMissingReturnsEmpty(t *testing.T) {
	if got := RequestIDFrom(context.Background()); got != "" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveCodeThroughStdWrap(t *testing.T) {
	m := NewMapper()
	m.Register(codeNotFound, "资源不存在")
	status, biz, msg := m.Resolve(fmt.Errorf("get asset 7: %w", xerror.NewCodeErr(codeNotFound)))
	if status != 404 || biz != 40001 || msg != "资源不存在" {
		t.Fatalf("got %d %d %q", status, biz, msg)
	}
}
