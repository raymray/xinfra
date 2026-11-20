// Package xhttp 定义各服务共用的 HTTP 响应契约，不依赖具体 Web 框架。
package xhttp

// Envelope 是统一响应信封。字段名属于对外契约，改动即 major 版本变更。
type Envelope struct {
	Code      int    `json:"code"`
	Message   string `json:"message"`
	Data      any    `json:"data,omitempty"`
	RequestID string `json:"request_id"`
}

// NewOK 构造成功响应；data 为 nil 时输出空对象，保证成功响应总带 data。
func NewOK(data any, requestID string) Envelope {
	if data == nil {
		data = struct{}{}
	}
	return Envelope{Code: 0, Message: "ok", Data: data, RequestID: requestID}
}

// NewFail 构造失败响应，不带 data。
func NewFail(code int, message, requestID string) Envelope {
	return Envelope{Code: code, Message: message, RequestID: requestID}
}
