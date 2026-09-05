package credential

import "context"

type Repository interface {
	Save(ctx context.Context, c Credential) error
	FindByID(ctx context.Context, id string) (Credential, error)
	FindByPublicKey(ctx context.Context, publicKey string) (Credential, error)
}