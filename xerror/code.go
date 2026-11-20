package xerror

import "fmt"

// Code 错误码接口，外部包通过 NewCode() 创建自定义错误码。
// Code is the error code interface. External packages create custom codes via NewCode().
type Code interface {
	Code() int
	HTTP() int
	String() string
}

// localCode 是 Code 的内部实现。
// localCode implements Code for internal usage.
type localCode struct {
	code int
	http int
}

// NewCode 创建自定义错误码，供外部包使用。
// NewCode creates a custom error code for use in any package.
//
//	var ErrUserNotFound = xerror.NewCode(10010, 404)
func NewCode(code int, http int) Code {
	return localCode{code: code, http: http}
}

func (c localCode) Code() int      { return c.code }
func (c localCode) HTTP() int      { return c.http }
func (c localCode) String() string { return fmt.Sprintf("%d", c.code) }

// CodeNil 表示无错误码。
// CodeNil indicates no error code.
var CodeNil Code = localCode{-1, 0}
