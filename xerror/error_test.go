package xerror

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ====================================================================================================
// Code
// ====================================================================================================

func TestNewCode(t *testing.T) {
	code := NewCode(10010, 404)
	assert.Equal(t, 10010, code.Code())
	assert.Equal(t, 404, code.HTTP())
	assert.Equal(t, "10010", code.String())
}

func TestCodeNil(t *testing.T) {
	assert.Equal(t, -1, CodeNil.Code())
	assert.Equal(t, 0, CodeNil.HTTP())
}

// ====================================================================================================
// New / Newf
// ====================================================================================================

func TestNew(t *testing.T) {
	err := New("test error")
	assert.NotNil(t, err)
	assert.Equal(t, "test error", err.Error())

	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, "test error", e.text)
	assert.Equal(t, CodeNil, e.code)
	assert.NotEmpty(t, string(e.stack))
}

func TestNewf(t *testing.T) {
	err := Newf("hello %s, count=%d", "world", 42)
	assert.NotNil(t, err)
	assert.Equal(t, "hello world, count=42", err.Error())
}

func TestNewSkip(t *testing.T) {
	err := NewSkip(3, "skipped error")
	assert.NotNil(t, err)
	assert.Equal(t, "skipped error", err.Error())
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.NotEmpty(t, string(e.stack))
}

// ====================================================================================================
// Wrap / Wrapf
// ====================================================================================================

func TestWrap(t *testing.T) {
	inner := errors.New("inner error")
	err := Wrap(inner, "outer")
	assert.NotNil(t, err)
	assert.Equal(t, "outer: inner error", err.Error())
	assert.True(t, errors.Is(err, inner))

	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, inner, e.err)
}

func TestWrap_Nil(t *testing.T) {
	assert.Nil(t, Wrap(nil, "text"))
}

func TestWrapf(t *testing.T) {
	inner := errors.New("db failed")
	err := Wrapf(inner, "query %s failed", "SELECT")
	assert.NotNil(t, err)
	assert.Equal(t, "query SELECT failed: db failed", err.Error())
	assert.True(t, errors.Is(err, inner))
}

func TestWrapf_Nil(t *testing.T) {
	assert.Nil(t, Wrapf(nil, "text %d", 1))
}

func TestWrapSkip(t *testing.T) {
	inner := errors.New("inner")
	err := WrapSkip(3, inner, "outer")
	assert.NotNil(t, err)
	assert.Equal(t, "outer: inner", err.Error())
	assert.True(t, errors.Is(err, inner))
}

func TestWrapSkip_Nil(t *testing.T) {
	assert.Nil(t, WrapSkip(3, nil, "text"))
}

// ====================================================================================================
// NewCode* / WrapCode*
// ====================================================================================================

func TestNewCodeErr(t *testing.T) {
	code := NewCode(10010, 404)
	err := NewCodeErr(code, "user not found")
	assert.NotNil(t, err)
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, code, e.code)
	assert.Equal(t, "user not found", e.text)
}

func TestNewCodeErr_NoText(t *testing.T) {
	code := NewCode(10010, 404)
	err := NewCodeErr(code)
	assert.NotNil(t, err)
	assert.Equal(t, "", err.Error())
}

func TestNewCodef(t *testing.T) {
	code := NewCode(500, 500)
	err := NewCodef(code, "server error: %s", "timeout")
	assert.NotNil(t, err)
	assert.Equal(t, "server error: timeout", err.Error())
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, code, e.code)
}

func TestWrapCode(t *testing.T) {
	code := NewCode(20001, 400)
	inner := errors.New("invalid input")
	err := WrapCode(code, inner, "param error")
	assert.NotNil(t, err)
	assert.Equal(t, "param error: invalid input", err.Error())
	assert.True(t, errors.Is(err, inner))
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, code, e.code)
}

func TestWrapCode_Nil(t *testing.T) {
	code := NewCode(20001, 400)
	assert.Nil(t, WrapCode(code, nil, "text"))
}

func TestWrapCode_NoText(t *testing.T) {
	code := NewCode(20001, 400)
	inner := errors.New("inner")
	err := WrapCode(code, inner)
	assert.NotNil(t, err)
	assert.Equal(t, "inner", err.Error())
}

