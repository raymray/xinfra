// Package xresty 是 resty 的薄封装，统一 ctx、query 与 header 的设置方式。
package xresty

import (
	"context"
	"net/http"

	"github.com/go-resty/resty/v2"
)

func NewClient() *resty.Client {
	return resty.New()
}

// Get 发送 GET；req 为 nil 时使用新的默认客户端。
func Get(ctx context.Context, req *resty.Request, url string, params, headers map[string]string) (*resty.Response, error) {
	if req == nil {
		req = NewClient().R()
	}
	if params != nil {
		req = req.SetQueryParams(params)
	}
	return Do(ctx, req, http.MethodGet, url, nil, headers)
}

func Post(ctx context.Context, req *resty.Request, url string, body any, headers map[string]string) (*resty.Response, error) {
	return Do(ctx, req, http.MethodPost, url, body, headers)
}

func Do(ctx context.Context, req *resty.Request, method, url string, body any, headers map[string]string) (*resty.Response, error) {
	if req == nil {
		req = NewClient().R()
	}
	if headers != nil {
		req = req.SetHeaders(headers)
	}
	req = req.SetContext(ctx)
	if body != nil {
		req = req.SetBody(body)
	}
	return req.Execute(method, url)
}
