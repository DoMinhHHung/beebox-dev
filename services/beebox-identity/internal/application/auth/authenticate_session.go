package auth

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type AuthenticateSessionInput struct {
	Token string
}

type AuthenticateSessionResult struct {
	UserID    identity.Identifier
	SessionID string
}

type AuthenticateSessionService struct {
	sessions SessionRepository
	clock    Clock
}

func NewAuthenticateSessionService(sessions SessionRepository, clock Clock) *AuthenticateSessionService {
	return &AuthenticateSessionService{
		sessions: sessions,
		clock:    clock,
	}
}

func (s *AuthenticateSessionService) AuthenticateSession(ctx context.Context, input AuthenticateSessionInput) (AuthenticateSessionResult, error) {
	if input.Token == "" {
		return AuthenticateSessionResult{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	storageID := hashSessionToken(input.Token)
	found, err := s.sessions.FindByID(ctx, storageID)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return AuthenticateSessionResult{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	case err != nil:
		return AuthenticateSessionResult{}, translateRepositoryError(err)
	}

	now := s.clock.Now()
	if !found.IsActive(now) {
		return AuthenticateSessionResult{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	return AuthenticateSessionResult{
		UserID:    found.UserID(),
		SessionID: found.ID(),
	}, nil
}
