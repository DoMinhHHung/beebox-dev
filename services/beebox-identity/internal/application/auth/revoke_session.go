package auth

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
)

type RevokeSessionInput struct {
	SessionID string
}

type RevokeSessionService struct {
	sessions SessionRepository
	clock    Clock
}

func NewRevokeSessionService(sessions SessionRepository, clock Clock) *RevokeSessionService {
	return &RevokeSessionService{
		sessions: sessions,
		clock:    clock,
	}
}

func (s *RevokeSessionService) RevokeSession(ctx context.Context, input RevokeSessionInput) error {
	if input.SessionID == "" {
		return apperror.New(apperror.CodeValidation, "invalid revoke session input")
	}

	foundSession, err := s.sessions.FindByID(ctx, input.SessionID)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return apperror.New(apperror.CodeUnauthenticated, "invalid session")
	case err != nil:
		return translateRepositoryError(err)
	}

	revoked, err := foundSession.Revoke(s.clock.Now())
	switch {
	case err == nil:
	case errors.Is(err, domain.ErrSessionRevoked):
		return apperror.New(apperror.CodeConflict, "session already revoked")
	case errors.Is(err, domain.ErrInvalidSession):
		return apperror.Wrap(apperror.CodeValidation, "invalid session state", err)
	case err != nil:
		return apperror.Wrap(apperror.CodeInternal, "internal error", err)
	}

	if err := s.sessions.Revoke(ctx, revoked); err != nil {
		return translateRepositoryError(err)
	}

	return nil
}