func TestWrapCodef(t *testing.T) {
	code := NewCode(20001, 400)
	inner := errors.New("bad field")
	err := WrapCodef(code, inner, "field %s invalid", "name")
	assert.NotNil(t, err)
	assert.Equal(t, "field name invalid: bad field", err.Error())
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, code, e.code)
}

func TestWrapCodef_Nil(t *testing.T) {
	assert.Nil(t, WrapCodef(NewCode(1, 1), nil, "text"))
}

func TestNewCodeSkip(t *testing.T) {
	code := NewCode(30001, 500)
	err := NewCodeSkip(code, 3, "timeout")
	assert.NotNil(t, err)
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, code, e.code)
}

func TestWrapCodeSkip(t *testing.T) {
	code := NewCode(30001, 500)
	inner := errors.New("conn refused")
	err := WrapCodeSkip(code, 3, inner, "dial failed")
	assert.NotNil(t, err)
	assert.Equal(t, "dial failed: conn refused", err.Error())
	assert.True(t, errors.Is(err, inner))
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, code, e.code)
}

func TestWrapCodeSkip_Nil(t *testing.T) {
	assert.Nil(t, WrapCodeSkip(NewCode(1, 1), 3, nil, "text"))
}

// ====================================================================================================
// Error methods
// ====================================================================================================

func TestError_NilReceiver(t *testing.T) {
	var e *Error
	assert.Equal(t, "", e.Error())
	assert.Equal(t, CodeNil, e.Code())
	assert.Nil(t, e.Cause())
	assert.Nil(t, e.Current())
	assert.Nil(t, e.Next())
	assert.Equal(t, "", e.Stack())
	assert.Nil(t, e.Unwrap())
}

func TestError_Error_WithUnderlying(t *testing.T) {
	err := &Error{text: "outer", err: errors.New("inner")}
	assert.Equal(t, "outer: inner", err.Error())
}

func TestError_Error_TextOnly(t *testing.T) {
	err := &Error{text: "standalone"}
	assert.Equal(t, "standalone", err.Error())
}

func TestError_Error_NoText(t *testing.T) {
	err := &Error{err: errors.New("inner")}
	assert.Equal(t, "inner", err.Error())
}

func TestError_Current(t *testing.T) {
	inner := errors.New("inner")
	err := Wrap(inner, "outer")
	current := Current(err)
	assert.Equal(t, "outer", current.Error())
	assert.True(t, errors.Is(err, inner))
	assert.False(t, errors.Is(current, inner))
}

func TestError_Next(t *testing.T) {
	inner := errors.New("inner")
	wrapped := Wrap(inner, "outer")
	assert.Equal(t, inner, Next(wrapped))

	standalone := New("standalone")
	assert.Nil(t, Next(standalone))
}

func TestError_Cause(t *testing.T) {
	root := errors.New("root")
	mid := Wrap(root, "mid")
	top := Wrap(mid, "top")
	assert.Equal(t, "root", Cause(top).Error())

	standalone := New("standalone")
	cause := Cause(standalone)
	assert.Equal(t, "standalone", cause.Error())
}

func TestError_Unwrap(t *testing.T) {
	inner := errors.New("inner")
	err := Wrap(inner, "outer")
	assert.Equal(t, inner, err.(*Error).Unwrap())
}

func TestError_Stack(t *testing.T) {
	err := New("stack test")
	stack := Stack(err)
	assert.Contains(t, stack, "stack test")
}

func TestError_Format(t *testing.T) {
	inner := errors.New("inner")
	err := Wrap(inner, "outer")

	assert.Equal(t, "outer: inner", fmt.Sprintf("%v", err))
	assert.Equal(t, "outer: inner", fmt.Sprintf("%s", err))
	assert.Equal(t, "outer", fmt.Sprintf("%-v", err))
	assert.Equal(t, "outer", fmt.Sprintf("%-s", err))

	full := fmt.Sprintf("%+v", err)
	assert.Contains(t, full, "outer: inner")
	assert.Contains(t, full, "outer")

	stackOnly := fmt.Sprintf("%+s", err)
	assert.Contains(t, stackOnly, "outer")
}

// ====================================================================================================
// Check functions
// ====================================================================================================

func TestGetCode(t *testing.T) {
	code := NewCode(10010, 404)
	err := NewCodeErr(code, "test")
	assert.Equal(t, code, GetCode(err))
	assert.Equal(t, CodeNil, GetCode(New("no code")))
	assert.Equal(t, CodeNil, GetCode(nil))
	assert.Equal(t, CodeNil, GetCode(errors.New("std error")))
}

