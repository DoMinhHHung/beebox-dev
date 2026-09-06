package auth

import (
	"context"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/owner"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/ownersession"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/tokengen"
)

type Service struct {
	ownerRepo   owner.Repository
	sessionRepo ownersession.Repository
	sessionTTL  time.Duration
}

func NewService(ownerRepo owner.Repository, sessionRepo ownersession.Repository, sessionTTL time.Duration) *Service {
	return &Service{ownerRepo: ownerRepo, sessionRepo: sessionRepo, sessionTTL: sessionTTL}
}

func (s *Service) SignUp(ctx context.Context, email, password string) (owner.Owner, error) {
	id, err := idgen.New()
	if err != nil {
		return owner.Owner{}, apperror.Wrap(apperror.CodeInternal, "failed to generate owner id", err)
	}

	o, err := owner.NewOwner(id, email, password, time.Now())
	if err != nil {
		return owner.Owner{}, err
	}

	if err := s.ownerRepo.Save(ctx, o); err != nil {
		return owner.Owner{}, err
	}

	return o, nil
}

func (s *Service) SignIn(ctx context.Context, email, password string) (string, error) {
	o, err := s.ownerRepo.FindByEmail(ctx, email)
	if err != nil {
		if apperror.CodeOf(err) == apperror.CodeNotFound {
			return "", apperror.New(apperror.CodeCredentialInvalid, "email or password is incorrect")
		}
		return "", err
	}

	if err := o.VerifyPassword(password); err != nil {
		return "", err
	}

	session, token, err := ownersession.NewSession(o.ID, s.sessionTTL, time.Now())
	if err != nil {
		return "", err
	}

	if err := s.sessionRepo.Save(ctx, session); err != nil {
		return "", err
	}

	return token, nil
}

func (s *Service) VerifySession(ctx context.Context, rawToken string) (string, error) {
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokengen.Hash(rawToken))
	if err != nil {
		if apperror.CodeOf(err) == apperror.CodeNotFound {
			return "", apperror.New(apperror.CodeCredentialInvalid, "session is invalid")
		}
		return "", err
	}

	if err := session.Verify(time.Now()); err != nil {
		return "", err
	}

	return session.OwnerID, nil
}

func (s *Service) SignOut(ctx context.Context, rawToken string) error {
	session, err := s.sessionRepo.FindByTokenHash(ctx, tokengen.Hash(rawToken))
	if err != nil {
		if apperror.CodeOf(err) == apperror.CodeNotFound {
			return nil
		}
		return err
	}

	revoked, err := session.Revoke(time.Now())
	if err != nil {
		if apperror.CodeOf(err) == apperror.CodeCredentialRevoked {
			return nil
		}
		return err
	}

	return s.sessionRepo.Save(ctx, revoked)
}
