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

type fakeUserRepository struct {
	findUser    user.User
	findErr     error
	createErr   error
	findCalls   int
	createCalls int
	createdUser user.User
}

func (f *fakeUserRepository) Create(_ context.Context, value user.User) error {
	f.createCalls++
	f.createdUser = value
	return f.createErr
}

func (f *fakeUserRepository) FindByIdentifier(context.Context, identity.Identifier) (user.User, error) {
	f.findCalls++
	return f.findUser, f.findErr
}

type fakeCredentialRepository struct {
	createErr         error
	createCalls       int
	createdCredential credential.Credential
}

func (f *fakeCredentialRepository) Create(_ context.Context, value credential.Credential) error {
	f.createCalls++
	f.createdCredential = value
	return f.createErr
}

func (f *fakeCredentialRepository) FindByUserID(context.Context, identity.Identifier) (credential.Credential, error) {
	return credential.Credential{}, ErrNotFound
}

type fakePasswordHasher struct {
	hashValue string
	hashErr   error
	hashCalls int
	plaintext string
}

func (f *fakePasswordHasher) Hash(_ context.Context, plaintext string) (string, error) {
	f.hashCalls++
	f.plaintext = plaintext
	return f.hashValue, f.hashErr
}

func (f *fakePasswordHasher) Verify(context.Context, string, string) error {
	return nil
}

type fakeClock struct {
	now   time.Time
	calls int
}

func (f *fakeClock) Now() time.Time {
	f.calls++
	return f.now
}

var _ UserRepository = (*fakeUserRepository)(nil)
var _ CredentialRepository = (*fakeCredentialRepository)(nil)
var _ PasswordHasher = (*fakePasswordHasher)(nil)
var _ Clock = (*fakeClock)(nil)

func TestSignUpSuccess(t *testing.T) {
	clock := &fakeClock{now: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)}
	users := &fakeUserRepository{findErr: ErrNotFound}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	service := NewSignUpService(users, credentials, hasher, clock)

	result, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "plain-secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID.String() != "user-1" {
		t.Fatalf("expected user ID user-1, got %q", result.UserID)
	}
	if users.createCalls != 1 || credentials.createCalls != 1 {
		t.Fatalf("expected one create call per repository, got users=%d credentials=%d", users.createCalls, credentials.createCalls)
	}
	if hasher.hashCalls != 1 || hasher.plaintext != "plain-secret" {
		t.Fatalf("expected hasher to receive signup password once")
	}
	if clock.calls != 1 {
		t.Fatalf("expected clock to be called once, got %d", clock.calls)
	}
	if credentials.createdCredential.PasswordHash() != "stored-hash" {
		t.Fatal("expected credential to contain the returned password hash")
	}
}

func TestSignUpRejectsInvalidInput(t *testing.T) {
	users := &fakeUserRepository{}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	result, err := service.SignUp(context.Background(), SignUpInput{Identifier: " ", Password: "secret"})
	if result != (SignUpResult{}) {
		t.Fatal("expected empty result")
	}
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if users.findCalls != 0 || users.createCalls != 0 || credentials.createCalls != 0 || hasher.hashCalls != 0 {
		t.Fatal("expected no dependency interaction for invalid input")
	}
}

func TestSignUpRejectsDuplicateIdentity(t *testing.T) {
	users := &fakeUserRepository{findUser: userForTest(t), findErr: nil}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict error, got %v", err)
	}
	if hasher.hashCalls != 0 || users.createCalls != 0 || credentials.createCalls != 0 {
		t.Fatal("expected no hashing or persistence for duplicate identity")
	}
}

func TestSignUpContinuesWhenIdentityIsNotFound(t *testing.T) {
	users := &fakeUserRepository{findErr: ErrNotFound}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignUpMapsLookupFailure(t *testing.T) {
	users := &fakeUserRepository{findErr: errors.New("repository unavailable")}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if hasher.hashCalls != 0 || users.createCalls != 0 || credentials.createCalls != 0 {
		t.Fatal("expected no hashing or persistence after lookup failure")
	}
}

func TestSignUpMapsHashFailure(t *testing.T) {
	hashErr := errors.New("hasher unavailable")
	users := &fakeUserRepository{findErr: ErrNotFound}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashErr: hashErr}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if users.createCalls != 0 || credentials.createCalls != 0 || clock.calls != 0 {
		t.Fatal("expected no domain creation or persistence after hash failure")
	}
}

func TestSignUpMapsUserRepositoryFailure(t *testing.T) {
	repositoryErr := errors.New("write unavailable")
	users := &fakeUserRepository{findErr: ErrNotFound, createErr: repositoryErr}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if credentials.createCalls != 0 {
		t.Fatal("expected credential persistence to wait for user persistence")
	}
}

func TestSignUpMapsCredentialRepositoryFailure(t *testing.T) {
	repositoryErr := errors.New("write unavailable")
	users := &fakeUserRepository{findErr: ErrNotFound}
	credentials := &fakeCredentialRepository{createErr: repositoryErr}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{now: time.Now()}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if users.createCalls != 1 || credentials.createCalls != 1 {
		t.Fatal("expected both persistence operations in sequential flow")
	}
}

func TestSignUpMapsDomainFailure(t *testing.T) {
	users := &fakeUserRepository{findErr: ErrNotFound}
	credentials := &fakeCredentialRepository{}
	hasher := &fakePasswordHasher{hashValue: "stored-hash"}
	clock := &fakeClock{}
	service := NewSignUpService(users, credentials, hasher, clock)

	_, err := service.SignUp(context.Background(), SignUpInput{Identifier: "user-1", Password: "secret"})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation error, got %v", err)
	}
	if users.createCalls != 0 || credentials.createCalls != 0 {
		t.Fatal("expected no persistence after domain failure")
	}
}

func userForTest(t *testing.T) user.User {
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