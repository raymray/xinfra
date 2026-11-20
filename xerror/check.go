package xerror

import "errors"

// ====================================================================================================
// 内部接口（用于错误链检查）/ Internal interfaces for chain inspection
// ====================================================================================================

type apiCode interface {
	Code() Code
}
type apiCause interface {
	Cause() error
}
type apiStack interface {
	Stack() string
}
type apiCurrent interface {
	Current() error
}
type apiNext interface {
	Next() error
}

type apiError interface {
	Error() string
}

// GetCode 从错误链中提取错误码。无码时返回 CodeNil。
// GetCode extracts the error code from the error chain. Returns CodeNil if no code is set.
// 沿标准库 Unwrap 链继续查找，fmt.Errorf("%w") 包装过的码同样可见；最外层的码优先。
func GetCode(err error) Code {
	for err != nil {
		if e, ok := err.(apiCode); ok {
			if c := e.Code(); c != CodeNil {
				return c
			}
		}
		switch u := err.(type) {
		case interface{ Unwrap() []error }:
			for _, e := range u.Unwrap() {
				if c := GetCode(e); c != CodeNil {
					return c
				}
			}
			return CodeNil
		case interface{ Unwrap() error }:
			err = u.Unwrap()
		default:
			return CodeNil
		}
	}
	return CodeNil
}

// Cause 返回根因错误（错误链最底层）。
// Cause returns the root cause error (bottom of the error chain).
func Cause(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(apiCause); ok {
		return e.Cause()
	}
	return err
}

// Stack 返回完整调用堆栈。不支持堆栈时回退到 err.Error()。
// Stack returns the full call stack as string. Falls back to err.Error() if unsupported.
func Stack(err error) string {
	if err == nil {
		return ""
	}
	if e, ok := err.(apiStack); ok {
		return e.Stack()
	}
	return err.Error()
}

// Current 返回当前层错误。
// Current returns the current-level error.
func Current(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(apiCurrent); ok {
		return e.Current()
	}
	return err
}

// Next 返回错误链中的下一层错误。
// Next returns the next-level error in the chain.
func Next(err error) error {
	if err == nil {
		return nil
	}
	if e, ok := err.(apiNext); ok {
		return e.Next()
	}
	return nil
}

// HasCode 判断错误链中是否包含指定错误码。
// HasCode checks whether the error chain contains the specified error code.
func HasCode(err error, code Code) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(apiCode); ok && e.Code().Code() == code.Code() {
		return true
	}

	if u, ok := err.(interface{ Unwrap() []error }); ok {
		for _, e := range u.Unwrap() {
			if HasCode(e, code) {
				return true
			}
		}
		return false
	}
	if n, ok := err.(apiNext); ok {
		return HasCode(n.Next(), code)
	}
	if u, ok := err.(interface{ Unwrap() error }); ok {
		return HasCode(u.Unwrap(), code)
	}
	return false
}

// AsError 从错误链中提取 *Error。兼容 errors.As。
// AsError extracts *Error from the error chain. Compatible with errors.As.
func AsError(err error) (*Error, bool) {
	if err == nil {
		return nil, false
	}
	var target *Error
	ok := errors.As(err, &target)
	return target, ok
}
