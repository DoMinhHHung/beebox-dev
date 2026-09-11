package auth

import (
	"context"
	"crypto/subtle"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type ResetPasswordInput struct {
	ResetID     string
	Token       string
	NewPassword string
}

type ResetPasswordResult struct {
	UserID identity.Identifier
}

type ResetPasswordService struct {
	passwordResets PasswordResetRepository
	credentials    CredentialRepository
	hasher         PasswordHasher
	clock          Clock
	tx             Transactor
}

func NewResetPasswordService(
	passwordResets PasswordResetRepository,
	credentials CredentialRepository,
	hasher PasswordHasher,
	clock Clock,
	tx Transactor,
) *ResetPasswordService {
	return &ResetPasswordService{
		passwordResets: passwordResets,
		credentials:    credentials,
		hasher:         hasher,
		clock:          clock,
		tx:             tx,
	}
}

func (s *ResetPasswordService) ResetPassword(ctx context.Context, input ResetPasswordInput) (ResetPasswordResult, error) {
	if input.ResetID == "" || input.Token == "" || input.NewPassword == "" {
		return ResetPasswordResult{}, apperror.New(apperror.CodeValidation, "invalid reset password input")
	}

	foundReset, err := s.passwordResets.FindByID(ctx, input.ResetID)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return ResetPasswordResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid password reset")
	case err != nil:
		return ResetPasswordResult{}, translateRepositoryError(err)
	}

	now := s.clock.Now()
	if !foundReset.IsPending(now) {
		if foundReset.IsUsed() {
			return ResetPasswordResult{}, apperror.New(apperror.CodeConflict, "password reset already used")
		}
		if foundReset.IsExpired(now) {
			return ResetPasswordResult{}, apperror.New(apperror.CodeConflict, "password reset expired")
		}
		return ResetPasswordResult{}, apperror.New(apperror.CodeConflict, "password reset not usable")
	}

	suppliedHash := hashPasswordResetToken(input.Token)
	if subtle.ConstantTimeCompare([]byte(suppliedHash), []byte(foundReset.TokenHash())) != 1 {
		return ResetPasswordResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid password reset")
	}

	foundCredential, err := s.credentials.FindByUserID(ctx, foundReset.UserID())
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return ResetPasswordResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid password reset")
	case err != nil:
		return ResetPasswordResult{}, translateRepositoryError(err)
	}

	newHash, err := s.hasher.Hash(ctx, input.NewPassword)
	if err != nil {
		return ResetPasswordResult{}, apperror.Wrap(apperror.CodeDependencyFailure, "password hashing failed", err)
	}

	updatedCredential, err := foundCredential.ChangePassword(newHash)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredential) {
			return ResetPasswordResult{}, apperror.Wrap(apperror.CodeValidation, "invalid credential state", err)
		}
		return ResetPasswordResult{}, apperror.Wrap(apperror.CodeInternal, "internal error", err)
	}

	consumed, err := foundReset.Consume(now)
	if err != nil {
		return ResetPasswordResult{}, translatePasswordResetDomainError(err)
	}

	persist := func(ctx context.Context) error {
		if err := s.credentials.Update(ctx, updatedCredential); err != nil {
			return err
		}
		return s.passwordResets.MarkUsed(ctx, consumed)
	}

	if s.tx != nil {
		if err := s.tx.WithinTransaction(ctx, persist); err != nil {
			return ResetPasswordResult{}, translateRepositoryError(err)
		}
	} else {
		if err := persist(ctx); err != nil {
			return ResetPasswordResult{}, translateRepositoryError(err)
		}
	}

	return ResetPasswordResult{UserID: foundReset.UserID()}, nil
}
