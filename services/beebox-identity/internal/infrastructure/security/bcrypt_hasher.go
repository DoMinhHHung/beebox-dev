package security

import (
	"context"

	"golang.org/x/crypto/bcrypt"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(_ context.Context, plaintext string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), h.cost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

func (h *BcryptPasswordHasher) Verify(_ context.Context, plaintext, passwordHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plaintext))
}

var _ auth.PasswordHasher = (*BcryptPasswordHasher)(nil)
