package passwordreset

import (
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

func TestNewPasswordReset(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)

	p, err := New("reset-1", userID, "token-hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.ID() != "reset-1" {
		t.Fatalf("unexpected id: %s", p.ID())
	}
	if p.UserID() != userID {
		t.Fatal("unexpected user id")
	}
	if p.TokenHash() != "token-hash" {
		t.Fatal("unexpected token hash")
	}
	if !p.IsPending(createdAt.Add(time.Minute)) {
		t.Fatal("expected pending")
	}
	if p.IsUsed() {
		t.Fatal("expected not used")
	}
	if p.IsExpired(createdAt.Add(time.Minute)) {
		t.Fatal("expected not expired")
	}
}

func TestNewPasswordResetRejectsInvalid(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)

	cases := []struct {
		name      string
		id        string
		userID    identity.Identifier
		tokenHash string
		createdAt time.Time
		expiresAt time.Time
	}{
		{"empty id", "", userID, "hash", createdAt, expiresAt},
		{"zero user", "reset-1", identity.Identifier(""), "hash", createdAt, expiresAt},
		{"empty hash", "reset-1", userID, "", createdAt, expiresAt},
		{"zero created", "reset-1", userID, "hash", time.Time{}, expiresAt},
		{"zero expires", "reset-1", userID, "hash", createdAt, time.Time{}},
		{"expires not after", "reset-1", userID, "hash", createdAt, createdAt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.id, tc.userID, tc.tokenHash, tc.createdAt, tc.expiresAt)
			if !errors.Is(err, domain.ErrInvalidPasswordReset) {
				t.Fatalf("expected invalid password reset, got %v", err)
			}
		})
	}
}

func TestPasswordResetExpired(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	p, err := New("reset-1", userID, "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if p.IsExpired(expiresAt.Add(-time.Nanosecond)) {
		t.Fatal("expected not expired before boundary")
	}
	if !p.IsExpired(expiresAt) {
		t.Fatal("expected expired at boundary")
	}
	if p.IsPending(expiresAt) {
		t.Fatal("expected not pending when expired")
	}
}

func TestPasswordResetConsume(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	p, err := New("reset-1", userID, "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	consumed, err := p.Consume(createdAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("unexpected consume error: %v", err)
	}
	if !consumed.IsUsed() {
		t.Fatal("expected used")
	}
	usedAt, ok := consumed.UsedAt()
	if !ok || !usedAt.Equal(createdAt.Add(time.Minute)) {
		t.Fatalf("unexpected used at: %v ok=%v", usedAt, ok)
	}
	if consumed.IsPending(createdAt.Add(2 * time.Minute)) {
		t.Fatal("expected not pending after use")
	}

	_, err = consumed.Consume(createdAt.Add(2 * time.Minute))
	if !errors.Is(err, domain.ErrPasswordResetUsed) {
		t.Fatalf("expected used error, got %v", err)
	}
}

func TestPasswordResetConsumeExpired(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	p, err := New("reset-1", userID, "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Consume(expiresAt)
	if !errors.Is(err, domain.ErrPasswordResetExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestPasswordResetConsumeInvalidTime(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	p, err := New("reset-1", userID, "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = p.Consume(time.Time{})
	if !errors.Is(err, domain.ErrInvalidPasswordReset) {
		t.Fatalf("expected invalid password reset, got %v", err)
	}
	_, err = p.Consume(createdAt.Add(-time.Second))
	if !errors.Is(err, domain.ErrInvalidPasswordReset) {
		t.Fatalf("expected invalid password reset, got %v", err)
	}
}
