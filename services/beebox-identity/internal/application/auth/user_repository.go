package auth

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
)

type UserRepository interface {
	Create(ctx context.Context, value user.User) error
	FindByIdentifier(ctx context.Context, identifier identity.Identifier) (user.User, error)
}
