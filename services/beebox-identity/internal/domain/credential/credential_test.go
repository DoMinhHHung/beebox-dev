package credential

import (
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

func TestNewPassword(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	password, err := NewPassword(userID, "stored-password-hash", createdAt)
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}
	if password.UserID() != userID {
		t.Fatalf("expected user ID %q, got %q", userID, password.UserID())
	}
	if password.PasswordHash() != "stored-password-hash" {
		t.Fatalf("expected stored hash to be retained")
	}
	if password.IsRevoked() {
		t.Fatal("expected new password to be active")
	}
}

func TestNewPasswordRejectsMissingState(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}

	_, err = NewPassword(userID, "", time.Now())
	if !errors.Is(err, domain.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential error, got %v", err)
	}
}

func TestPasswordRevoke(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	password, err := NewPassword(userID, "stored-password-hash", createdAt)
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}

	revoked, err := password.Revoke(createdAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected revoke error: %v", err)
	}
	if !revoked.IsRevoked() {
		t.Fatal("expected password to be revoked")
	}
	if password.IsRevoked() {
		t.Fatal("expected original password value to remain unchanged")
	}
}

func TestChangePassword(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	password, err := NewPassword(userID, "old-hash", createdAt)
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}

	changed, err := password.ChangePassword("new-hash")
	if err != nil {
		t.Fatalf("unexpected change password error: %v", err)
	}
	if changed.PasswordHash() != "new-hash" {
		t.Fatalf("expected new hash, got %s", changed.PasswordHash())
	}
	if password.PasswordHash() != "old-hash" {
		t.Fatal("expected original credential to remain unchanged")
	}
	if changed.UserID() != userID {
		t.Fatal("expected user id to remain")
	}
}

func TestChangePasswordRejectsEmptyHash(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	password, err := NewPassword(userID, "old-hash", time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}
	_, err = password.ChangePassword("")
	if !errors.Is(err, domain.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential, got %v", err)
	}
}

func TestChangePasswordRejectsRevoked(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	password, err := NewPassword(userID, "old-hash", createdAt)
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}
	revoked, err := password.Revoke(createdAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected revoke error: %v", err)
	}
	_, err = revoked.ChangePassword("new-hash")
	if !errors.Is(err, domain.ErrInvalidCredential) {
		t.Fatalf("expected invalid credential, got %v", err)
	}
}
