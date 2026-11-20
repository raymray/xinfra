package xerror

import "fmt"

// New 创建带堆栈的错误。
// New creates an error with stack trace.
func New(text string) error {
	return &Error{stack: callers(), text: text, code: CodeNil}
}

// Newf 创建带堆栈和格式化消息的错误。
// Newf creates an error with stack trace and formatted message.
func Newf(format string, args ...any) error {
	return &Error{stack: callers(), text: fmt.Sprintf(format, args...), code: CodeNil}
}

// Wrap 包裹错误并附加文本和堆栈。err 为 nil 时返回 nil。
// Wrap wraps an error with text and stack trace. Returns nil if err is nil.
func Wrap(err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{err: err, stack: callers(), text: text, code: GetCode(err)}
}

// Wrapf 包裹错误并附加格式化文本和堆栈。err 为 nil 时返回 nil。
// Wrapf wraps an error with formatted text and stack trace. Returns nil if err is nil.
func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{err: err, stack: callers(), text: fmt.Sprintf(format, args...), code: GetCode(err)}
}

// NewCodeErr 创建带错误码和堆栈的错误。
// NewCodeErr creates an error with the given code and stack trace.
func NewCodeErr(code Code, text ...string) error {
	msg := ""
	if len(text) > 0 {
		msg = text[0]
	}
	return &Error{stack: callers(), text: msg, code: code}
}

// NewCodef 创建带错误码、格式化消息和堆栈的错误。
// NewCodef creates an error with the given code, formatted message, and stack trace.
func NewCodef(code Code, format string, args ...any) error {
	return &Error{stack: callers(), text: fmt.Sprintf(format, args...), code: code}
}

// WrapCode 用错误码包裹已有错误。err 为 nil 时返回 nil。
// WrapCode wraps an error with a code and text. Returns nil if err is nil.
func WrapCode(code Code, err error, text ...string) error {
	if err == nil {
		return nil
	}
	msg := ""
	if len(text) > 0 {
		msg = text[0]
	}
	return &Error{err: err, stack: callers(), text: msg, code: code}
}

// WrapCodef 用错误码包裹已有错误（格式化文本）。err 为 nil 时返回 nil。
// WrapCodef wraps an error with a code and formatted text. Returns nil if err is nil.
func WrapCodef(code Code, err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	return &Error{err: err, stack: callers(), text: fmt.Sprintf(format, args...), code: code}
}

// NewSkip 创建带堆栈的错误，跳过指定帧数。
// NewSkip creates an error with stack trace, skipping the specified number of frames.
func NewSkip(skip int, text string) error {
	return &Error{stack: callers(skip), text: text, code: CodeNil}
}

// WrapSkip 包裹错误并附加文本，跳过指定堆栈帧数。
// WrapSkip wraps an error with text, skipping the specified number of stack frames.
func WrapSkip(skip int, err error, text string) error {
	if err == nil {
		return nil
	}
	return &Error{err: err, stack: callers(skip), text: text, code: GetCode(err)}
}

// NewCodeSkip 创建带错误码和堆栈的错误，跳过指定帧数。
// NewCodeSkip creates an error with code and stack trace, skipping specified frames.
func NewCodeSkip(code Code, skip int, text ...string) error {
	msg := ""
	if len(text) > 0 {
		msg = text[0]
	}
	return &Error{stack: callers(skip), text: msg, code: code}
}

// WrapCodeSkip 用错误码包裹错误，跳过指定堆栈帧数。
// WrapCodeSkip wraps an error with code, skipping specified stack frames.
func WrapCodeSkip(code Code, skip int, err error, text ...string) error {
	if err == nil {
		return nil
	}
	msg := ""
	if len(text) > 0 {
		msg = text[0]
	}
	return &Error{err: err, stack: callers(skip), text: msg, code: code}
}
