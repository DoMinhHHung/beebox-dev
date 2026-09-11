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
	revokeErr   error
	revokeCalls int
	revokedArg  session.Session
}

func (f *fakeRevokeSessionRepository) Create(context.Context, session.Session) error {
	return nil
}

func (f *fakeRevokeSessionRepository) FindByID(context.Context, string) (session.Session, error) {
	f.findCalls++
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

func revokeSessionForTest(t *testing.T) session.Session {
	t.Helper()
	identifier, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	createdAt := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
	expiresAt := createdAt.Add(time.Hour)
	value, err := session.New("session-1", identifier, createdAt, expiresAt)
	if err != nil {
		t.Fatalf("unexpected session error: %v", err)
	}
	return value
}

func TestRevokeSessionSuccess(t *testing.T) {
	fixture := revokeSessionForTest(t)
	now := fixture.CreatedAt().Add(30 * time.Minute)
	sessions := &fakeRevokeSessionRepository{findSession: fixture}
	clock := &fakeRevokeSessionClock{now: now}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "session-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sessions.findCalls != 1 {
		t.Fatalf("expected one lookup, got %d", sessions.findCalls)
	}
	if sessions.revokeCalls != 1 {
		t.Fatalf("expected one persistence call, got %d", sessions.revokeCalls)
	}
	if clock.nowCalls != 1 {
		t.Fatalf("expected clock to be consulted once, got %d", clock.nowCalls)
	}
}

func TestRevokeSessionRejectsEmptySessionID(t *testing.T) {
	sessions := &fakeRevokeSessionRepository{}
	clock := &fakeRevokeSessionClock{now: time.Date(2026, time.January, 1, 1, 0, 0, 0, time.UTC)}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: ""})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if sessions.findCalls != 0 || sessions.revokeCalls != 0 || clock.nowCalls != 0 {
		t.Fatal("expected no dependency interaction for invalid input")
	}
}

func TestRevokeSessionRejectsUnknownSession(t *testing.T) {
	sessions := &fakeRevokeSessionRepository{findErr: ErrNotFound}
	clock := &fakeRevokeSessionClock{now: time.Date(2026, time.January, 1, 1, 0, 0, 0, time.UTC)}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "unknown"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
	if sessions.revokeCalls != 0 {
		t.Fatal("expected no persistence call for an unknown session")
	}
	if clock.nowCalls != 0 {
		t.Fatal("expected no clock interaction for an unknown session")
	}
}

func TestRevokeSessionMapsFindByIDFailure(t *testing.T) {
	sessions := &fakeRevokeSessionRepository{findErr: errors.New("repository unavailable")}
	clock := &fakeRevokeSessionClock{now: time.Date(2026, time.January, 1, 1, 0, 0, 0, time.UTC)}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "session-1"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if sessions.revokeCalls != 0 {
		t.Fatal("expected no persistence call after a lookup failure")
	}
	if clock.nowCalls != 0 {
		t.Fatal("expected no clock interaction after a lookup failure")
	}
}

func TestRevokeSessionRejectsAlreadyRevokedSession(t *testing.T) {
	fixture := revokeSessionForTest(t)
	alreadyRevoked, err := fixture.Revoke(fixture.CreatedAt().Add(10 * time.Minute))
	if err != nil {
		t.Fatalf("unexpected error preparing fixture: %v", err)
	}

	sessions := &fakeRevokeSessionRepository{findSession: alreadyRevoked}
	clock := &fakeRevokeSessionClock{now: fixture.CreatedAt().Add(30 * time.Minute)}
	service := NewRevokeSessionService(sessions, clock)

	revokeErr := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "session-1"})
	if !apperror.IsCode(revokeErr, apperror.CodeConflict) {
		t.Fatalf("expected conflict error, got %v", revokeErr)
	}
	if sessions.revokeCalls != 0 {
		t.Fatal("expected no persistence call for an already revoked session")
	}
}

func TestRevokeSessionMapsDomainRejection(t *testing.T) {
	fixture := revokeSessionForTest(t)
	sessions := &fakeRevokeSessionRepository{findSession: fixture}
	clock := &fakeRevokeSessionClock{now: fixture.CreatedAt().Add(-time.Minute)}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "session-1"})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if sessions.revokeCalls != 0 {
		t.Fatal("expected no persistence call for a domain-rejected revoke")
	}
}

func TestRevokeSessionMapsPersistenceFailure(t *testing.T) {
	fixture := revokeSessionForTest(t)
	sessions := &fakeRevokeSessionRepository{
		findSession: fixture,
		revokeErr:   errors.New("repository unavailable"),
	}
	clock := &fakeRevokeSessionClock{now: fixture.CreatedAt().Add(30 * time.Minute)}
	service := NewRevokeSessionService(sessions, clock)

	err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "session-1"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if sessions.revokeCalls != 1 {
		t.Fatalf("expected one attempted persistence call, got %d", sessions.revokeCalls)
	}
}

func TestRevokeSessionPersistsRevokedSession(t *testing.T) {
	fixture := revokeSessionForTest(t)
	now := fixture.CreatedAt().Add(30 * time.Minute)
	sessions := &fakeRevokeSessionRepository{findSession: fixture}
	clock := &fakeRevokeSessionClock{now: now}
	service := NewRevokeSessionService(sessions, clock)

	if err := service.RevokeSession(context.Background(), RevokeSessionInput{SessionID: "session-1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if sessions.revokedArg.ID() != fixture.ID() {
		t.Fatalf("expected persisted session ID %q, got %q", fixture.ID(), sessions.revokedArg.ID())
	}
	if !sessions.revokedArg.IsRevoked() {
		t.Fatal("expected persisted session to be revoked")
	}
	revokedAt, ok := sessions.revokedArg.RevokedAt()
	if !ok || !revokedAt.Equal(now) {
		t.Fatalf("expected persisted revokedAt %v, got %v (ok=%v)", now, revokedAt, ok)
	}
}