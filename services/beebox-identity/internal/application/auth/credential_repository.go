package auth

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type CredentialRepository interface {
	Create(ctx context.Context, value credential.Credential) error
	FindByUserID(ctx context.Context, userID identity.Identifier) (credential.Credential, error)
}
