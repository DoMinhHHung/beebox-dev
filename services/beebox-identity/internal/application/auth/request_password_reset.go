package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
)

const (
	defaultPasswordResetTTL = time.Hour
	passwordResetTokenBytes = 32
)

type RequestPasswordResetInput struct {
	Identifier string
}

type RequestPasswordResetResult struct {
	ResetID   string
	Token     string
	ExpiresAt time.Time
}

type RequestPasswordResetService struct {
	users          UserRepository
	passwordResets PasswordResetRepository
	clock          Clock
	ttl            time.Duration
	randReader     io.Reader
}

func NewRequestPasswordResetService(users UserRepository, passwordResets PasswordResetRepository, clock Clock) *RequestPasswordResetService {
	return &RequestPasswordResetService{
		users:          users,
		passwordResets: passwordResets,
		clock:          clock,
		ttl:            defaultPasswordResetTTL,
		randReader:     rand.Reader,
	}
}

func (s *RequestPasswordResetService) RequestPasswordReset(ctx context.Context, input RequestPasswordResetInput) (RequestPasswordResetResult, error) {
	identifier, err := identity.NewIdentifier(input.Identifier)
	if err != nil {
		return RequestPasswordResetResult{}, apperror.Wrap(apperror.CodeValidation, "invalid request password reset input", err)
	}

	foundUser, err := s.users.FindByIdentifier(ctx, identifier)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return RequestPasswordResetResult{}, apperror.New(apperror.CodeNotFound, "identity not found")
	case err != nil:
		return RequestPasswordResetResult{}, translateRepositoryError(err)
	}

	token, err := generatePasswordResetToken(s.randReader)
	if err != nil {
		return RequestPasswordResetResult{}, apperror.Wrap(apperror.CodeInternal, "password reset token generation failed", err)
	}
	tokenHash := hashPasswordResetToken(token)

	id, err := generatePasswordResetID(s.randReader)
	if err != nil {
		return RequestPasswordResetResult{}, apperror.Wrap(apperror.CodeInternal, "password reset id generation failed", err)
	}

	now := s.clock.Now()
	expiresAt := now.Add(s.ttl)
	reset, err := passwordreset.New(id, foundUser.ID(), tokenHash, now, expiresAt)
	if err != nil {
		return RequestPasswordResetResult{}, translatePasswordResetDomainError(err)
	}

	if err := s.passwordResets.Create(ctx, reset); err != nil {
		return RequestPasswordResetResult{}, translateRepositoryError(err)
	}

	return RequestPasswordResetResult{
		ResetID:   reset.ID(),
		Token:     token,
		ExpiresAt: reset.ExpiresAt(),
	}, nil
}

func generatePasswordResetToken(r io.Reader) (string, error) {
	buf := make([]byte, passwordResetTokenBytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func generatePasswordResetID(r io.Reader) (string, error) {
	buf := make([]byte, 16)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashPasswordResetToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func translatePasswordResetDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidPasswordReset):
		return apperror.Wrap(apperror.CodeValidation, "invalid password reset", err)
	case errors.Is(err, domain.ErrPasswordResetUsed):
		return apperror.New(apperror.CodeConflict, "password reset already used")
	case errors.Is(err, domain.ErrPasswordResetExpired):
		return apperror.New(apperror.CodeConflict, "password reset expired")
	default:
		return apperror.Wrap(apperror.CodeInternal, "internal error", err)
	}
}