func TestCause(t *testing.T) {
	root := errors.New("root")
	wrapped := Wrap(root, "wrapped")
	assert.Equal(t, root, Cause(wrapped))
	assert.Equal(t, root, Cause(root))
	assert.Nil(t, Cause(nil))
}

func TestStack(t *testing.T) {
	err := New("stack test")
	s := Stack(err)
	assert.Contains(t, s, "stack test")

	assert.Equal(t, "", Stack(nil))
	assert.Equal(t, "plain error", Stack(errors.New("plain error")))
}

func TestCurrent(t *testing.T) {
	inner := errors.New("inner")
	wrapped := Wrap(inner, "outer")
	current := Current(wrapped)
	assert.Equal(t, "outer", current.Error())

	assert.Nil(t, Current(nil))
	assert.Equal(t, "std", Current(errors.New("std")).Error())
}

func TestNext(t *testing.T) {
	inner := errors.New("inner")
	wrapped := Wrap(inner, "outer")
	assert.Equal(t, inner, Next(wrapped))

	assert.Nil(t, Next(nil))
	assert.Nil(t, Next(New("standalone")))
	assert.Nil(t, Next(errors.New("std")))
}

func TestHasCode(t *testing.T) {
	code := NewCode(10010, 404)
	err := NewCodeErr(code, "test")
	assert.True(t, HasCode(err, code))
	assert.False(t, HasCode(err, NewCode(99999, 500)))
	assert.False(t, HasCode(nil, code))

	// Wrapped error inherits code
	inner := NewCodeErr(code, "inner")
	wrapped := Wrap(inner, "outer")
	assert.True(t, HasCode(wrapped, code))
}

func TestAsError(t *testing.T) {
	err := New("test")
	e, ok := AsError(err)
	assert.True(t, ok)
	assert.Equal(t, "test", e.text)

	e, ok = AsError(nil)
	assert.False(t, ok)
	assert.Nil(t, e)

	e, ok = AsError(errors.New("std"))
	assert.False(t, ok)
	assert.Nil(t, e)
}

// ====================================================================================================
// errors.Is / errors.As compatibility
// ====================================================================================================

func TestErrorsIs(t *testing.T) {
	target := errors.New("target")
	err := Wrap(target, "wrapper")
	assert.True(t, errors.Is(err, target))
	assert.False(t, errors.Is(err, errors.New("other")))

	deep := Wrap(Wrap(Wrap(target, "l3"), "l2"), "l1")
	assert.True(t, errors.Is(deep, target))
}

func TestErrorsAs(t *testing.T) {
	err := New("test")
	var e *Error
	assert.True(t, errors.As(err, &e))
	assert.Equal(t, "test", e.text)

	inner := errors.New("std")
	wrapped := Wrap(inner, "outer")
	assert.True(t, errors.Is(wrapped, inner))
}

// ====================================================================================================
// Join
// ====================================================================================================

func TestJoin_Nil(t *testing.T) {
	assert.Nil(t, Join())
	assert.Nil(t, Join(nil, nil, nil))
}

func TestJoin_Single(t *testing.T) {
	single := errors.New("single")
	joined := Join(single)
	assert.Equal(t, single, joined)
}

func TestJoin_Multiple(t *testing.T) {
	e1 := New("error 1")
	e2 := New("error 2")
	e3 := errors.New("error 3")

	joined := Join(e1, e2, e3, nil)
	assert.NotNil(t, joined)

	msg := joined.Error()
	assert.Contains(t, msg, "error 1")
	assert.Contains(t, msg, "error 2")
	assert.Contains(t, msg, "error 3")

	assert.True(t, errors.Is(joined, e1))
	assert.True(t, errors.Is(joined, e2))
	assert.True(t, errors.Is(joined, e3))
}

func TestJoin_ErrorsMethod(t *testing.T) {
	e1 := New("e1")
	e2 := errors.New("e2")
	joined := Join(e1, e2)
	assert.NotNil(t, joined)

	var m *multiError
	assert.True(t, errors.As(joined, &m))
	errs := m.Errors()
	assert.Len(t, errs, 2)
	assert.Equal(t, e1, errs[0])
	assert.Equal(t, e2, errs[1])
}

