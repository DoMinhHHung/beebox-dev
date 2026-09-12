package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
)

type fakeAuthenticateSessionRepository struct {
	findSession session.Session
	findErr     error
	findCalls   int
	findID      string
}

func (f *fakeAuthenticateSessionRepository) Create(context.Context, session.Session) error {
	return nil
}

func (f *fakeAuthenticateSessionRepository) FindByID(_ context.Context, id string) (session.Session, error) {
	f.findCalls++
	f.findID = id
	return f.findSession, f.findErr
}

func (f *fakeAuthenticateSessionRepository) Revoke(context.Context, session.Session) error {
	return nil
}

func (f *fakeAuthenticateSessionRepository) RevokeAllByUserID(context.Context, identity.Identifier, time.Time) error {
	return nil
}

var _ SessionRepository = (*fakeAuthenticateSessionRepository)(nil)

func activeSessionForAuthTest(t *testing.T, token string, now time.Time) session.Session {
	t.Helper()
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("identifier: %v", err)
	}
	s, err := session.New(hashSessionToken(token), userID, now.Add(-time.Hour), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	return s
}

func TestAuthenticateSessionSuccess(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "opaque-client-session-token-value-32b-hex-style-aaaa"
	sessions := &fakeAuthenticateSessionRepository{findSession: activeSessionForAuthTest(t, token, now)}
	service := NewAuthenticateSessionService(sessions, &fakeClock{now: now})

	result, err := service.AuthenticateSession(context.Background(), AuthenticateSessionInput{Token: token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID.String() != "user-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if result.SessionID != hashSessionToken(token) {
		t.Fatal("expected storage session id")
	}
	if sessions.findID != hashSessionToken(token) {
		t.Fatal("expected lookup by hashed token")
	}
}

func TestAuthenticateSessionEmptyToken(t *testing.T) {
	service := NewAuthenticateSessionService(&fakeAuthenticateSessionRepository{}, &fakeClock{now: time.Now().UTC()})
	_, err := service.AuthenticateSession(context.Background(), AuthenticateSessionInput{Token: ""})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateSessionUnknown(t *testing.T) {
	sessions := &fakeAuthenticateSessionRepository{findErr: ErrNotFound}
	service := NewAuthenticateSessionService(sessions, &fakeClock{now: time.Now().UTC()})
	_, err := service.AuthenticateSession(context.Background(), AuthenticateSessionInput{Token: "missing"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateSessionExpired(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token"
	userID, _ := identity.NewIdentifier("user-1")
	created := now.Add(-2 * time.Hour)
	expires := now.Add(-time.Hour)
	s, _ := session.New(hashSessionToken(token), userID, created, expires)
	sessions := &fakeAuthenticateSessionRepository{findSession: s}
	service := NewAuthenticateSessionService(sessions, &fakeClock{now: now})
	_, err := service.AuthenticateSession(context.Background(), AuthenticateSessionInput{Token: token})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateSessionRevoked(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token"
	s := activeSessionForAuthTest(t, token, now)
	revoked, err := s.Revoke(now.Add(-time.Minute))
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	sessions := &fakeAuthenticateSessionRepository{findSession: revoked}
	service := NewAuthenticateSessionService(sessions, &fakeClock{now: now})
	_, err = service.AuthenticateSession(context.Background(), AuthenticateSessionInput{Token: token})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestAuthenticateSessionRepositoryFailure(t *testing.T) {
	sessions := &fakeAuthenticateSessionRepository{findErr: errors.New("db")}
	service := NewAuthenticateSessionService(sessions, &fakeClock{now: time.Now().UTC()})
	_, err := service.AuthenticateSession(context.Background(), AuthenticateSessionInput{Token: "token"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}
