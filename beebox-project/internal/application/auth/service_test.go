package auth

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
)

func newTestService() *Service {
	return NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
}

func TestSignUp_Valid(t *testing.T) {
	svc := newTestService()

	o, err := svc.SignUp(context.Background(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if o.ID == "" {
		t.Fatal("expected generated owner id")
	}
}

func TestSignUp_DuplicateEmail(t *testing.T) {
	svc := newTestService()

	if _, err := svc.SignUp(context.Background(), "minh@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error on first sign up: %v", err)
	}

	_, err := svc.SignUp(context.Background(), "minh@example.com", "anotherpass")
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %s", apperror.CodeOf(err))
	}
}

func TestSignIn_Valid_ThenVerifySessionReturnsOwnerID(t *testing.T) {
	svc := newTestService()

	o, err := svc.SignUp(context.Background(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error signing up: %v", err)
	}

	token, err := svc.SignIn(context.Background(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("expected no error signing in, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	ownerID, err := svc.VerifySession(context.Background(), token)
	if err != nil {
		t.Fatalf("expected valid session, got %v", err)
	}
	if ownerID != o.ID {
		t.Fatalf("expected owner id %q, got %q", o.ID, ownerID)
	}
}

func TestSignIn_WrongPassword(t *testing.T) {
	svc := newTestService()

	if _, err := svc.SignUp(context.Background(), "minh@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error signing up: %v", err)
	}

	_, err := svc.SignIn(context.Background(), "minh@example.com", "wrong-password")
	if apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected CodeCredentialInvalid, got %s", apperror.CodeOf(err))
	}
}

func TestSignIn_UnknownEmail(t *testing.T) {
	svc := newTestService()

	_, err := svc.SignIn(context.Background(), "ghost@example.com", "password123")
	if apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected CodeCredentialInvalid (not CodeNotFound, to avoid email enumeration), got %s", apperror.CodeOf(err))
	}
}

func TestVerifySession_InvalidToken(t *testing.T) {
	svc := newTestService()

	_, err := svc.VerifySession(context.Background(), "not-a-real-token")
	if apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected CodeCredentialInvalid, got %s", apperror.CodeOf(err))
	}
}

func TestSignOut_ThenVerifySessionFails(t *testing.T) {
	svc := newTestService()

	if _, err := svc.SignUp(context.Background(), "minh@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error signing up: %v", err)
	}

	token, err := svc.SignIn(context.Background(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error signing in: %v", err)
	}

	if err := svc.SignOut(context.Background(), token); err != nil {
		t.Fatalf("expected no error signing out, got %v", err)
	}

	_, err = svc.VerifySession(context.Background(), token)
	if err == nil {
		t.Fatal("expected error verifying session after sign out")
	}
}

func TestSignOut_UnknownToken_ReturnsNil(t *testing.T) {
	svc := newTestService()

	if err := svc.SignOut(context.Background(), "never-issued-token"); err != nil {
		t.Fatalf("expected no error for unknown token, got %v", err)
	}
}
