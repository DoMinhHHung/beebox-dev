package owner

import (
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func TestNewOwner_Valid(t *testing.T) {
	o, err := NewOwner("owner-1", "minh@example.com", "password123", time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if o.Email != "minh@example.com" {
		t.Fatalf("expected email to be set, got %q", o.Email)
	}
	if o.PasswordHash == "" || o.PasswordHash == "password123" {
		t.Fatal("expected password to be hashed, not stored in plaintext")
	}
}

func TestNewOwner_InvalidEmail(t *testing.T) {
	_, err := NewOwner("owner-1", "not-an-email", "password123", time.Now())
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestNewOwner_PasswordTooShort(t *testing.T) {
	_, err := NewOwner("owner-1", "minh@example.com", "short", time.Now())
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestVerifyPassword_Correct(t *testing.T) {
	o, err := NewOwner("owner-1", "minh@example.com", "password123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error building owner: %v", err)
	}
	if err := o.VerifyPassword("password123"); err != nil {
		t.Fatalf("expected correct password to verify, got %v", err)
	}
}

func TestVerifyPassword_Incorrect(t *testing.T) {
	o, err := NewOwner("owner-1", "minh@example.com", "password123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error building owner: %v", err)
	}
	err = o.VerifyPassword("wrong-password")
	assertCode(t, err, apperror.CodeCredentialInvalid)
}

func assertCode(t *testing.T, err error, expected apperror.Code) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apperror.CodeOf(err) != expected {
		t.Fatalf("expected %s, got %s", expected, apperror.CodeOf(err))
	}
}