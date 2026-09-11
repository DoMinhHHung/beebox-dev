package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
)

type fakeResetPasswordResetRepository struct {
	findValue     passwordreset.PasswordReset
	findErr       error
	findCalls     int
	markUsedErr   error
	markUsedCalls int
	marked        passwordreset.PasswordReset
}

func (f *fakeResetPasswordResetRepository) Create(context.Context, passwordreset.PasswordReset) error {
	return nil
}

func (f *fakeResetPasswordResetRepository) FindByID(_ context.Context, _ string) (passwordreset.PasswordReset, error) {
	f.findCalls++
	return f.findValue, f.findErr
}

func (f *fakeResetPasswordResetRepository) FindPending(context.Context, identity.Identifier) (passwordreset.PasswordReset, error) {
	return passwordreset.PasswordReset{}, ErrNotFound
}

func (f *fakeResetPasswordResetRepository) MarkUsed(_ context.Context, value passwordreset.PasswordReset) error {
	f.markUsedCalls++
	f.marked = value
	return f.markUsedErr
}

type fakeResetCredentialRepository struct {
	findValue   credential.Credential
	findErr     error
	findCalls   int
	updateErr   error
	updateCalls int
	updated     credential.Credential
}

func (f *fakeResetCredentialRepository) Create(context.Context, credential.Credential) error {
	return nil
}

func (f *fakeResetCredentialRepository) FindByUserID(_ context.Context, _ identity.Identifier) (credential.Credential, error) {
	f.findCalls++
	return f.findValue, f.findErr
}

func (f *fakeResetCredentialRepository) Update(_ context.Context, value credential.Credential) error {
	f.updateCalls++
	f.updated = value
	return f.updateErr
}

type fakeResetPasswordHasher struct {
	hashValue string
	hashErr   error
	hashCalls int
	plaintext string
}

func (f *fakeResetPasswordHasher) Hash(_ context.Context, plaintext string) (string, error) {
	f.hashCalls++
	f.plaintext = plaintext
	return f.hashValue, f.hashErr
}

func (f *fakeResetPasswordHasher) Verify(context.Context, string, string) error {
	return nil
}

var _ PasswordResetRepository = (*fakeResetPasswordResetRepository)(nil)
var _ CredentialRepository = (*fakeResetCredentialRepository)(nil)
var _ PasswordHasher = (*fakeResetPasswordHasher)(nil)

func pendingResetForTest(t *testing.T, userID identity.Identifier, token string, now time.Time) passwordreset.PasswordReset {
	t.Helper()
	reset, err := passwordreset.New("reset-1", userID, hashPasswordResetToken(token), now.Add(-time.Minute), now.Add(time.Hour))
	if err != nil {
		t.Fatalf("password reset: %v", err)
	}
	return reset
}

func credentialForResetTest(t *testing.T, userID identity.Identifier, now time.Time) credential.Credential {
	t.Helper()
	c, err := credential.NewPassword(userID, "old-hash", now.Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("credential: %v", err)
	}
	return c
}

func TestResetPasswordSuccess(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("identifier: %v", err)
	}
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899"
	resets := &fakeResetPasswordResetRepository{findValue: pendingResetForTest(t, userID, token, now)}
	credentials := &fakeResetCredentialRepository{findValue: credentialForResetTest(t, userID, now)}
	hasher := &fakeResetPasswordHasher{hashValue: "new-stored-hash"}
	service := NewResetPasswordService(resets, credentials, hasher, &fakeClock{now: now})

	result, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID:     "reset-1",
		Token:       token,
		NewPassword: "new-secret",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.UserID != userID {
		t.Fatalf("unexpected user id: %s", result.UserID)
	}
	if resets.findCalls != 1 || credentials.findCalls != 1 || credentials.updateCalls != 1 || resets.markUsedCalls != 1 {
		t.Fatalf("unexpected call counts findReset=%d findCred=%d update=%d mark=%d",
			resets.findCalls, credentials.findCalls, credentials.updateCalls, resets.markUsedCalls)
	}
	if !resets.marked.IsUsed() {
		t.Fatal("expected reset to be marked used")
	}
	if credentials.updated.PasswordHash() != "new-stored-hash" {
		t.Fatalf("expected updated hash, got %s", credentials.updated.PasswordHash())
	}
	if credentials.updated.PasswordHash() == "new-secret" {
		t.Fatal("plaintext password must not be persisted")
	}
	if hasher.hashCalls != 1 || hasher.plaintext != "new-secret" {
		t.Fatal("expected hasher to receive plaintext once")
	}
}

func TestResetPasswordValidation(t *testing.T) {
	service := NewResetPasswordService(
		&fakeResetPasswordResetRepository{},
		&fakeResetCredentialRepository{},
		&fakeResetPasswordHasher{},
		&fakeClock{now: time.Now().UTC()},
	)
	cases := []ResetPasswordInput{
		{ResetID: "", Token: "t", NewPassword: "p"},
		{ResetID: "r", Token: "", NewPassword: "p"},
		{ResetID: "r", Token: "t", NewPassword: ""},
	}
	for _, input := range cases {
		_, err := service.ResetPassword(context.Background(), input)
		if !apperror.IsCode(err, apperror.CodeValidation) {
			t.Fatalf("expected validation for %+v, got %v", input, err)
		}
	}
}