func TestJoin_Stack(t *testing.T) {
	e1 := New("first")
	e2 := New("second")
	joined := Join(e1, e2)

	stack := Stack(joined)
	assert.Contains(t, stack, "first")
	assert.Contains(t, stack, "second")
}

func TestJoin_Cause(t *testing.T) {
	root := errors.New("root cause")
	e1 := Wrap(root, "layer1")
	e2 := New("standalone")
	joined := Join(e1, e2)

	cause := Cause(joined)
	assert.Equal(t, root, cause)
}

func TestJoin_Current(t *testing.T) {
	e1 := New("e1")
	e2 := errors.New("e2")
	joined := Join(e1, e2)

	current := Current(joined)
	assert.Equal(t, "", current.Error())
	assert.False(t, errors.Is(current, e1))
	assert.False(t, errors.Is(current, e2))
}

func TestJoin_Next(t *testing.T) {
	e1 := New("e1")
	e2 := errors.New("e2")
	joined := Join(e1, e2)
	assert.Equal(t, e1, Next(joined))
}

func TestJoin_Format(t *testing.T) {
	e1 := New("err1")
	e2 := New("err2")
	joined := Join(e1, e2)

	v := fmt.Sprintf("%v", joined)
	assert.Contains(t, v, "err1")
	assert.Contains(t, v, "err2")

	pv := fmt.Sprintf("%+v", joined)
	assert.Contains(t, pv, "err1")
	assert.Contains(t, pv, "err2")
}

func TestJoin_Unwrap(t *testing.T) {
	e1 := errors.New("e1")
	e2 := errors.New("e2")
	joined := Join(e1, e2)

	var m *multiError
	assert.True(t, errors.As(joined, &m))
	unwrapped := m.Unwrap()
	assert.Equal(t, []error{e1, e2}, unwrapped)
}

func TestJoin_NilReceiver(t *testing.T) {
	var m *multiError
	assert.Nil(t, m.Errors())
	assert.Nil(t, m.Unwrap())
	assert.Equal(t, "", m.Error())
	assert.Equal(t, CodeNil, m.Code())
	assert.Nil(t, m.Cause())
	assert.Nil(t, m.Current())
	assert.Nil(t, m.Next())
	assert.Equal(t, "", m.Stack())
}

// ====================================================================================================
// JoinCode
// ====================================================================================================

func TestJoinCode(t *testing.T) {
	code := NewCode(500, 500)
	e1 := New("err1")
	e2 := New("err2")

	err := JoinCode(code, e1, e2)
	assert.NotNil(t, err)
	assert.True(t, HasCode(err, code))
	assert.True(t, errors.Is(err, e1))
	assert.True(t, errors.Is(err, e2))
}

func TestJoinCode_Nil(t *testing.T) {
	code := NewCode(500, 500)
	assert.Nil(t, JoinCode(code))
	assert.Nil(t, JoinCode(code, nil))
}

func TestJoinCode_Single(t *testing.T) {
	code := NewCode(500, 500)
	e1 := New("err1")
	err := JoinCode(code, e1)
	assert.Equal(t, e1, err)
}

// ====================================================================================================
// JoinSkip
// ====================================================================================================

func TestJoinSkip(t *testing.T) {
	e1 := New("e1")
	e2 := New("e2")
	err := JoinSkip(3, e1, e2)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, e1))
	assert.True(t, errors.Is(err, e2))
}

// ====================================================================================================
// JoinCodeSkip
// ====================================================================================================

func TestJoinCodeSkip(t *testing.T) {
	code := NewCode(500, 500)
	e1 := New("e1")
	e2 := New("e2")
	err := JoinCodeSkip(code, 3, e1, e2)
	assert.NotNil(t, err)
	assert.True(t, HasCode(err, code))
	assert.True(t, errors.Is(err, e1))
	assert.True(t, errors.Is(err, e2))
}

// ====================================================================================================
// HasCode with multiError
// ====================================================================================================

func TestHasCode_JoinedErrors(t *testing.T) {
	code := NewCode(10010, 404)
	e1 := NewCodeErr(code, "err1")
	e2 := New("err2")

	joined := Join(e1, e2)
	assert.True(t, HasCode(joined, code))
}

