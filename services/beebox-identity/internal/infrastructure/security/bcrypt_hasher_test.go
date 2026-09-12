package security

import (
	"context"
	"strings"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
)

func TestHashAccepts72BytePassword(t *testing.T) {
	h := NewBcryptPasswordHasher()
	password := strings.Repeat("a", 72)
	hash, err := h.Hash(context.Background(), password)
	if err != nil {
		t.Fatalf("expected 72-byte password to hash, got %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
}

func TestHashRejects73BytePassword(t *testing.T) {
	h := NewBcryptPasswordHasher()
	password := strings.Repeat("a", 73)
	_, err := h.Hash(context.Background(), password)
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
}