func TestResetPasswordNotFound(t *testing.T) {
	service := NewResetPasswordService(
		&fakeResetPasswordResetRepository{findErr: ErrNotFound},
		&fakeResetCredentialRepository{},
		&fakeResetPasswordHasher{},
		&fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "missing", Token: "token", NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
}

func TestResetPasswordIncorrectToken(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "correct-token-value-with-enough-length-for-hex-style-secret-01"
	resets := &fakeResetPasswordResetRepository{findValue: pendingResetForTest(t, userID, token, now)}
	credentials := &fakeResetCredentialRepository{findValue: credentialForResetTest(t, userID, now)}
	service := NewResetPasswordService(resets, credentials, &fakeResetPasswordHasher{hashValue: "h"}, &fakeClock{now: now})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: "wrong-token", NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if credentials.updateCalls != 0 || resets.markUsedCalls != 0 {
		t.Fatal("must not update credential or consume reset on bad token")
	}
}

func TestResetPasswordExpired(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	created := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	expires := created.Add(time.Hour)
	token := "token-value"
	reset, _ := passwordreset.New("reset-1", userID, hashPasswordResetToken(token), created, expires)
	resets := &fakeResetPasswordResetRepository{findValue: reset}
	service := NewResetPasswordService(resets, &fakeResetCredentialRepository{}, &fakeResetPasswordHasher{}, &fakeClock{now: expires})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestResetPasswordAlreadyUsed(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token-value"
	reset := pendingResetForTest(t, userID, token, now)
	used, err := reset.Consume(now.Add(-30 * time.Second))
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	resets := &fakeResetPasswordResetRepository{findValue: used}
	service := NewResetPasswordService(resets, &fakeResetCredentialRepository{}, &fakeResetPasswordHasher{}, &fakeClock{now: now})

	_, err = service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
}

func TestResetPasswordCredentialNotFound(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token-value"
	resets := &fakeResetPasswordResetRepository{findValue: pendingResetForTest(t, userID, token, now)}
	credentials := &fakeResetCredentialRepository{findErr: ErrNotFound}
	service := NewResetPasswordService(resets, credentials, &fakeResetPasswordHasher{hashValue: "h"}, &fakeClock{now: now})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if resets.markUsedCalls != 0 {
		t.Fatal("must not consume reset when credential missing")
	}
}

func TestResetPasswordHashFailureDoesNotConsume(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token-value"
	resets := &fakeResetPasswordResetRepository{findValue: pendingResetForTest(t, userID, token, now)}
	credentials := &fakeResetCredentialRepository{findValue: credentialForResetTest(t, userID, now)}
	hasher := &fakeResetPasswordHasher{hashErr: errors.New("hasher down")}
	service := NewResetPasswordService(resets, credentials, hasher, &fakeClock{now: now})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if credentials.updateCalls != 0 || resets.markUsedCalls != 0 {
		t.Fatal("must not update or consume when hashing fails")
	}
}

func TestResetPasswordUpdateFailureDoesNotConsume(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token-value"
	resets := &fakeResetPasswordResetRepository{findValue: pendingResetForTest(t, userID, token, now)}
	credentials := &fakeResetCredentialRepository{
		findValue: credentialForResetTest(t, userID, now),
		updateErr: errors.New("write fail"),
	}
	service := NewResetPasswordService(resets, credentials, &fakeResetPasswordHasher{hashValue: "h"}, &fakeClock{now: now})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if resets.markUsedCalls != 0 {
		t.Fatal("must not consume reset when credential update fails")
	}
}

func TestResetPasswordMarkUsedFailure(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token-value"
	resets := &fakeResetPasswordResetRepository{
		findValue:   pendingResetForTest(t, userID, token, now),
		markUsedErr: errors.New("mark fail"),
	}
	credentials := &fakeResetCredentialRepository{findValue: credentialForResetTest(t, userID, now)}
	service := NewResetPasswordService(resets, credentials, &fakeResetPasswordHasher{hashValue: "h"}, &fakeClock{now: now})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if credentials.updateCalls != 1 {
		t.Fatal("credential update should have succeeded before mark used failure")
	}
}

func TestResetPasswordRepositoryFindFailure(t *testing.T) {
	service := NewResetPasswordService(
		&fakeResetPasswordResetRepository{findErr: errors.New("db")},
		&fakeResetCredentialRepository{},
		&fakeResetPasswordHasher{},
		&fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)},
	)
	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: "token", NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestResetPasswordCredentialLookupFailure(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	token := "token-value"
	resets := &fakeResetPasswordResetRepository{findValue: pendingResetForTest(t, userID, token, now)}
	credentials := &fakeResetCredentialRepository{findErr: errors.New("db")}
	service := NewResetPasswordService(resets, credentials, &fakeResetPasswordHasher{hashValue: "h"}, &fakeClock{now: now})

	_, err := service.ResetPassword(context.Background(), ResetPasswordInput{
		ResetID: "reset-1", Token: token, NewPassword: "secret",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}
