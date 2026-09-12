package auth

import (
	"context"
	"crypto/subtle"
	"errors"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

type VerifyInput struct {
	UserID string
	Type   string
	Target string
	Code   string
}

type VerifyResult struct {
	VerificationID string
	UserID         identity.Identifier
	Type           verification.Type
	Target         string
}

type VerifyService struct {
	verifications VerificationRepository
	clock         Clock
	codeSecret    string
}

func NewVerifyService(verifications VerificationRepository, clock Clock, codeSecret string) *VerifyService {
	return &VerifyService{
		verifications: verifications,
		clock:         clock,
		codeSecret:    codeSecret,
	}
}

func (s *VerifyService) Verify(ctx context.Context, input VerifyInput) (VerifyResult, error) {
	userID, err := identity.NewIdentifier(input.UserID)
	if err != nil {
		return VerifyResult{}, apperror.Wrap(apperror.CodeValidation, "invalid verify input", err)
	}
	vtype := verification.Type(strings.TrimSpace(input.Type))
	if !vtype.Valid() {
		return VerifyResult{}, apperror.New(apperror.CodeValidation, "invalid verify input")
	}
	target := strings.TrimSpace(input.Target)
	if target == "" || input.Code == "" {
		return VerifyResult{}, apperror.New(apperror.CodeValidation, "invalid verify input")
	}

	found, err := s.verifications.FindPending(ctx, userID, vtype, target)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return VerifyResult{}, apperror.New(apperror.CodeNotFound, "verification not found")
	case err != nil:
		return VerifyResult{}, translateRepositoryError(err)
	}

	now := s.clock.Now()
	if !found.IsPending(now) {
		if found.IsUsed() {
			return VerifyResult{}, apperror.New(apperror.CodeConflict, "verification already used")
		}
		if found.IsExpired(now) {
			return VerifyResult{}, apperror.New(apperror.CodeConflict, "verification expired")
		}
		return VerifyResult{}, apperror.New(apperror.CodeConflict, "verification not usable")
	}

	suppliedHash := hashVerificationCode(s.codeSecret, input.Code)
	if subtle.ConstantTimeCompare([]byte(suppliedHash), []byte(found.CodeHash())) != 1 {
		return VerifyResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid verification code")
	}

	consumed, err := found.Consume(now)
	if err != nil {
		return VerifyResult{}, translateVerificationDomainError(err)
	}

	if err := s.verifications.MarkUsed(ctx, consumed); err != nil {
		if errors.Is(err, ErrNotFound) {
			return VerifyResult{}, apperror.New(apperror.CodeConflict, "verification already used")
		}
		return VerifyResult{}, translateRepositoryError(err)
	}

	return VerifyResult{
		VerificationID: consumed.ID(),
		UserID:         consumed.UserID(),
		Type:           consumed.Type(),
		Target:         consumed.Target(),
	}, nil
}
