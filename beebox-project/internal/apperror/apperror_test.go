package apperror

import (
	"errors"
	"testing"
)

func TestError_MessageFormatting(t *testing.T) {
	err := New(CodeInvalidInput, "HTTP_PORT must be a valid port number")
	expected := "INVALID_INPUT: HTTP_PORT must be a valid port number"

	if err.Error() != expected {
		t.Fatalf("expected %q, got %q", expected, err.Error())
	}
}

func TestWrap_PreservesUnderlyingError(t *testing.T) {
	underlying := errors.New("strconv failure")
	err := Wrap(CodeInvalidInput, "HTTP_PORT parse failed", underlying)

	if !errors.Is(err, underlying) {
		t.Fatal("expected wrapped error to unwrap to underlying error")
	}
}

func TestCodeOf_ReturnsInternalForUnknownError(t *testing.T) {
	plain := errors.New("something else")

	if CodeOf(plain) != CodeInternal {
		t.Fatalf("expected CodeInternal for non-apperror error, got %s", CodeOf(plain))
	}
}

func TestCodeOf_ReturnsAppErrorCode(t *testing.T) {
	err := New(CodeConflict, "duplicate")

	if CodeOf(err) != CodeConflict {
		t.Fatalf("expected CodeConflict, got %s", CodeOf(err))
	}
}
