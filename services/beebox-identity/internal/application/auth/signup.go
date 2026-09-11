package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
)

type SignUpInput struct {
	Identifier string
	Password   string
}

type SignUpResult struct {
	UserID identity.Identifier
}

type SignUpService struct {
	users       UserRepository
	credentials CredentialRepository
	hasher      PasswordHasher
	clock       Clock
}

func NewSignUpService(users UserRepository, credentials CredentialRepository, hasher PasswordHasher, clock Clock) *SignUpService {
	return &SignUpService{
		users:       users,
		credentials: credentials,
		hasher:      hasher,
		clock:       clock,
	}
}

func (s *SignUpService) SignUp(ctx context.Context, input SignUpInput) (SignUpResult, error) {
	if strings.TrimSpace(input.Identifier) == "" || input.Password == "" {
		return SignUpResult{}, apperror.New(apperror.CodeValidation, "invalid signup input")
	}

	identifier, err := identity.NewIdentifier(input.Identifier)
	if err != nil {
		return SignUpResult{}, translateDomainError(err)
	}

	_, err = s.users.FindByIdentifier(ctx, identifier)
	switch {
	case err == nil:
		return SignUpResult{}, apperror.New(apperror.CodeConflict, "identity already exists")
	case errors.Is(err, ErrNotFound):
	case err != nil:
		return SignUpResult{}, translateRepositoryError(err)
	}

	passwordHash, err := s.hasher.Hash(ctx, input.Password)
	if err != nil {
		return SignUpResult{}, apperror.Wrap(apperror.CodeDependencyFailure, "password hashing failed", err)
	}

	now := s.clock.Now()
	newUser, err := user.New(identifier, now)
	if err != nil {
		return SignUpResult{}, translateDomainError(err)
	}
	newCredential, err := credential.NewPassword(identifier, passwordHash, now)
	if err != nil {
		return SignUpResult{}, translateDomainError(err)
	}

	if err := s.users.Create(ctx, newUser); err != nil {
		return SignUpResult{}, translateRepositoryError(err)
	}
	if err := s.credentials.Create(ctx, newCredential); err != nil {
		return SignUpResult{}, translateRepositoryError(err)
	}

	return SignUpResult{UserID: newUser.ID()}, nil
}

func translateDomainError(err error) error {
	if errors.Is(err, domain.ErrInvalidIdentityIdentifier) ||
		errors.Is(err, domain.ErrInvalidUser) ||
		errors.Is(err, domain.ErrInvalidCredential) {
		return apperror.Wrap(apperror.CodeValidation, "invalid signup input", err)
	}
	return apperror.Wrap(apperror.CodeInternal, "internal error", err)
}

func translateRepositoryError(err error) error {
	return apperror.Wrap(apperror.CodeDependencyFailure, "identity dependency failed", err)
}
