package owner

import (
	"regexp"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"golang.org/x/crypto/bcrypt"
)

var emailPattern = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

const (
	minPasswordLength = 8
	maxPasswordBytes  = 72 // bcrypt input limit
)

type Owner struct {
	ID           string
	Email        string
	PasswordHash string
	CreatedAt    time.Time
}

func NewOwner(id, email, plaintextPassword string, createdAt time.Time) (Owner, error) {
	if !emailPattern.MatchString(email) {
		return Owner{}, apperror.New(apperror.CodeInvalidInput, "email must be a valid email address")
	}
	if len(plaintextPassword) < minPasswordLength {
		return Owner{}, apperror.New(apperror.CodeInvalidInput, "password must be at least 8 characters")
	}
	if len(plaintextPassword) > maxPasswordBytes {
		return Owner{}, apperror.New(apperror.CodeInvalidInput, "password must be at most 72 bytes")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(plaintextPassword), bcrypt.DefaultCost)
	if err != nil {
		return Owner{}, apperror.Wrap(apperror.CodeInternal, "failed to hash password", err)
	}

	return Owner{
		ID:           id,
		Email:        email,
		PasswordHash: string(hash),
		CreatedAt:    createdAt,
	}, nil
}

func (o Owner) VerifyPassword(plaintext string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(o.PasswordHash), []byte(plaintext)); err != nil {
		return apperror.New(apperror.CodeCredentialInvalid, "email or password is incorrect")
	}
	return nil
}
