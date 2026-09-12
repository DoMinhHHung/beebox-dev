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
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
)

const (
	defaultSessionTTL = 24 * time.Hour
	sessionTokenBytes = 32
)

type SignInInput struct {
	Identifier string
	Password   string
}

type SignInResult struct {
	UserID    identity.Identifier
	SessionID string
	ExpiresAt time.Time
}

type SignInService struct {
	users       UserRepository
	credentials CredentialRepository
	sessions    SessionRepository
	hasher      PasswordHasher
	clock       Clock
	ttl         time.Duration
	randReader  io.Reader
}

func NewSignInService(
	users UserRepository,
	credentials CredentialRepository,
	sessions SessionRepository,
	hasher PasswordHasher,
	clock Clock,
) *SignInService {
	return &SignInService{
		users:       users,
		credentials: credentials,
		sessions:    sessions,
		hasher:      hasher,
		clock:       clock,
		ttl:         defaultSessionTTL,
		randReader:  rand.Reader,
	}
}

func (s *SignInService) SignIn(ctx context.Context, input SignInInput) (SignInResult, error) {
	if input.Password == "" {
		return SignInResult{}, apperror.New(apperror.CodeValidation, "invalid signin input")
	}

	identifier, err := identity.NewIdentifier(input.Identifier)
	if err != nil {
		return SignInResult{}, apperror.Wrap(apperror.CodeValidation, "invalid signin input", err)
	}

	foundUser, err := s.users.FindByIdentifier(ctx, identifier)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	case err != nil:
		return SignInResult{}, translateRepositoryError(err)
	}

	foundCredential, err := s.credentials.FindByUserID(ctx, foundUser.ID())
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	case err != nil:
		return SignInResult{}, translateRepositoryError(err)
	}

	if foundCredential.IsRevoked() {
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	}

	if err := s.hasher.Verify(ctx, input.Password, foundCredential.PasswordHash()); err != nil {
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	}

	token, err := generateSessionToken(s.randReader)
	if err != nil {
		return SignInResult{}, apperror.Wrap(apperror.CodeInternal, "session token generation failed", err)
	}
	storageID := hashSessionToken(token)

	now := s.clock.Now()
	expiresAt := now.Add(s.ttl)
	newSession, err := session.New(storageID, foundUser.ID(), now, expiresAt)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidSession) {
			return SignInResult{}, apperror.Wrap(apperror.CodeInternal, "invalid session state", err)
		}
		return SignInResult{}, apperror.Wrap(apperror.CodeInternal, "internal error", err)
	}

	if err := s.sessions.Create(ctx, newSession); err != nil {
		return SignInResult{}, translateRepositoryError(err)
	}

	return SignInResult{
		UserID:    foundUser.ID(),
		SessionID: token,
		ExpiresAt: newSession.ExpiresAt(),
	}, nil
}

func generateSessionToken(r io.Reader) (string, error) {
	buf := make([]byte, sessionTokenBytes)
	if _, err := io.ReadFull(r, buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func hashSessionToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
