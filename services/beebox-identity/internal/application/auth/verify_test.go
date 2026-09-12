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

func TestVerifySuccess(t *testing.T) {
	userID, err := identity.NewIdentifier("user-1")
	if err != nil {
		t.Fatalf("identifier: %v", err)
	}
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	code := "482913"
	codeHash := hashVerificationCode("test-verification-secret", code)
	pending, err := verification.New("ver-1", userID, verification.TypeEmail, "user@example.com", codeHash, now.Add(-time.Minute), now.Add(14*time.Minute))
	if err != nil {
		t.Fatalf("verification: %v", err)
	}

	repo := &fakeVerificationRepository{findValue: pending}
	service := NewVerifyService(repo, &fakeClock{now: now}, "test-verification-secret")

	result, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
		Code:   code,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VerificationID != "ver-1" || result.UserID != userID || result.Target != "user@example.com" {
		t.Fatalf("unexpected result: %+v", result)
	}
	if repo.findCalls != 1 || repo.markUsedCalls != 1 {
		t.Fatalf("expected find and mark used, got find=%d mark=%d", repo.findCalls, repo.markUsedCalls)
	}
	if !repo.marked.IsUsed() {
		t.Fatal("expected marked verification to be used")
	}
}

func TestVerifyValidation(t *testing.T) {
	service := NewVerifyService(&fakeVerificationRepository{}, &fakeClock{now: time.Now().UTC()}, "test-verification-secret")
	cases := []VerifyInput{
		{UserID: "", Type: "email", Target: "a@b.com", Code: "123456"},
		{UserID: "user-1", Type: "oauth", Target: "a@b.com", Code: "123456"},
		{UserID: "user-1", Type: "email", Target: "", Code: "123456"},
		{UserID: "user-1", Type: "email", Target: "a@b.com", Code: ""},
	}
	for _, input := range cases {
		_, err := service.Verify(context.Background(), input)
		if !apperror.IsCode(err, apperror.CodeValidation) {
			t.Fatalf("expected validation for %+v, got %v", input, err)
		}
	}
}

func TestVerifyNotFound(t *testing.T) {
	repo := &fakeVerificationRepository{findErr: ErrNotFound}
	service := NewVerifyService(repo, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, "test-verification-secret")
	_, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
		Code:   "123456",
	})
	if !apperror.IsCode(err, apperror.CodeNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestVerifyInvalidCode(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	pending, _ := verification.New("ver-1", userID, verification.TypeEmail, "user@example.com", hashVerificationCode("test-verification-secret", "111111"), now.Add(-time.Minute), now.Add(14*time.Minute))
	repo := &fakeVerificationRepository{findValue: pending}
	service := NewVerifyService(repo, &fakeClock{now: now}, "test-verification-secret")

	_, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
		Code:   "999999",
	})
	if !apperror.IsCode(err, apperror.CodeUnauthenticated) {
		t.Fatalf("expected unauthenticated, got %v", err)
	}
	if repo.markUsedCalls != 0 {
		t.Fatal("must not mark used on invalid code")
	}
}

func TestVerifyExpired(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	created := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	expires := created.Add(15 * time.Minute)
	pending, _ := verification.New("ver-1", userID, verification.TypePhone, "+1555", hashVerificationCode("test-verification-secret", "123456"), created, expires)
	repo := &fakeVerificationRepository{findValue: pending}
	service := NewVerifyService(repo, &fakeClock{now: expires}, "test-verification-secret")

	_, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "phone",
		Target: "+1555",
		Code:   "123456",
	})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict for expired, got %v", err)
	}
	if repo.markUsedCalls != 0 {
		t.Fatal("must not mark used when expired")
	}
}

func TestVerifyAlreadyUsed(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	created := time.Date(2026, time.January, 1, 10, 0, 0, 0, time.UTC)
	expires := created.Add(15 * time.Minute)
	pending, _ := verification.New("ver-1", userID, verification.TypeEmail, "a@b.com", hashVerificationCode("test-verification-secret", "123456"), created, expires)
	used, err := pending.Consume(created.Add(time.Minute))
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	repo := &fakeVerificationRepository{findValue: used}
	service := NewVerifyService(repo, &fakeClock{now: created.Add(2 * time.Minute)}, "test-verification-secret")

	_, err = service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "a@b.com",
		Code:   "123456",
	})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict for used, got %v", err)
	}
}

func TestVerifyRepositoryFindFailure(t *testing.T) {
	repo := &fakeVerificationRepository{findErr: errors.New("db")}
	service := NewVerifyService(repo, &fakeClock{now: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)}, "test-verification-secret")
	_, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "a@b.com",
		Code:   "123456",
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestVerifyMarkUsedFailure(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	code := "482913"
	pending, _ := verification.New("ver-1", userID, verification.TypeEmail, "user@example.com", hashVerificationCode("test-verification-secret", code), now.Add(-time.Minute), now.Add(14*time.Minute))
	repo := &fakeVerificationRepository{findValue: pending, markUsedErr: errors.New("write fail")}
	service := NewVerifyService(repo, &fakeClock{now: now}, "test-verification-secret")

	_, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
		Code:   code,
	})
	if !apperror.IsCode(err, apperror.CodeDependencyFailure) {
		t.Fatalf("expected dependency failure, got %v", err)
	}
}

func TestVerifyMarkUsedAlreadyConsumed(t *testing.T) {
	userID, _ := identity.NewIdentifier("user-1")
	now := time.Date(2026, time.January, 1, 12, 0, 0, 0, time.UTC)
	code := "482913"
	pending, _ := verification.New("ver-1", userID, verification.TypeEmail, "user@example.com", hashVerificationCode("test-verification-secret", code), now.Add(-time.Minute), now.Add(14*time.Minute))
	repo := &fakeVerificationRepository{findValue: pending, markUsedErr: ErrNotFound}
	service := NewVerifyService(repo, &fakeClock{now: now}, "test-verification-secret")

	_, err := service.Verify(context.Background(), VerifyInput{
		UserID: "user-1",
		Type:   "email",
		Target: "user@example.com",
		Code:   code,
	})
	if !apperror.IsCode(err, apperror.CodeConflict) {
		t.Fatalf("expected conflict for concurrent consume, got %v", err)
	}
}
