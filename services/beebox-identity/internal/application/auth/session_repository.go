package auth

import (
	"context"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
)

type SessionRepository interface {
	Create(ctx context.Context, value session.Session) error
	FindByID(ctx context.Context, id string) (session.Session, error)
	Revoke(ctx context.Context, value session.Session) error
	RevokeAllByUserID(ctx context.Context, userID identity.Identifier, revokedAt time.Time) error
}
