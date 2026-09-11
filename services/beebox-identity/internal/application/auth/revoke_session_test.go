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

type fakeRevokeSessionRepository struct {
	findSession session.Session
	findErr     error
	findCalls   int
	findID      string
	revokeErr   error
	revokeCalls int
	revokedArg  session.Session
}

func (f *fakeRevokeSessionRepository) Create(context.Context, session.Session) error {
	return nil
}

func (f *fakeRevokeSessionRepository) FindByID(_ context.Context, id string) (session.Session, error) {
	f.findCalls++
	f.findID = id
	return f.findSession, f.findErr
}

func (f *fakeRevokeSessionRepository) Revoke(_ context.Context, value session.Session) error {
	f.revokeCalls++
	f.revokedArg = value
	return f.revokeErr
}

var _ SessionRepository = (*fakeRevokeSessionRepository)(nil)

type fakeRevokeSessionClock struct {
	now      time.Time
	nowCalls int
}

func (f *fakeRevokeSessionClock) Now() time.Time {
	f.nowCalls++
	return f.now
}

var _ Clock = (*fakeRevokeSessionClock)(nil)

func revokeSessionForTest(t *testing.T, token string) session.Session {
	t.Helper()
	identifier, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	value, err := session.New(hashSessionToken(token), identifier, createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected session error: %v", err)
	}
	return value
}

func TestRevokeSessionSuccess(t *testing.T) {
	token := "client-session-token"
	fixture := revokeSessionForTest(t, token)
	now := fixture.CreatedAt().Add(30 * time.Minute)
	sessions := &fakeRevokeSessionRepository{findSession: fixture}
	clock := &fakeRevokeSessionClock{now: now}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: token})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.findCalls != 1 {
		t.Fatalf("expected one lookup, got %d", sessions.findCalls)
	}
	if sessions.findID != hashSessionToken(token) {
		t.Fatal("expected lookup by hashed token")
	}
	if sessions.revokeCalls != 1 {
		t.Fatalf("expected one revoke, got %d", sessions.revokeCalls)
	}
	if !sessions.revokedArg.IsRevoked() {
		t.Fatal("expected revoked session")
	}
}

func TestRevokeSessionValidation(t *testing.T) {
	service := NewRevokeSessionService(&fakeRevokeSessionRepository{}, &fakeRevokeSessionClock{now: time.Now().UTC()})
	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: ""})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}

func TestRevokeSessionNotFound(t *testing.T) {
	sessions := &fakeRevokeSessionRepository{findErr: ErrNotFound}
	service := NewRevokeSessionService(sessions, &fakeRevokeSessionClock{now: time.Now().UTC()})
	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "missing"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestRevokeSessionAlreadyRevoked(t *testing.T) {
	token := "token"
	fixture := revokeSessionForTest(t, token)
	revoked, err := fixture.Revoke(fixture.CreatedAt().Add(time.Minute))
	if err != nil {
		t.Fatalf("revoke fixture: %v", err)
	}
	sessions := &fakeRevokeSessionRepository{findSession: revoked}
	service := NewRevokeSessionService(sessions, &fakeRevokeSessionClock{now: fixture.CreatedAt().Add(2 * time.Minute)})
	err = service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: token})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestRevokeSessionRepositoryFailure(t *testing.T) {
	sessions := &fakeRevokeSessionRepository{findErr: errors.New("db")}
	service := NewRevokeSessionService(sessions, &fakeRevokeSessionClock{now: time.Now().UTC()})
	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "token"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}
