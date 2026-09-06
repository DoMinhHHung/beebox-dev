package ownersession

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/tokengen"
)

type Session struct {
	ID        string
	OwnerID   string
	TokenHash string
	CreatedAt time.Time
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func NewSession(ownerID string, ttl time.Duration, createdAt time.Time) (Session, string, error) {
	if ownerID == "" {
		return Session{}, "", apperror.New(apperror.CodeInvalidInput, "owner id must not be empty")
	}
	if ttl <= 0 {
		return Session{}, "", apperror.New(apperror.CodeInvalidInput, "ttl must be greater than zero")
	}

	id, err := idgen.New()
	if err != nil {
		return Session{}, "", apperror.Wrap(apperror.CodeInternal, "failed to generate session id", err)
	}

	token, err := tokengen.New(32)
	if err != nil {
		return Session{}, "", err
	}

	s := Session{
		ID:        id,
		OwnerID:   ownerID,
		TokenHash: tokengen.Hash(token),
		CreatedAt: createdAt,
		ExpiresAt: createdAt.Add(ttl),
	}

	return s, token, nil
}

func (s Session) Verify(now time.Time) error {
	if s.RevokedAt != nil {
		return apperror.New(apperror.CodeCredentialRevoked, "session has been revoked")
	}
	if now.After(s.ExpiresAt) {
		return apperror.New(apperror.CodeCredentialInvalid, "session has expired")
	}
	return nil
}

func (s Session) Revoke(now time.Time) (Session, error) {
	if s.RevokedAt != nil {
		return Session{}, apperror.New(apperror.CodeCredentialRevoked, "session already revoked")
	}
	revoked := s
	revoked.RevokedAt = &now
	return revoked, nil
}
