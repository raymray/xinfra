package xerror

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

// Error 是核心错误类型，携带 Code、堆栈和底层错误。
// Error is the core error type carrying Code, stack trace, and wrapped error.
type Error struct {
	err   error  // 底层错误 / Underlying wrapped error.
	stack stack  // 创建时捕获的调用堆栈 / Call stack captured at creation time.
	text  string // 当前层错误文本 / Current-level error text.
	code  Code   // 错误码，未设置时为 CodeNil / Error code, CodeNil if unset.
}

const defaultSkip = 2

// Error 实现 error 接口，返回完整错误字符串。
// Error implements error interface, returns full error string.
func (e *Error) Error() string {
	if e == nil {
		return ""
	}
	s := e.text
	if e.err != nil {
		if s != "" {
			s += ": "
		}
		s += e.err.Error()
	}
	return s
}

// Code 返回错误码。未设置时返回 CodeNil。
// Code returns the error code. Returns CodeNil if unset.
func (e *Error) Code() Code {
	if e == nil {
		return CodeNil
	}
	return e.code
}

// Cause 返回根因错误（错误链最底层）。
// Cause returns the root cause error (bottom of the error chain).
func (e *Error) Cause() error {
	if e == nil {
		return nil
	}
	loop := e
	for loop != nil {
		if loop.err != nil {
			if next, ok := errors.AsType[*Error](loop.err); ok {
				loop = next
			} else {
				return loop.err
			}
		} else {
			return errors.New(loop.text)
		}
	}
	return nil
}

// Current 返回当前层错误（不包含底层错误）。
// Current returns the current-level error without underlying error.
func (e *Error) Current() error {
	if e == nil {
		return nil
	}
	return &Error{stack: e.stack, text: e.text, code: e.code}
}

// Next 返回错误链中的下一层错误。
// Next returns the next-level error in the chain.
func (e *Error) Next() error {
	if e == nil {
		return nil
	}
	return e.err
}

// Stack 返回完整调用堆栈的字符串表示。
// Stack returns the full call stack as a formatted string.
func (e *Error) Stack() string {
	if e == nil {
		return ""
	}
	index := 1
	buf := bytes.NewBuffer(nil)
	writeStackTo(buf, index, e)
	return buf.String()
}

// Format 支持 fmt 格式化：
//
//	%v, %s   : 完整错误字符串 / Full error string
//	%-v, %-s : 仅当前层错误字符串 / Current-level error string only
//	%+v      : 错误字符串 + 完整堆栈 / Error string + full stack trace
//	%+s      : 仅完整堆栈 / Full stack trace only
func (e *Error) Format(s fmt.State, verb rune) {
	switch verb {
	case 's', 'v':
		switch {
		case s.Flag('-'):
			if e.text != "" {
				_, _ = io.WriteString(s, e.text)
			} else {
				_, _ = io.WriteString(s, e.Error())
			}
		case s.Flag('+'):
			if verb == 's' {
				_, _ = io.WriteString(s, e.Stack())
			} else {
				_, _ = io.WriteString(s, e.Error()+"\n"+e.Stack())
			}
		default:
			_, _ = io.WriteString(s, e.Error())
		}
	}
}

// Unwrap 支持 errors.Is / errors.As 链式匹配。
// Unwrap supports errors.Is / errors.As chain matching.
func (e *Error) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.err
}
