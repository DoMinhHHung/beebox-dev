package credential

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type Credential struct {
	userID       identity.Identifier
	passwordHash string
	createdAt    time.Time
	revokedAt    *time.Time
}

func NewPassword(userID identity.Identifier, passwordHash string, createdAt time.Time) (Credential, error) {
	if userID.IsZero() || passwordHash == "" || createdAt.IsZero() {
		return Credential{}, domain.ErrInvalidCredential
	}
	return Credential{
		userID:       userID,
		passwordHash: passwordHash,
		createdAt:    createdAt,
	}, nil
}

func (p Credential) UserID() identity.Identifier {
	return p.userID
}

func (p Credential) PasswordHash() string {
	return p.passwordHash
}

func (p Credential) CreatedAt() time.Time {
	return p.createdAt
}

func (p Credential) IsRevoked() bool {
	return p.revokedAt != nil
}

func (p Credential) Revoke(at time.Time) (Credential, error) {
	if at.IsZero() || at.Before(p.createdAt) {
		return Credential{}, domain.ErrInvalidCredential
	}
	if p.IsRevoked() {
		return Credential{}, domain.ErrInvalidCredential
	}
	p.revokedAt = &at
	return p, nil
}

func (p Credential) RevokedAt() (time.Time, bool) {
	if p.revokedAt == nil {
		return time.Time{}, false
	}
	return *p.revokedAt, true
}

func (p Credential) ChangePassword(newPasswordHash string) (Credential, error) {
	if newPasswordHash == "" || p.IsRevoked() {
		return Credential{}, domain.ErrInvalidCredential
	}
	p.passwordHash = newPasswordHash
	return p, nil
}
