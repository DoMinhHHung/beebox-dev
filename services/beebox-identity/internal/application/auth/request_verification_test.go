package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

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
	repo := &fakeVerificationRepository{}
	mailer := &fakeVerificationMailer{}
	service := NewRequestVerificationService(repo, mailer, &fakeVerificationSMSSender{}, &fakeClock{now: now})

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
	service := NewRequestVerificationService(&fakeVerificationRepository{}, &fakeVerificationMailer{}, sms, &fakeClock{now: now})
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

func TestRequestVerificationDeliveryFailure(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	mailer := &fakeVerificationMailer{err: errors.New("smtp down")}
	service := NewRequestVerificationService(&fakeVerificationRepository{}, mailer, &fakeVerificationSMSSender{}, &fakeClock{now: now})
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestRequestVerificationValidation(t *testing.T) {
	service := NewRequestVerificationService(&fakeVerificationRepository{}, &fakeVerificationMailer{}, &fakeVerificationSMSSender{}, &fakeClock{now: time.Now().UTC()})
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{UserID: "", Type: "email", Target: "a"})
	if !apperror.IsCode(err, apperror.CodeValidation) {
		t.Fatalf("expected validation, got %v", err)
	}
}
