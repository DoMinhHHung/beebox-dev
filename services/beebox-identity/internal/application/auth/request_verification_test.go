package auth

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

type fakeVerificationUserRepository struct {
	findUser  user.User
	findErr   error
	findCalls int
}

func (f *fakeVerificationUserRepository) Create(context.Context, user.User) error { return nil }
func (f *fakeVerificationUserRepository) FindByIdentifier(context.Context, identity.Identifier) (user.User, error) {
	f.findCalls++
	return f.findUser, f.findErr
}

type fakeVerificationMailer struct {
	err   error
	calls int
	last  VerificationDeliveryMessage
}

func (f *fakeVerificationMailer) SendVerificationEmail(_ context.Context, message VerificationDeliveryMessage) error {
	f.calls++
	f.last = message
	return f.err
}

type fakeVerificationSMSSender struct {
	err   error
	calls int
	last  VerificationDeliveryMessage
}

func (f *fakeVerificationSMSSender) SendVerificationSMS(_ context.Context, message VerificationDeliveryMessage) error {
	f.calls++
	f.last = message
	return f.err
}

type fakeVerificationRepository struct {
	createErr     error
	createCalls   int
	created       verification.Verification
	findValue     verification.Verification
	findErr       error
	findCalls     int
	markUsedErr   error
	markUsedCalls int
	marked        verification.Verification
}

func (f *fakeVerificationRepository) Create(_ context.Context, value verification.Verification) error {
	f.createCalls++
	f.created = value
	return f.createErr
}

func (f *fakeVerificationRepository) FindPending(context.Context, identity.Identifier, verification.Type, string) (verification.Verification, error) {
	f.findCalls++
	return f.findValue, f.findErr
}

func (f *fakeVerificationRepository) MarkUsed(_ context.Context, value verification.Verification) error {
	f.markUsedCalls++
	f.marked = value
	return f.markUsedErr
}

var _ VerificationRepository = (*fakeVerificationRepository)(nil)

func TestRequestVerificationSuccessEmail(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	userID, _ := identity.NewIdentifier("user-1")
	u, _ := user.New(userID, now)
	users := &fakeVerificationUserRepository{findUser: u}
	repo := &fakeVerificationRepository{}
	mailer := &fakeVerificationMailer{}
	service := NewRequestVerificationService(users, repo, mailer, &fakeVerificationSMSSender{}, &fakeClock{now: now}, "test-verification-secret")

	result, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VerificationID == "" {
		t.Fatal("expected verification id")
	}
	if !result.ExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("unexpected expires: %v", result.ExpiresAt)
	}
	if repo.createCalls != 1 {
		t.Fatal("expected create")
	}
	if mailer.calls != 1 || mailer.last.Code == "" || len(mailer.last.Code) != 6 {
		t.Fatal("expected email delivery with code")
	}
	if repo.created.CodeHash() == mailer.last.Code {
		t.Fatal("must not persist plaintext code")
	}
}

func TestRequestVerificationSuccessPhone(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	sms := &fakeVerificationSMSSender{}
	service := NewRequestVerificationService(&fakeVerificationUserRepository{findUser: mustUser(t)}, &fakeVerificationRepository{}, &fakeVerificationMailer{}, sms, &fakeClock{now: now}, "test-verification-secret")
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "phone",
		Target: "+15551234567",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sms.calls != 1 {
		t.Fatal("expected sms delivery")
	}
}

func TestRequestVerificationDeliveryFailureStillAccepted(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	mailer := &fakeVerificationMailer{err: errors.New("smtp down")}
	repo := &fakeVerificationRepository{}
	service := NewRequestVerificationService(&fakeVerificationUserRepository{findUser: mustUser(t)}, repo, mailer, &fakeVerificationSMSSender{}, &fakeClock{now: now}, "test-verification-secret")
	result, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
	})
	if err != nil {
		t.Fatalf("delivery failure must not change public success, got %v", err)
	}
	if result.VerificationID == "" {
		t.Fatal("expected verification id")
	}
	if repo.createCalls != 1 {
		t.Fatal("expected verification created for existing user")
	}
}

func TestRequestVerificationUnknownUserEnumerationSafe(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	users := &fakeVerificationUserRepository{findErr: ErrNotFound}
	repo := &fakeVerificationRepository{}
	mailer := &fakeVerificationMailer{}
	sms := &fakeVerificationSMSSender{}
	service := NewRequestVerificationService(users, repo, mailer, sms, &fakeClock{now: now}, "test-verification-secret")
	result, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "missing-user",
		Type:   "email",
		Target: "missing@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VerificationID == "" || result.ExpiresAt.IsZero() {
		t.Fatal("expected same public success shape")
	}
	if repo.createCalls != 0 {
		t.Fatal("must not create verification for unknown user")
	}
	if mailer.calls != 0 || sms.calls != 0 {
		t.Fatal("must not deliver for unknown user")
	}
}

func TestRequestVerificationUserLookupFailure(t *testing.T) {
	users := &fakeVerificationUserRepository{findErr: errors.New("db down")}
	repo := &fakeVerificationRepository{}
	service := NewRequestVerificationService(users, repo, &fakeVerificationMailer{}, &fakeVerificationSMSSender{}, &fakeClock{now: time.Now().UTC()}, "test-verification-secret")
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
	if repo.createCalls != 0 {
		t.Fatal("must not create on lookup failure")
	}
}

func TestRequestVerificationValidation(t *testing.T) {
	service := NewRequestVerificationService(&fakeVerificationUserRepository{findUser: mustUser(t)}, &fakeVerificationRepository{}, &fakeVerificationMailer{}, &fakeVerificationSMSSender{}, &fakeClock{now: time.Now().UTC()}, "test-verification-secret")
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{UserID: "", Type: "email", Target: "a"})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}

func mustUser(t *testing.T) user.User {
	t.Helper()
	id, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("id: %v", err)
	}
	u, err := user.New(id, time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("user: %v", err)
	}
	return u
}

func TestHashVerificationCodeDeterministic(t *testing.T) {
	secret := "test-verification-secret"
	a := hashVerificationCode(secret, "123456")
	b := hashVerificationCode(secret, "123456")
	if a != b || a == "" {
		t.Fatalf("expected deterministic non-empty digest")
	}
}

func TestHashVerificationCodeChangesWithSecret(t *testing.T) {
	a := hashVerificationCode("secret-a", "123456")
	b := hashVerificationCode("secret-b", "123456")
	if a == b {
		t.Fatal("expected different digests for different secrets")
	}
}

func TestRequestVerificationStoresHMACDigest(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	users := &fakeVerificationUserRepository{findUser: mustUser(t)}
	repo := &fakeVerificationRepository{}
	mailer := &fakeVerificationMailer{}
	service := NewRequestVerificationService(users, repo, mailer, &fakeVerificationSMSSender{}, &fakeClock{now: now}, "test-verification-secret")
	service.randReader = bytes.NewReader(make([]byte, 64))
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if repo.createCalls != 1 {
		t.Fatal("expected verification created")
	}
	stored := repo.created.CodeHash()
	if stored == "" {
		t.Fatal("expected stored code hash")
	}
	if stored == hashVerificationCode("other-secret", "000000") {
		t.Fatal("stored hash should not match unrelated secret/code")
	}
}
