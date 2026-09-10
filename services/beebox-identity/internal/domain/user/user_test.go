package user

import (
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

func TestNewUser(t *testing.T) {
	id, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

	user, err := New(id, createdAt)
	if err != nil {
		t.Fatalf("unexpected user error: %v", err)
	}
	if user.ID() != id {
		t.Fatalf("expected user ID %q, got %q", id, user.ID())
	}
	if !user.CreatedAt().Equal(createdAt) {
		t.Fatalf("expected creation time %v, got %v", createdAt, user.CreatedAt())
	}
}

func TestNewUserRejectsMissingValues(t *testing.T) {
	id, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}

	_, err = New(id, time.Time{})
	if !errors.Is(err, domain.ErrInvalidUser) {
		t.Fatalf("expected invalid user error, got %v", err)
	}
}