func TestHasCode_JoinedDeep(t *testing.T) {
	code := NewCode(20001, 500)
	root := NewCodeErr(code, "root")
	wrapped := Wrap(root, "wrapped")

	joined := Join(wrapped, New("other"))
	assert.True(t, HasCode(joined, code))
}

// ====================================================================================================
// Mixed: Wrap chain + Join
// ====================================================================================================

func TestWrapThenJoin(t *testing.T) {
	root1 := errors.New("db timeout")
	root2 := errors.New("cache miss")
	wrapped1 := Wrap(root1, "db query failed")
	wrapped2 := Wrap(root2, "cache lookup failed")

	joined := Join(wrapped1, wrapped2)
	assert.NotNil(t, joined)
	assert.True(t, errors.Is(joined, root1))
	assert.True(t, errors.Is(joined, root2))

	stack := Stack(joined)
	assert.Contains(t, stack, "db query failed")
	assert.Contains(t, stack, "cache lookup failed")
}

func TestJoinOfJoins(t *testing.T) {
	a := errors.New("a")
	b := errors.New("b")
	c := errors.New("c")
	d := errors.New("d")
	e1 := Join(a, b)
	e2 := Join(c, d)

	joined := Join(e1, e2)
	assert.NotNil(t, joined)

	assert.True(t, errors.Is(joined, a))
	assert.True(t, errors.Is(joined, b))
	assert.True(t, errors.Is(joined, c))
	assert.True(t, errors.Is(joined, d))
}

// ====================================================================================================
// Error format separator
// ====================================================================================================

func TestJoin_ErrorFormat(t *testing.T) {
	e1 := errors.New("aaa")
	e2 := errors.New("bbb")
	joined := Join(e1, e2)

	msg := joined.Error()
	parts := strings.Split(msg, "\n")
	assert.Len(t, parts, 2)
	assert.Equal(t, "aaa", strings.TrimSpace(parts[0]))
	assert.Equal(t, "bbb", strings.TrimSpace(parts[1]))
}

func TestJoin_ErrorFormat_WithText(t *testing.T) {
	e1 := errors.New("a")
	e2 := errors.New("b")
	joined := &multiError{text: "multiple failures", errs: []error{e1, e2}}

	msg := joined.Error()
	assert.Contains(t, msg, "multiple failures")
	assert.Contains(t, msg, "a")
	assert.Contains(t, msg, "b")
}

// ====================================================================================================
// Joinf / JoinCodef
// ====================================================================================================

func TestJoinf(t *testing.T) {
	e1 := New("err1")
	e2 := New("err2")

	err := Joinf("batch failed: %d errors", []error{e1, e2}, 2)
	assert.NotNil(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "batch failed: 2 errors")
	assert.Contains(t, msg, "err1")
	assert.Contains(t, msg, "err2")
	assert.True(t, errors.Is(err, e1))
	assert.True(t, errors.Is(err, e2))
}

func TestJoinf_Nil(t *testing.T) {
	assert.Nil(t, Joinf("text", nil))
	assert.Nil(t, Joinf("text", []error{nil}))
}

func TestJoinf_Single(t *testing.T) {
	e1 := New("err1")
	err := Joinf("text", []error{e1})
	assert.Equal(t, e1, err)
}

func TestJoinCodef(t *testing.T) {
	code := NewCode(500, 500)
	e1 := New("err1")
	e2 := New("err2")

	err := JoinCodef(code, "batch: %s", []error{e1, e2}, "timeout")
	assert.NotNil(t, err)
	msg := err.Error()
	assert.Contains(t, msg, "batch: timeout")
	assert.True(t, HasCode(err, code))
	assert.True(t, errors.Is(err, e1))
	assert.True(t, errors.Is(err, e2))
}

func TestJoinCodef_Nil(t *testing.T) {
	code := NewCode(500, 500)
	assert.Nil(t, JoinCodef(code, "text", nil))
}

func TestJoinCodef_Single(t *testing.T) {
	code := NewCode(500, 500)
	e1 := New("err1")
	err := JoinCodef(code, "text", []error{e1})
	assert.Equal(t, e1, err)
}

func TestJoinf_Stack(t *testing.T) {
	e1 := New("first")
	e2 := New("second")
	err := Joinf("multi: %d items", []error{e1, e2}, 2)

	stack := Stack(err)
	assert.Contains(t, stack, "multi: 2 items")
	assert.Contains(t, stack, "first")
	assert.Contains(t, stack, "second")
}
