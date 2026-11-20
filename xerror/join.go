package xerror

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

// multiError 是 Join 产生的多错误类型，实现 error 和 Unwrap() []error。
// multiError is the multi-error type produced by Join, implementing error and Unwrap() []error.
type multiError struct {
	errs  []error
	stack stack
	text  string
	code  Code
}

// Error 返回所有错误的拼接字符串，多个错误间用换行分隔。
func (e *multiError) Error() string {
	if e == nil {
		return ""
	}
	var parts []string
	if e.text != "" {
		parts = append(parts, e.text)
	}
	for _, er := range e.errs {
		parts = append(parts, er.Error())
	}
	return strings.Join(parts, "\n")
}

// Code 返回错误码。
func (e *multiError) Code() Code {
	if e == nil {
		return CodeNil
	}
	return e.code
}

// Cause 返回第一个错误的根因。
func (e *multiError) Cause() error {
	if e == nil || len(e.errs) == 0 {
		return nil
	}
	return Cause(e.errs[0])
}

// Current 返回当前层错误（不含底层错误）。
func (e *multiError) Current() error {
	if e == nil {
		return nil
	}
	return &multiError{text: e.text, code: e.code}
}

// Next 返回第一个底层错误。
func (e *multiError) Next() error {
	if e == nil || len(e.errs) == 0 {
		return nil
	}
	return e.errs[0]
}

// Errors 返回所有底层错误。
func (e *multiError) Errors() []error {
	if e == nil {
		return nil
	}
	return e.errs
}

// Stack 返回所有错误的堆栈信息。
func (e *multiError) Stack() string {
	if e == nil {
		return ""
	}
	index := 1
	buf := bytes.NewBuffer(nil)
	writeStackTo(buf, index, e)
	return buf.String()
}

// Format 支持 fmt 格式化，与 Error 一致。
func (e *multiError) Format(s fmt.State, verb rune) {
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

// Unwrap 返回所有底层错误，供 errors.Is / errors.As 遍历。
func (e *multiError) Unwrap() []error {
	if e == nil {
		return nil
	}
	return e.errs
}

// Join 将多个错误合并为一个带堆栈的错误。
// 仅一个非 nil 错误时原样返回；全部为 nil 时返回 nil。
func Join(errs ...error) error {
	return joinErrs(nil, 0, "", nil, errs...)
}

// JoinCode 将多个错误合并为一个带错误码和堆栈的错误。
func JoinCode(code Code, errs ...error) error {
	return joinErrs(code, 0, "", nil, errs...)
}

// JoinSkip 将多个错误合并为一个带堆栈的错误，跳过指定帧数。
func JoinSkip(skip int, errs ...error) error {
	return joinErrs(nil, skip, "", nil, errs...)
}

// JoinCodeSkip 将多个错误合并为一个带错误码和堆栈的错误，跳过指定帧数。
func JoinCodeSkip(code Code, skip int, errs ...error) error {
	return joinErrs(code, skip, "", nil, errs...)
}

// Joinf 将多个错误合并为一个带格式化文本和堆栈的错误。
func Joinf(format string, errs []error, args ...any) error {
	return joinErrs(nil, 0, format, args, errs...)
}

// JoinCodef 将多个错误合并为一个带错误码、格式化文本和堆栈的错误。
func JoinCodef(code Code, format string, errs []error, args ...any) error {
	return joinErrs(code, 0, format, args, errs...)
}

func joinErrs(code Code, skip int, format string, args []any, errs ...error) error {
	var nonNil []error
	for _, e := range errs {
		if e != nil {
			nonNil = append(nonNil, e)
		}
	}
	switch len(nonNil) {
	case 0:
		return nil
	case 1:
		return nonNil[0]
	}
	if code == nil {
		code = CodeNil
	}
	text := ""
	if format != "" {
		text = fmt.Sprintf(format, args...)
	}
	return &multiError{errs: nonNil, stack: callers(skip), code: code, text: text}
}
