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
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
)

type fakePasswordResetRepository struct {
	createErr     error
	createCalls   int
	created       passwordreset.PasswordReset
	findValue     passwordreset.PasswordReset
	findErr       error
	findCalls     int
	markUsedErr   error
	markUsedCalls int
	marked        passwordreset.PasswordReset
}

func (f *fakePasswordResetRepository) Create(_ context.Context, value passwordreset.PasswordReset) error {
	f.createCalls++
	f.created = value
	return f.createErr
}

func (f *fakePasswordResetRepository) FindPending(_ context.Context, _ identity.Identifier) (passwordreset.PasswordReset, error) {
	f.findCalls++
	return f.findValue, f.findErr
}

func (f *fakePasswordResetRepository) MarkUsed(_ context.Context, value passwordreset.PasswordReset) error {
	f.markUsedCalls++
	f.marked = value
	return f.markUsedErr
}

var _ PasswordResetRepository = (*fakePasswordResetRepository)(nil)

type fakePasswordResetUserRepository struct {
	findUser  user.User
	findErr   error
	findCalls int
}

func (f *fakePasswordResetUserRepository) Create(context.Context, user.User) error {
	return nil
}

func (f *fakePasswordResetUserRepository) FindByIdentifier(context.Context, identity.Identifier) (user.User, error) {
	f.findCalls++
	return f.findUser, f.findErr
}

var _ UserRepository = (*fakePasswordResetUserRepository)(nil)

func TestRequestPasswordResetSuccess(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("identifier: %v", err)
	}
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	foundUser, err := user.New(userID, now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("user: %v", err)
	}

	users := &fakePasswordResetUserRepository{findUser: foundUser}
	resets := &fakePasswordResetRepository{}
	service := NewRequestPasswordResetService(users, resets, &fakeClock{now: now})

	result, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{
		Identifier: "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Token == "" || len(result.Token) != 64 {
		t.Fatalf("expected 64-char hex token, got %q", result.Token)
	}
	if result.ResetID == "" {
		t.Fatal("expected reset id")
	}
	if !result.ExpiresAt.Equal(now.Add(time.Hour)) {
		t.Fatalf("unexpected expires at: %v", result.ExpiresAt)
	}
	if users.findCalls != 1 {
		t.Fatalf("expected one user lookup, got %d", users.findCalls)
	}
	if resets.createCalls != 1 {
		t.Fatalf("expected one create, got %d", resets.createCalls)
	}
	created := resets.created
	if created.TokenHash() == result.Token {
		t.Fatal("plaintext token must not be stored")
	}
	sum := sha256.Sum256([]byte(result.Token))
	expectedHash := hex.EncodeToString(sum[:])
	if created.TokenHash() != expectedHash {
		t.Fatalf("expected hash of token, got %s want %s", created.TokenHash(), expectedHash)
	}
	if created.UserID() != userID {
		t.Fatal("unexpected user id on created reset")
	}
	if !created.IsPending(now) {
		t.Fatal("expected pending")
	}
}

func TestRequestPasswordResetValidation(t *testing.T) {
	service := NewRequestPasswordResetService(
		&fakePasswordResetUserRepository{},
		&fakePasswordResetRepository{},
		&fakeClock{now: time.Now().UTC()},
	)
	_, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: ""})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}

func TestRequestPasswordResetUserNotFound(t *testing.T) {
	users := &fakePasswordResetUserRepository{findErr: ErrNotFound}
	resets := &fakePasswordResetRepository{}
	service := NewRequestPasswordResetService(users, resets, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})

	_, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "missing"})
	if !apperror.IsCode(err, apperror.CodeNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if resets.createCalls != 0 {
		t.Fatal("must not create recovery when user is missing")
	}
}

func TestRequestPasswordResetUserLookupFailure(t *testing.T) {
	users := &fakePasswordResetUserRepository{findErr: errors.New("db down")}
	resets := &fakePasswordResetRepository{}
	service := NewRequestPasswordResetService(users, resets, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})

	_, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "user-1"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if resets.createCalls != 0 {
		t.Fatal("must not create after lookup failure")
	}
}

func TestRequestPasswordResetPersistenceFailure(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	foundUser, _ := user.New(userID, now.Add(-time.Hour))
	users := &fakePasswordResetUserRepository{findUser: foundUser}
	resets := &fakePasswordResetRepository{createErr: errors.New("write fail")}
	service := NewRequestPasswordResetService(users, resets, &fakeClock{now: now})

	_, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "user-1"})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestGeneratePasswordResetTokenUsesCryptoRand(t *testing.T) {
	fixed := bytes.Repeat([]byte{0xab}, passwordResetTokenBytes)
	r := bytes.NewReader(fixed)
	token, err := generatePasswordResetToken(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(token) != 64 {
		t.Fatalf("expected 64 hex chars, got %d", len(token))
	}
	expected := hex.EncodeToString(fixed)
	if token != expected {
		t.Fatalf("unexpected token encoding")
	}
}

func TestHashPasswordResetToken(t *testing.T) {
	h1 := hashPasswordResetToken("abc")
	h2 := hashPasswordResetToken("abc")
	h3 := hashPasswordResetToken("xyz")
	if h1 != h2 {
		t.Fatal("hash must be deterministic")
	}
	if h1 == h3 {
		t.Fatal("different tokens must hash differently")
	}
	if h1 == "abc" {
		t.Fatal("hash must not equal plaintext")
	}
}

func TestGeneratePasswordResetTokenRejectsBadReader(t *testing.T) {
	_, err := generatePasswordResetToken(io.LimitReader(bytes.NewReader(nil), 0))
	if err == nil {
		t.Fatal("expected error from empty reader")
	}
}
