package auth

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
)

type PasswordResetRepository interface {
	Create(ctx context.Context, value passwordreset.PasswordReset) error
	FindPending(ctx context.Context, userID identity.Identifier) (passwordreset.PasswordReset, error)
	MarkUsed(ctx context.Context, value passwordreset.PasswordReset) error
}
