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
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

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

func (f *fakeVerificationRepository) FindPending(_ context.Context, _ identity.Identifier, _ verification.Type, _ string) (verification.Verification, error) {
	f.findCalls++
	return f.findValue, f.findErr
}

func (f *fakeVerificationRepository) MarkUsed(_ context.Context, value verification.Verification) error {
	f.markUsedCalls++
	f.marked = value
	return f.markUsedErr
}

var _ VerificationRepository = (*fakeVerificationRepository)(nil)

func TestRequestVerificationSuccess(t *testing.T) {
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	clock := &fakeClock{now: now}
	repo := &fakeVerificationRepository{}
	service := NewRequestVerificationService(repo, clock)

	result, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Code == "" || len(result.Code) != 6 {
		t.Fatalf("expected 6-digit code, got %q", result.Code)
	}
	if result.VerificationID == "" {
		t.Fatal("expected verification id")
	}
	if !result.ExpiresAt.Equal(now.Add(15 * time.Minute)) {
		t.Fatalf("unexpected expires at: %v", result.ExpiresAt)
	}
	if repo.createCalls != 1 {
		t.Fatalf("expected one create, got %d", repo.createCalls)
	}
	created := repo.created
	if created.CodeHash() == result.Code {
		t.Fatal("plaintext code must not be stored")
	}
	sum := sha256.Sum256([]byte(result.Code))
	expectedHash := hex.EncodeToString(sum[:])
	if created.CodeHash() != expectedHash {
		t.Fatalf("expected hash of code, got %s want %s", created.CodeHash(), expectedHash)
	}
	if created.Target() != "user@example.com" || created.Type() != verification.TypeEmail {
		t.Fatal("unexpected created verification fields")
	}
	if !created.IsPending(now) {
		t.Fatal("expected pending")
	}
}

func TestRequestVerificationValidation(t *testing.T) {
	service := NewRequestVerificationService(&fakeVerificationRepository{}, &fakeClock{now: time.Now().UTC()})

	cases := []RequestVerificationInput{
		{UserID: "", Type: "email", Target: "a@b.com"},
		{UserID: "user-1", Type: "oauth", Target: "a@b.com"},
		{UserID: "user-1", Type: "email", Target: ""},
		{UserID: "user-1", Type: "", Target: "a@b.com"},
	}
	for _, input := range cases {
		_, err := service.RequestVerification(context.Background(), input)
		if !apperror.IsCode(err, apperror.CodeValidation) {
			t.Fatalf("expected validation for %+v, got %v", input, err)
		}
	}
}

func TestRequestVerificationRepositoryFailure(t *testing.T) {
	repo := &fakeVerificationRepository{createErr: errors.New("db down")}
	service := NewRequestVerificationService(repo, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)})
	_, err := service.RequestVerification(context.Background(), RequestVerificationInput{
		UserID: "user-1",
		Type:   "phone",
		Target: "+15550001111",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestGenerateNumericCodeUsesCryptoRand(t *testing.T) {
	fixed := bytes.Repeat([]byte{0x01, 0x02, 0x03, 0x04}, 8)
	r := bytes.NewReader(fixed)
	code, err := generateNumericCode(r, 6)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != 6 {
		t.Fatalf("expected 6 digits, got %q", code)
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			t.Fatalf("non-digit in code: %q", code)
		}
	}
}

func TestHashVerificationCode(t *testing.T) {
	h1 := hashVerificationCode("123456")
	h2 := hashVerificationCode("123456")
	h3 := hashVerificationCode("654321")
	if h1 != h2 {
		t.Fatal("hash must be deterministic")
	}
	if h1 == h3 {
		t.Fatal("different codes must hash differently")
	}
	if h1 == "123456" {
		t.Fatal("hash must not equal plaintext")
	}
}

func TestGenerateNumericCodeRejectsBadReader(t *testing.T) {
	_, err := generateNumericCode(io.LimitReader(bytes.NewReader(nil), 0), 6)
	if err == nil {
		t.Fatal("expected error from empty reader")
	}
}
