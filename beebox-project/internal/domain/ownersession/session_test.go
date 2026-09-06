package ownersession

import (
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func TestNewSession_Valid(t *testing.T) {
	createdAt := time.Now()
	ttl := time.Hour

	s, token, err := NewSession("owner-1", ttl, createdAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty raw token")
	}
	if s.TokenHash == token {
		t.Fatal("expected TokenHash to differ from raw token")
	}
	if !s.ExpiresAt.Equal(createdAt.Add(ttl)) {
		t.Fatalf("expected ExpiresAt %v, got %v", createdAt.Add(ttl), s.ExpiresAt)
	}
}

func TestNewSession_EmptyOwnerID(t *testing.T) {
	_, _, err := NewSession("", time.Hour, time.Now())
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestNewSession_InvalidTTL(t *testing.T) {
	_, _, err := NewSession("owner-1", 0, time.Now())
	assertCode(t, err, apperror.CodeInvalidInput)
}

func TestVerify_Valid(t *testing.T) {
	createdAt := time.Now()
	s, _, err := NewSession("owner-1", time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error building session: %v", err)
	}
	if err := s.Verify(createdAt.Add(30 * time.Minute)); err != nil {
		t.Fatalf("expected valid session, got %v", err)
	}
}

func TestVerify_Expired(t *testing.T) {
	createdAt := time.Now()
	s, _, err := NewSession("owner-1", time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error building session: %v", err)
	}
	err = s.Verify(createdAt.Add(2 * time.Hour))
	assertCode(t, err, apperror.CodeCredentialInvalid)
}

func TestVerify_Revoked(t *testing.T) {
	createdAt := time.Now()
	s, _, err := NewSession("owner-1", time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error building session: %v", err)
	}
	revoked, err := s.Revoke(createdAt.Add(10 * time.Minute))
	if err != nil {
		t.Fatalf("unexpected error revoking session: %v", err)
	}
	err = revoked.Verify(createdAt.Add(20 * time.Minute))
	assertCode(t, err, apperror.CodeCredentialRevoked)
}

func TestRevoke_Valid(t *testing.T) {
	createdAt := time.Now()
	s, _, err := NewSession("owner-1", time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error building session: %v", err)
	}
	revoked, err := s.Revoke(createdAt.Add(5 * time.Minute))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}
}

func TestRevoke_AlreadyRevoked(t *testing.T) {
	createdAt := time.Now()
	s, _, err := NewSession("owner-1", time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error building session: %v", err)
	}
	revoked, err := s.Revoke(createdAt.Add(5 * time.Minute))
	if err != nil {
		t.Fatalf("unexpected error revoking session: %v", err)
	}
	_, err = revoked.Revoke(createdAt.Add(10 * time.Minute))
	assertCode(t, err, apperror.CodeCredentialRevoked)
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
