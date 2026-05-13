package errors

import (
	stderrors "errors"
	"strings"
	"testing"
)

func TestError_FormatsWithCause(t *testing.T) {
	e := &Error{
		Component: "lock",
		Problem:   "cannot acquire lock",
		Cause:     "flock busy",
	}
	got := e.Error()
	want := "[lock] cannot acquire lock: flock busy"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_FormatsWithoutCause(t *testing.T) {
	e := &Error{
		Component: "config",
		Problem:   "samuel.toml missing",
	}
	got := e.Error()
	want := "[config] samuel.toml missing"
	if got != want {
		t.Errorf("Error() = %q, want %q", got, want)
	}
}

func TestError_NilSafe(t *testing.T) {
	var e *Error
	if got := e.Error(); got != "" {
		t.Errorf("nil Error() = %q, want empty", got)
	}
	if got := e.Unwrap(); got != nil {
		t.Errorf("nil Unwrap() = %v, want nil", got)
	}
	if got := e.Wrap(stderrors.New("x")); got != nil {
		t.Errorf("nil Wrap() = %v, want nil", got)
	}
}

func TestError_Wrap_PopulatesCauseFromErr(t *testing.T) {
	base := &Error{Component: "samuel-skills", Problem: "symlink failed"}
	inner := stderrors.New("permission denied")
	wrapped := base.Wrap(inner)

	if wrapped.Cause != "permission denied" {
		t.Errorf("Wrap should populate Cause from err; got %q", wrapped.Cause)
	}
	if !stderrors.Is(wrapped, inner) {
		t.Errorf("errors.Is should traverse Wrap chain")
	}
}

func TestError_Wrap_PreservesExplicitCause(t *testing.T) {
	base := &Error{
		Component: "lock",
		Problem:   "register failed",
		Cause:     "explicit cause",
	}
	inner := stderrors.New("inner cause")
	wrapped := base.Wrap(inner)

	if wrapped.Cause != "explicit cause" {
		t.Errorf("Wrap should not overwrite explicit Cause; got %q", wrapped.Cause)
	}
	if !stderrors.Is(wrapped, inner) {
		t.Errorf("Wrap should still preserve underlying err for errors.Is")
	}
}

func TestError_Wrap_DoesNotMutateReceiver(t *testing.T) {
	base := &Error{Component: "plugin", Problem: "x"}
	_ = base.Wrap(stderrors.New("inner"))
	if base.wrapped != nil || base.Cause != "" {
		t.Errorf("Wrap mutated receiver: wrapped=%v cause=%q", base.wrapped, base.Cause)
	}
}

func TestIsRecoverable_TrueForRecoverableErrors(t *testing.T) {
	e := &Error{Component: "x", Problem: "y", Recoverable: true}
	if !IsRecoverable(e) {
		t.Errorf("IsRecoverable should return true for Recoverable=true")
	}
}

func TestIsRecoverable_FalseForBareError(t *testing.T) {
	if IsRecoverable(stderrors.New("plain")) {
		t.Errorf("IsRecoverable should return false for non-*Error")
	}
}

func TestIsRecoverable_TraversesJoined(t *testing.T) {
	inner := &Error{Component: "x", Problem: "y", Recoverable: true}
	wrapper := stderrors.New("wrapper")
	combined := stderrors.Join(wrapper, inner)
	if !IsRecoverable(combined) {
		t.Errorf("IsRecoverable should traverse joined errors")
	}
}

func TestError_AsAcrossWrap(t *testing.T) {
	base := &Error{Component: "samuel-skills", Problem: "sync"}
	wrapped := base.Wrap(stderrors.New("inner"))
	var target *Error
	if !stderrors.As(wrapped, &target) {
		t.Fatalf("errors.As failed to extract *Error")
	}
	if !strings.Contains(target.Error(), "samuel-skills") {
		t.Errorf("extracted Error did not preserve Component")
	}
}
