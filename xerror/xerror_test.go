package xerror

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func TestWrapfKeepsBothMessages(t *testing.T) {
	err := Wrapf(New("a"), "b")
	if !strings.Contains(err.Error(), "a") || !strings.Contains(err.Error(), "b") {
		t.Fatalf("Error() = %q, want both a and b", err.Error())
	}
}

func TestErrorsIsPassesThrough(t *testing.T) {
	sentinel := errors.New("sentinel")
	if !errors.Is(Wrapf(sentinel, "ctx"), sentinel) {
		t.Fatal("errors.Is should see wrapped sentinel")
	}
}

func TestGetCodeFromWrapCode(t *testing.T) {
	c := NewCode(40001, 404)
	if got := GetCode(WrapCode(c, errors.New("x"))); got.Code() != 40001 {
		t.Fatalf("GetCode = %d, want 40001", got.Code())
	}
}

func TestJoinHasInnerCode(t *testing.T) {
	c := NewCode(40002, 400)
	err := Join(errors.New("plain"), NewCodeErr(c, "coded"))
	if !HasCode(err, c) {
		t.Fatal("HasCode should find inner code after Join")
	}
}

func TestStackNotEmpty(t *testing.T) {
	if Stack(New("boom")) == "" {
		t.Fatal("Stack should not be empty")
	}
}

func TestGetCodeThroughStdWrap(t *testing.T) {
	c := NewCode(40401, 404)
	err := fmt.Errorf("svc: %w", NewCodeErr(c, "missing"))
	if got := GetCode(err); got.Code() != 40401 {
		t.Fatalf("GetCode through fmt %%w = %d, want 40401", got.Code())
	}
}

func TestGetCodeThroughWrapOfStdWrap(t *testing.T) {
	c := NewCode(40401, 404)
	err := Wrap(fmt.Errorf("svc: %w", NewCodeErr(c, "missing")), "handler")
	if got := GetCode(err); got.Code() != 40401 {
		t.Fatalf("GetCode = %d, want 40401", got.Code())
	}
}

func TestHasCodeThroughStdWrap(t *testing.T) {
	c := NewCode(40402, 404)
	if !HasCode(fmt.Errorf("x: %w", NewCodeErr(c)), c) {
		t.Fatal("HasCode should see code through fmt %w")
	}
}

func TestGetCodeOuterCodeWins(t *testing.T) {
	inner, outer := NewCode(1, 400), NewCode(2, 409)
	err := WrapCode(outer, fmt.Errorf("x: %w", NewCodeErr(inner)))
	if got := GetCode(err); got.Code() != 2 {
		t.Fatalf("GetCode = %d, want outer 2", got.Code())
	}
}
