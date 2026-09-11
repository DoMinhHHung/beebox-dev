package auth

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type SignInInput struct {
	Identifier string
	Password   string
}

type SignInResult struct {
	UserID identity.Identifier
}

type SignInService struct {
	users       UserRepository
	credentials CredentialRepository
	hasher      PasswordHasher
}

func NewSignInService(users UserRepository, credentials CredentialRepository, hasher PasswordHasher) *SignInService {
	return &SignInService{
		users:       users,
		credentials: credentials,
		hasher:      hasher,
	}
}

func (s *SignInService) SignIn(ctx context.Context, input SignInInput) (SignInResult, error) {
	if input.Password == "" {
		return SignInResult{}, apperror.New(apperror.CodeValidation, "invalid signin input")
	}

	identifier, err := identity.NewIdentifier(input.Identifier)
	if err != nil {
		return SignInResult{}, apperror.Wrap(apperror.CodeValidation, "invalid signin input", err)
	}

	foundUser, err := s.users.FindByIdentifier(ctx, identifier)
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	case err != nil:
		return SignInResult{}, translateRepositoryError(err)
	}

	foundCredential, err := s.credentials.FindByUserID(ctx, foundUser.ID())
	switch {
	case err == nil:
	case errors.Is(err, ErrNotFound):
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	case err != nil:
		return SignInResult{}, translateRepositoryError(err)
	}

	if err := s.hasher.Verify(ctx, input.Password, foundCredential.PasswordHash()); err != nil {
		return SignInResult{}, apperror.New(apperror.CodeUnauthenticated, "invalid credentials")
	}

	return SignInResult{UserID: foundUser.ID()}, nil
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
