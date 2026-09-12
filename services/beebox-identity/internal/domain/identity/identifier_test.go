package identity

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
)

func TestNewIdentifier(t *testing.T) {
	id, err := NewIdentifier(" user-1 ")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id.String() != "user-1" {
		t.Fatalf("expected normalized identifier, got %q", id.String())
	}
}

func TestNewIdentifierRejectsBlankValue(t *testing.T) {
	_, err := NewIdentifier("  ")
	if !errors.Is(err, domain.ErrInvalidIdentityIdentifier) {
		t.Fatalf("expected invalid identifier error, got %v", err)
	}
}
