package verification

import (
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

func TestNewVerification(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(15 * time.Minute)

	v, err := New("ver-1", userID, TypeEmail, "user@example.com", "hash-value", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if v.ID() != "ver-1" {
		t.Fatalf("unexpected id: %s", v.ID())
	}
	if v.UserID() != userID {
		t.Fatalf("unexpected user id")
	}
	if v.Type() != TypeEmail {
		t.Fatalf("unexpected type")
	}
	if v.Target() != "user@example.com" {
		t.Fatalf("unexpected target")
	}
	if v.CodeHash() != "hash-value" {
		t.Fatalf("unexpected code hash")
	}
	if !v.IsPending(createdAt.Add(time.Minute)) {
		t.Fatal("expected pending verification")
	}
	if v.IsUsed() {
		t.Fatal("expected not used")
	}
	if v.IsExpired(createdAt.Add(time.Minute)) {
		t.Fatal("expected not expired")
	}
}

func TestNewVerificationRejectsInvalid(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(15 * time.Minute)

	cases := []struct {
		name      string
		id        string
		userID    identity.Identifier
		vtype     Type
		target    string
		codeHash  string
		createdAt time.Time
		expiresAt time.Time
	}{
		{"empty id", "", userID, TypeEmail, "a@b.com", "hash", createdAt, expiresAt},
		{"zero user", "ver-1", identity.Identifier(""), TypeEmail, "a@b.com", "hash", createdAt, expiresAt},
		{"invalid type", "ver-1", userID, Type("oauth"), "a@b.com", "hash", createdAt, expiresAt},
		{"empty target", "ver-1", userID, TypeEmail, "", "hash", createdAt, expiresAt},
		{"empty hash", "ver-1", userID, TypeEmail, "a@b.com", "", createdAt, expiresAt},
		{"zero created", "ver-1", userID, TypeEmail, "a@b.com", "hash", time.Time{}, expiresAt},
		{"zero expires", "ver-1", userID, TypeEmail, "a@b.com", "hash", createdAt, time.Time{}},
		{"expires not after created", "ver-1", userID, TypeEmail, "a@b.com", "hash", createdAt, createdAt},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := New(tc.id, tc.userID, tc.vtype, tc.target, tc.codeHash, tc.createdAt, tc.expiresAt)
			if !errors.Is(err, domain.ErrInvalidVerification) {
				t.Fatalf("expected invalid verification, got %v", err)
			}
		})
	}
}

func TestTypeValid(t *testing.T) {
	if !TypeEmail.Valid() || !TypePhone.Valid() {
		t.Fatal("expected email and phone to be valid")
	}
	if Type("").Valid() || Type("oauth").Valid() {
		t.Fatal("expected unsupported types to be invalid")
	}
}

func TestVerificationExpired(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(15 * time.Minute)
	v, err := New("ver-1", userID, TypePhone, "+15551212", "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}
	if v.IsExpired(expiresAt.Add(-time.Nanosecond)) {
		t.Fatal("expected not expired before boundary")
	}
	if !v.IsExpired(expiresAt) {
		t.Fatal("expected expired at boundary")
	}
	if v.IsPending(expiresAt) {
		t.Fatal("expected not pending when expired")
	}
}

func TestVerificationConsume(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(15 * time.Minute)
	v, err := New("ver-1", userID, TypeEmail, "user@example.com", "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}

	consumed, err := v.Consume(createdAt.Add(time.Minute))
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
	if !errors.Is(err, domain.ErrVerificationUsed) {
		t.Fatalf("expected used error, got %v", err)
	}
}

func TestVerificationConsumeExpired(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(15 * time.Minute)
	v, err := New("ver-1", userID, TypeEmail, "user@example.com", "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}

	_, err = v.Consume(expiresAt)
	if !errors.Is(err, domain.ErrVerificationExpired) {
		t.Fatalf("expected expired error, got %v", err)
	}
}

func TestVerificationConsumeInvalidTime(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(15 * time.Minute)
	v, err := New("ver-1", userID, TypeEmail, "user@example.com", "hash", createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected verification error: %v", err)
	}

	_, err = v.Consume(time.Time{})
	if !errors.Is(err, domain.ErrInvalidVerification) {
		t.Fatalf("expected invalid verification, got %v", err)
	}
	_, err = v.Consume(createdAt.Add(-time.Second))
	if !errors.Is(err, domain.ErrInvalidVerification) {
		t.Fatalf("expected invalid verification, got %v", err)
	}
}