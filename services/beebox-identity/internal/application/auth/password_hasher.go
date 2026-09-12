package auth

import "context"

type PasswordHasher interface {
	Hash(ctx context.Context, plaintext string) (string, error)
	Verify(ctx context.Context, plaintext, passwordHash string) error
}
