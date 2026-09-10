package apperror_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
)

func TestNew_SetsCodeAndMessage(t *testing.T) {
	err := apperror.New(apperror.CodeValidation, "invalid input")

	if err.Code != apperror.CodeValidation {
		t.Fatalf("expected code %q, got %q", apperror.CodeValidation, err.Code)
	}
	if err.Error() != "invalid input" {
		t.Fatalf("expected message %q, got %q", "invalid input", err.Error())
	}
	if err.Cause != nil {
		t.Fatalf("expected nil cause, got %v", err.Cause)
	}
}

func TestWrap_PreservesCauseForUnwrap(t *testing.T) {
	cause := errors.New("connection refused")
	err := apperror.Wrap(apperror.CodeDependencyFailure, "database unavailable", cause)

	if !errors.Is(err, cause) {
		t.Fatalf("expected errors.Is to find wrapped cause")
	}
	if err.Error() != "database unavailable" {
		t.Fatalf("expected message to hide cause, got %q", err.Error())
	}
}

func TestIsCode_MatchesWrappedError(t *testing.T) {
	err := apperror.New(apperror.CodeNotFound, "project not found")

	if !apperror.IsCode(err, apperror.CodeNotFound) {
		t.Fatalf("expected IsCode to match NOT_FOUND")
	}
	if apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected IsCode to reject CONFLICT")
	}
}

func TestIsCode_ReturnsFalseForNonAppError(t *testing.T) {
	err := errors.New("plain error")

	if apperror.IsCode(err, apperror.CodeInternal) {
		t.Fatalf("expected IsCode to reject a non-apperror error")
	}
}

func TestCodeOf_DefaultsToInternalForUnknownError(t *testing.T) {
	err := errors.New("unexpected panic recovery")

	if apperror.CodeOf(err) != apperror.CodeInternal {
		t.Fatalf("expected CodeOf to default to INTERNAL, got %q", apperror.CodeOf(err))
	}
}

func TestCodeOf_ReturnsActualCodeForAppError(t *testing.T) {
	err := apperror.New(apperror.CodeForbidden, "owner only")

	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("expected CodeOf to return FORBIDDEN, got %q", apperror.CodeOf(err))
	}
}
