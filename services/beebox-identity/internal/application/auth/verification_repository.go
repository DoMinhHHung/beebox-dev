package auth

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

type VerificationRepository interface {
	Create(ctx context.Context, value verification.Verification) error
	FindPending(ctx context.Context, userID identity.Identifier, vtype verification.Type, target string) (verification.Verification, error)
	MarkUsed(ctx context.Context, value verification.Verification) error
}
