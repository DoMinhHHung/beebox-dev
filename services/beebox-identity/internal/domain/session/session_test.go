package session

import (
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

func TestNewSession(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)

	session, err := New("session-1", userID, createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected session error: %v", err)
	}
	if !session.IsActive(createdAt.Add(30 * time.Minute)) {
		t.Fatal("expected session to be active")
	}
	if session.IsExpired(expiresAt.Add(-time.Nanosecond)) {
		t.Fatal("expected session to remain active before expiration")
	}
	if !session.IsExpired(expiresAt) {
		t.Fatal("expected session to expire at expiration time")
	}
}

func TestNewSessionRejectsInvalidTimes(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	_, err = New("session-1", userID, createdAt, createdAt)
	if !errors.Is(err, domain.ErrInvalidSession) {
		t.Fatalf("expected invalid session error, got %v", err)
	}
}

func TestSessionRevoke(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	session, err := New("session-1", userID, createdAt, createdAt.Add(time.Hour))
	if err != nil {
		t.Fatalf("unexpected session error: %v", err)
	}

	revoked, err := session.Revoke(createdAt.Add(30 * time.Minute))
	if err != nil {
		t.Fatalf("unexpected revoke error: %v", err)
	}
	if !revoked.IsRevoked() {
		t.Fatal("expected session to be revoked")
	}
	if revoked.IsActive(createdAt.Add(45 * time.Minute)) {
		t.Fatal("expected revoked session to be inactive")
	}

	_, err = revoked.Revoke(createdAt.Add(45 * time.Minute))
	if !errors.Is(err, domain.ErrSessionRevoked) {
		t.Fatalf("expected already revoked error, got %v", err)
	}
}
