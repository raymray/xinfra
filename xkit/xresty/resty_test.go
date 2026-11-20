package xresty

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetSendsParamsAndHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Query().Get("q") != "1" || r.Header.Get("X-A") != "b" {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(200)
	}))
	defer srv.Close()
	resp, err := Get(context.Background(), nil, srv.URL, map[string]string{"q": "1"}, map[string]string{"X-A": "b"})
	if err != nil || resp.StatusCode() != 200 {
		t.Fatalf("status=%v err=%v", resp.StatusCode(), err)
	}
}

func TestPostBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		if r.Method != http.MethodPost || string(b) != "hello" {
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(201)
	}))
	defer srv.Close()
	resp, err := Post(context.Background(), nil, srv.URL, "hello", nil)
	if err != nil || resp.StatusCode() != 201 {
		t.Fatalf("status=%v err=%v", resp.StatusCode(), err)
	}
}

func TestDoUsesGivenMethod(t *testing.T) {
	for _, m := range []string{http.MethodPut, http.MethodPatch, http.MethodDelete} {
		t.Run(m, func(t *testing.T) {
			var got string
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { got = r.Method }))
			defer srv.Close()
			if _, err := Do(context.Background(), nil, m, srv.URL, nil, nil); err != nil {
				t.Fatal(err)
			}
			if got != m {
				t.Fatalf("server saw %s, want %s", got, m)
			}
		})
	}
}
