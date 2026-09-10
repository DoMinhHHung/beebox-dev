package apperror

import (
	"errors"
	"testing"
)

func TestErrorCodeAndCause(t *testing.T) {
	cause := errors.New("cause")
	err := Wrap(CodeValidation, "invalid input", cause)

	if !IsCode(err, CodeValidation) {
		t.Fatal("expected validation code")
	}
	if !errors.Is(err, cause) {
		t.Fatal("expected cause to be preserved")
	}
	if got := CodeOf(err); got != CodeValidation {
		t.Fatalf("expected code %q, got %q", CodeValidation, got)
	}
}

func TestCodeOfUnknownError(t *testing.T) {
	if got := CodeOf(errors.New("unknown")); got != CodeInternal {
		t.Fatalf("expected code %q, got %q", CodeInternal, got)
	}
}
