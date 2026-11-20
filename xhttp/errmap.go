package xhttp

import (
	"net/http"
	"sync"

	"github.com/raymray/xinfra/xerror"
)

const (
	CodeInternal    = 50000
	messageInternal = "internal error"
)

// Mapper 把错误映射为 HTTP 状态与响应码。错误码表由各服务注册，Mapper 不含业务含义。
type Mapper struct {
	mu       sync.RWMutex
	messages map[int]string
}

func NewMapper() *Mapper {
	return &Mapper{messages: make(map[int]string)}
}

// Register 登记错误码对外展示的 message。
func (m *Mapper) Register(code xerror.Code, message string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages[code.Code()] = message
}

// Resolve 解析错误。原始错误文本永远不进入返回的 message：
// 已注册的码用注册的 message；带码未注册用 HTTP 状态文本；无码一律 500。
func (m *Mapper) Resolve(err error) (httpStatus, bizCode int, message string) {
	if err == nil {
		return http.StatusOK, 0, "ok"
	}
	code := xerror.GetCode(err)
	if code == xerror.CodeNil {
		return http.StatusInternalServerError, CodeInternal, messageInternal
	}
	status := code.HTTP()
	if status <= 0 {
		status = http.StatusInternalServerError
	}
	m.mu.RLock()
	msg, ok := m.messages[code.Code()]
	m.mu.RUnlock()
	if !ok {
		msg = http.StatusText(status)
	}
	return status, code.Code(), msg
}
