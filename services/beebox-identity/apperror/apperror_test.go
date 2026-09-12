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

func TestIsCodeAndCodeOfTypedNil(t *testing.T) {
	var typedNil *Error
	err := error(typedNil)

	if IsCode(err, CodeValidation) {
		t.Fatal("typed nil *Error must not match any code")
	}
	if got := CodeOf(err); got != CodeInternal {
		t.Fatalf("expected CodeInternal for typed nil, got %q", got)
	}
}

func TestIsCodeFalseForNilError(t *testing.T) {
	if IsCode(nil, CodeInternal) {
		t.Fatal("nil error must not match")
	}
	if got := CodeOf(nil); got != CodeInternal {
		t.Fatalf("expected CodeInternal for nil, got %q", got)
	}
}
