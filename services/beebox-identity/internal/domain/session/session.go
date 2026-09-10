package session

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type Session struct {
	id        string
	userID    identity.Identifier
	createdAt time.Time
	expiresAt time.Time
	revokedAt *time.Time
}

func New(id string, userID identity.Identifier, createdAt, expiresAt time.Time) (Session, error) {
	if id == "" || userID.IsZero() || createdAt.IsZero() || expiresAt.IsZero() || !expiresAt.After(createdAt) {
		return Session{}, domain.ErrInvalidSession
	}
	return Session{
		id:        id,
		userID:    userID,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}, nil
}

func (s Session) ID() string {
	return s.id
}

func (s Session) UserID() identity.Identifier {
	return s.userID
}

func (s Session) CreatedAt() time.Time {
	return s.createdAt
}

func (s Session) ExpiresAt() time.Time {
	return s.expiresAt
}

func (s Session) IsExpired(at time.Time) bool {
	return !at.Before(s.expiresAt)
}

func (s Session) IsRevoked() bool {
	return s.revokedAt != nil
}

func (s Session) IsActive(at time.Time) bool {
	return !at.IsZero() && !s.IsRevoked() && !s.IsExpired(at)
}

func (s Session) Revoke(at time.Time) (Session, error) {
	if at.IsZero() || at.Before(s.createdAt) {
		return Session{}, domain.ErrInvalidSession
	}
	if s.IsRevoked() {
		return Session{}, domain.ErrSessionRevoked
	}
	s.revokedAt = &at
	return s, nil
}

func (s Session) RevokedAt() (time.Time, bool) {
	if s.revokedAt == nil {
		return time.Time{}, false
	}
	return *s.revokedAt, true
}
