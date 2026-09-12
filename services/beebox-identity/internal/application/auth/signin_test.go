package auth

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
)

type fakeSignInUserRepository struct {
	findUser  user.User
	findErr   error
	findCalls int
}

func (f *fakeSignInUserRepository) Create(context.Context, user.User) error {
	return nil
}

func (f *fakeSignInUserRepository) FindByIdentifier(context.Context, identity.Identifier) (user.User, error) {
	f.findCalls++
	return f.findUser, f.findErr
}

type fakeSignInCredentialRepository struct {
	findCredential credential.Credential
	findErr        error
	findCalls      int
}

func (f *fakeSignInCredentialRepository) Create(context.Context, credential.Credential) error {
	return nil
}

func (f *fakeSignInCredentialRepository) FindByUserID(context.Context, identity.Identifier) (credential.Credential, error) {
	f.findCalls++
	return f.findCredential, f.findErr
}

func (f *fakeSignInCredentialRepository) Update(context.Context, credential.Credential) error {
	return nil
}

type fakeSignInPasswordHasher struct {
	verifyErr       error
	verifyCalls     int
	verifyPlaintext string
	verifyHash      string
}

func (f *fakeSignInPasswordHasher) Hash(context.Context, string) (string, error) {
	return "", nil
}

func (f *fakeSignInPasswordHasher) Verify(_ context.Context, plaintext, passwordHash string) error {
	f.verifyCalls++
	f.verifyPlaintext = plaintext
	f.verifyHash = passwordHash
	return f.verifyErr
}

type fakeSignInSessionRepository struct {
	createErr   error
	createCalls int
	created     session.Session
}

func (f *fakeSignInSessionRepository) Create(_ context.Context, value session.Session) error {
	f.createCalls++
	f.created = value
	return f.createErr
}

func (f *fakeSignInSessionRepository) FindByID(context.Context, string) (session.Session, error) {
	return session.Session{}, ErrNotFound
}

func (f *fakeSignInSessionRepository) Revoke(context.Context, session.Session) error {
	return nil
}

func (f *fakeSignInSessionRepository) RevokeAllByUserID(context.Context, identity.Identifier, time.Time) error {
	return nil
}

var _ UserRepository = (*fakeSignInUserRepository)(nil)
var _ CredentialRepository = (*fakeSignInCredentialRepository)(nil)
var _ PasswordHasher = (*fakeSignInPasswordHasher)(nil)
var _ SessionRepository = (*fakeSignInSessionRepository)(nil)

func signInUserForTest(t *testing.T) user.User {
	t.Helper()
	identifier, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	value, err := user.New(identifier, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected user error: %v", err)
	}
	return value
}

func signInCredentialForTest(t *testing.T) credential.Credential {
	t.Helper()
	identifier, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	value, err := credential.NewPassword(identifier, "stored-hash", time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}
	return value
}

func TestSignInSuccess(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findCredential: signInCredentialForTest(t)}
	sessions := &fakeSignInSessionRepository{}
	hasher := &fakeSignInPasswordHasher{}
	clock := &fakeClock{now: now}
	service := NewSignInService(users, credentials, sessions, hasher, clock)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "plain-secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID.String() != "user-1" {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if result.SessionID == "" || len(result.SessionID) != 64 {
		t.Fatalf("expected opaque session token, got %q", result.SessionID)
	}
	if !result.ExpiresAt.Equal(now.Add(24 * time.Hour)) {
		t.Fatalf("unexpected expires at: %v", result.ExpiresAt)
	}
	if sessions.createCalls != 1 {
		t.Fatalf("expected session create, got %d", sessions.createCalls)
	}
	if sessions.created.ID() == result.SessionID {
		t.Fatal("plaintext session token must not be persisted as session id")
	}
	sum := sha256.Sum256([]byte(result.SessionID))
	expected := hex.EncodeToString(sum[:])
	if sessions.created.ID() != expected {
		t.Fatalf("expected stored session id to be token hash")
	}
	if sessions.created.UserID().String() != "user-1" {
		t.Fatal("unexpected session user id")
	}
	if !sessions.created.IsActive(now) {
		t.Fatal("expected active session")
	}
	if hasher.verifyCalls != 1 || hasher.verifyPlaintext != "plain-secret" {
		t.Fatal("expected password verification")
	}
}

func TestSignInInvalidCredentialsUserMissing(t *testing.T) {
	users := &fakeSignInUserRepository{findErr: ErrNotFound}
	service := NewSignInService(users, &fakeSignInCredentialRepository{}, &fakeSignInSessionRepository{}, &fakeSignInPasswordHasher{}, &fakeClock{now: time.Now().UTC()})
	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "missing", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestSignInInvalidCredentialsPassword(t *testing.T) {
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findCredential: signInCredentialForTest(t)}
	hasher := &fakeSignInPasswordHasher{verifyErr: errors.New("mismatch")}
	service := NewSignInService(users, credentials, &fakeSignInSessionRepository{}, hasher, &fakeClock{now: time.Now().UTC()})
	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "wrong"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestSignInValidation(t *testing.T) {
	service := NewSignInService(&fakeSignInUserRepository{}, &fakeSignInCredentialRepository{}, &fakeSignInSessionRepository{}, &fakeSignInPasswordHasher{}, &fakeClock{now: time.Now().UTC()})
	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: ""})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}

func TestSignInSessionRepositoryFailure(t *testing.T) {
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findCredential: signInCredentialForTest(t)}
	sessions := &fakeSignInSessionRepository{createErr: errors.New("db")}
	service := NewSignInService(users, credentials, sessions, &fakeSignInPasswordHasher{}, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestSignInDoesNotCreateSessionOnAuthFailure(t *testing.T) {
	users := &fakeSignInUserRepository{findErr: ErrNotFound}
	sessions := &fakeSignInSessionRepository{}
	service := NewSignInService(users, &fakeSignInCredentialRepository{}, sessions, &fakeSignInPasswordHasher{}, &fakeClock{now: time.Now().UTC()})
	_, _ = service.SignIn(context.Background(), SignInInput{Identifier: "x", Password: "y"})
	if sessions.createCalls != 0 {
		t.Fatal("must not create session on auth failure")
	}
}

func TestGenerateSessionTokenUsesCryptoRand(t *testing.T) {
	fixed := bytes.Repeat([]byte{0xcd}, sessionTokenBytes)
	token, err := generateSessionToken(bytes.NewReader(fixed))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != hex.EncodeToString(fixed) {
		t.Fatal("unexpected token encoding")
	}
}

func TestGenerateSessionTokenRejectsBadReader(t *testing.T) {
	_, err := generateSessionToken(io.LimitReader(bytes.NewReader(nil), 0))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestHashSessionToken(t *testing.T) {
	h1 := hashSessionToken("token-a")
	h2 := hashSessionToken("token-a")
	h3 := hashSessionToken("token-b")
	if h1 != h2 || h1 == h3 || h1 == "token-a" {
		t.Fatal("unexpected hash behavior")
	}
}

func TestSignInRejectsRevokedCredential(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	cred := signInCredentialForTest(t)
	revoked, err := cred.Revoke(now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("revoke: %v", err)
	}
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findCredential: revoked}
	sessions := &fakeSignInSessionRepository{}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, sessions, hasher, &fakeClock{now: now})

	_, err = service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "plain-secret"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if sessions.createCalls != 0 {
		t.Fatal("must not create session for revoked credential")
	}
	if hasher.verifyCalls != 0 {
		t.Fatal("must not verify password for revoked credential")
	}
}
