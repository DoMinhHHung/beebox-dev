package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
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

func (f *fakeSignInCredentialRepository) Update(context.Context, credential.Credential) error {
	return nil
}

func (f *fakeSignInCredentialRepository) FindByUserID(context.Context, identity.Identifier) (credential.Credential, error) {
	f.findCalls++
	return f.findCredential, f.findErr
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

var _ UserRepository = (*fakeSignInUserRepository)(nil)
var _ CredentialRepository = (*fakeSignInCredentialRepository)(nil)
var _ PasswordHasher = (*fakeSignInPasswordHasher)(nil)

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
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findCredential: signInCredentialForTest(t)}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "plain-secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID.String() != "user-1" {
		t.Fatalf("expected user ID user-1, got %q", result.UserID)
	}
	if users.findCalls != 1 {
		t.Fatalf("expected one user lookup, got %d", users.findCalls)
	}
	if credentials.findCalls != 1 {
		t.Fatalf("expected one credential lookup, got %d", credentials.findCalls)
	}
	if hasher.verifyCalls != 1 || hasher.verifyPlaintext != "plain-secret" || hasher.verifyHash != "stored-hash" {
		t.Fatalf("expected hasher to verify signin password against stored hash")
	}
}

func TestSignInRejectsEmptyIdentifier(t *testing.T) {
	users := &fakeSignInUserRepository{}
	credentials := &fakeSignInCredentialRepository{}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "", Password: "secret"})
	if result != (SignInResult{}) {
		t.Fatal("expected empty result")
	}
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if users.findCalls != 0 || credentials.findCalls != 0 || hasher.verifyCalls != 0 {
		t.Fatal("expected no dependency interaction for invalid input")
	}
}

func TestSignInRejectsEmptyPassword(t *testing.T) {
	users := &fakeSignInUserRepository{}
	credentials := &fakeSignInCredentialRepository{}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: ""})
	if result != (SignInResult{}) {
		t.Fatal("expected empty result")
	}
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if users.findCalls != 0 || credentials.findCalls != 0 || hasher.verifyCalls != 0 {
		t.Fatal("expected no dependency interaction for invalid input")
	}
}

func TestSignInRejectsInvalidIdentifierPerDomain(t *testing.T) {
	users := &fakeSignInUserRepository{}
	credentials := &fakeSignInCredentialRepository{}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "   ", Password: "secret"})
	if result != (SignInResult{}) {
		t.Fatal("expected empty result")
	}
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if users.findCalls != 0 || credentials.findCalls != 0 || hasher.verifyCalls != 0 {
		t.Fatal("expected no dependency interaction for invalid input")
	}
}

func TestSignInRejectsUnknownUser(t *testing.T) {
	users := &fakeSignInUserRepository{findErr: ErrNotFound}
	credentials := &fakeSignInCredentialRepository{}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
	if credentials.findCalls != 0 || hasher.verifyCalls != 0 {
		t.Fatal("expected no credential lookup or password verification for unknown user")
	}
}

func TestSignInMapsUserRepositoryFailure(t *testing.T) {
	users := &fakeSignInUserRepository{findErr: errors.New("repository unavailable")}
	credentials := &fakeSignInCredentialRepository{}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if credentials.findCalls != 0 || hasher.verifyCalls != 0 {
		t.Fatal("expected no credential lookup or password verification after user repository failure")
	}
}

func TestSignInRejectsMissingCredential(t *testing.T) {
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findErr: ErrNotFound}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
	if hasher.verifyCalls != 0 {
		t.Fatal("expected no password verification when credential is missing")
	}
}

func TestSignInMapsCredentialRepositoryFailure(t *testing.T) {
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findErr: errors.New("repository unavailable")}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	_, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if hasher.verifyCalls != 0 {
		t.Fatal("expected no password verification after credential repository failure")
	}
}

func TestSignInRejectsWrongPassword(t *testing.T) {
	users := &fakeSignInUserRepository{findUser: signInUserForTest(t)}
	credentials := &fakeSignInCredentialRepository{findCredential: signInCredentialForTest(t)}
	hasher := &fakeSignInPasswordHasher{verifyErr: errors.New("password mismatch")}
	service := NewSignInService(users, credentials, hasher)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "user-1", Password: "wrong-secret"})
	if result != (SignInResult{}) {
		t.Fatal("expected empty result")
	}
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated error, got %v", err)
	}
}

func TestSignInReturnsMatchingUserID(t *testing.T) {
	identifier, err := identity.NewIdentifier("distinct-user")
	if err != nil {
		t.Fatalf("unexpected identifier error: %v", err)
	}
	distinctUser, err := user.New(identifier, time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected user error: %v", err)
	}
	distinctCredential, err := credential.NewPassword(identifier, "stored-hash", time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("unexpected credential error: %v", err)
	}

	users := &fakeSignInUserRepository{findUser: distinctUser}
	credentials := &fakeSignInCredentialRepository{findCredential: distinctCredential}
	hasher := &fakeSignInPasswordHasher{}
	service := NewSignInService(users, credentials, hasher)

	result, err := service.SignIn(context.Background(), SignInInput{Identifier: "distinct-user", Password: "plain-secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != identifier {
		t.Fatalf("expected user ID %q, got %q", identifier, result.UserID)
	}
}
