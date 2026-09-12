package security

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

const bcryptMaxPasswordBytes = 72

type BcryptPasswordHasher struct {
	cost int
}

func NewBcryptPasswordHasher() *BcryptPasswordHasher {
	return &BcryptPasswordHasher{cost: bcrypt.DefaultCost}
}

func (h *BcryptPasswordHasher) Hash(_ context.Context, plaintext string) (string, error) {
	if len([]byte(plaintext)) > bcryptMaxPasswordBytes {
		return "", apperror.New(apperror.CodeValidation, "password exceeds maximum length")
	}
	hashed, err := bcrypt.GenerateFromPassword([]byte(plaintext), h.cost)
	if err != nil {
		if errors.Is(err, bcrypt.ErrPasswordTooLong) {
			return "", apperror.New(apperror.CodeValidation, "password exceeds maximum length")
		}
		return "", err
	}
	return string(hashed), nil
}

func (h *BcryptPasswordHasher) Verify(_ context.Context, plaintext, passwordHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(plaintext))
}

var _ auth.PasswordHasher = (*BcryptPasswordHasher)(nil)
