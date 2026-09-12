package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
)

type fakePasswordResetMailer struct {
	err   error
	calls int
	last  PasswordResetDeliveryMessage
}

func (f *fakePasswordResetMailer) SendPasswordReset(_ context.Context, message PasswordResetDeliveryMessage) error {
	f.calls++
	f.last = message
	return f.err
}

type fakePasswordResetUserRepository struct {
	findUser user.User
	findErr  error
}

func (f *fakePasswordResetUserRepository) Create(context.Context, user.User) error { return nil }
func (f *fakePasswordResetUserRepository) FindByIdentifier(context.Context, identity.Identifier) (user.User, error) {
	return f.findUser, f.findErr
}

type fakePasswordResetRepository struct {
	createErr   error
	createCalls int
	created     passwordreset.PasswordReset
}

func (f *fakePasswordResetRepository) Create(_ context.Context, value passwordreset.PasswordReset) error {
	f.createCalls++
	f.created = value
	return f.createErr
}
func (f *fakePasswordResetRepository) FindByID(context.Context, string) (passwordreset.PasswordReset, error) {
	return passwordreset.PasswordReset{}, ErrNotFound
}
func (f *fakePasswordResetRepository) FindPending(context.Context, identity.Identifier) (passwordreset.PasswordReset, error) {
	return passwordreset.PasswordReset{}, ErrNotFound
}
func (f *fakePasswordResetRepository) MarkUsed(context.Context, passwordreset.PasswordReset) error {
	return nil
}

func TestRequestPasswordResetSuccess(t *testing.T) {
	userID, _ := identity.NewIdentifier("user@example.com")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	foundUser, _ := user.New(userID, now.Add(-time.Hour))
	users := &fakePasswordResetUserRepository{findUser: foundUser}
	resets := &fakePasswordResetRepository{}
	mailer := &fakePasswordResetMailer{}
	service := NewRequestPasswordResetService(users, resets, mailer, &fakeClock{now: now})

	result, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "user@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "accepted" {
		t.Fatalf("expected accepted, got %q", result.Status)
	}
	if resets.createCalls != 1 {
		t.Fatal("expected create")
	}
	if mailer.calls != 1 || mailer.last.Token == "" || mailer.last.ResetID == "" {
		t.Fatal("expected delivery with token")
	}
	if mailer.last.Target != "user@example.com" {
		t.Fatalf("expected email target, got %q", mailer.last.Target)
	}
	if resets.created.TokenHash() == mailer.last.Token {
		t.Fatal("must not persist plaintext token")
	}
}

func TestRequestPasswordResetUnknownUser(t *testing.T) {
	users := &fakePasswordResetUserRepository{findErr: ErrNotFound}
	resets := &fakePasswordResetRepository{}
	mailer := &fakePasswordResetMailer{}
	service := NewRequestPasswordResetService(users, resets, mailer, &fakeClock{now: time.Now().UTC()})
	result, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "missing@example.com"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Status != "accepted" {
		t.Fatal("expected accepted")
	}
	if resets.createCalls != 0 || mailer.calls != 0 {
		t.Fatal("must not create or deliver for unknown user")
	}
}

func TestRequestPasswordResetDeliveryFailureStillAccepted(t *testing.T) {
	userID, _ := identity.NewIdentifier("user@example.com")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	foundUser, _ := user.New(userID, now)
	users := &fakePasswordResetUserRepository{findUser: foundUser}
	resets := &fakePasswordResetRepository{}
	mailer := &fakePasswordResetMailer{err: errors.New("smtp")}
	service := NewRequestPasswordResetService(users, resets, mailer, &fakeClock{now: now})
	result, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "user@example.com"})
	if err != nil {
		t.Fatalf("delivery failure must not change public success, got %v", err)
	}
	if result.Status != "accepted" {
		t.Fatalf("expected accepted, got %q", result.Status)
	}
	if resets.createCalls != 1 {
		t.Fatal("expected reset created")
	}
}

func TestRequestPasswordResetNonEmailIdentifier(t *testing.T) {
	service := NewRequestPasswordResetService(&fakePasswordResetUserRepository{}, &fakePasswordResetRepository{}, &fakePasswordResetMailer{}, &fakeClock{now: time.Now().UTC()})
	_, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: "user-1"})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}

func TestRequestPasswordResetValidation(t *testing.T) {
	service := NewRequestPasswordResetService(&fakePasswordResetUserRepository{}, &fakePasswordResetRepository{}, &fakePasswordResetMailer{}, &fakeClock{now: time.Now().UTC()})
	_, err := service.RequestPasswordReset(context.Background(), RequestPasswordResetInput{Identifier: ""})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}
