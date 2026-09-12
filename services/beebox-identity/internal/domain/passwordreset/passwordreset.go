package passwordreset

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type PasswordReset struct {
	id        string
	userID    identity.Identifier
	tokenHash string
	createdAt time.Time
	expiresAt time.Time
	usedAt    *time.Time
}

func New(id string, userID identity.Identifier, tokenHash string, createdAt, expiresAt time.Time) (PasswordReset, error) {
	if id == "" || userID.IsZero() || tokenHash == "" || createdAt.IsZero() || expiresAt.IsZero() || !expiresAt.After(createdAt) {
		return PasswordReset{}, domain.ErrInvalidPasswordReset
	}
	return PasswordReset{
		id:        id,
		userID:    userID,
		tokenHash: tokenHash,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}, nil
}

func (p PasswordReset) ID() string {
	return p.id
}

func (p PasswordReset) UserID() identity.Identifier {
	return p.userID
}

func (p PasswordReset) TokenHash() string {
	return p.tokenHash
}

func (p PasswordReset) CreatedAt() time.Time {
	return p.createdAt
}

func (p PasswordReset) ExpiresAt() time.Time {
	return p.expiresAt
}

func (p PasswordReset) IsUsed() bool {
	return p.usedAt != nil
}

func (p PasswordReset) UsedAt() (time.Time, bool) {
	if p.usedAt == nil {
		return time.Time{}, false
	}
	return *p.usedAt, true
}

func (p PasswordReset) IsExpired(at time.Time) bool {
	return !at.Before(p.expiresAt)
}

func (p PasswordReset) IsPending(at time.Time) bool {
	return !p.IsUsed() && !p.IsExpired(at)
}

func (p PasswordReset) Consume(at time.Time) (PasswordReset, error) {
	if at.IsZero() || at.Before(p.createdAt) {
		return PasswordReset{}, domain.ErrInvalidPasswordReset
	}
	if p.IsUsed() {
		return PasswordReset{}, domain.ErrPasswordResetUsed
	}
	if p.IsExpired(at) {
		return PasswordReset{}, domain.ErrPasswordResetExpired
	}
	p.usedAt = &at
	return p, nil
}
