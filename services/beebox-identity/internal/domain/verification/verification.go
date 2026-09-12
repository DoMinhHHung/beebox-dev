package verification

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type Type string

const (
	TypeEmail Type = "email"
	TypePhone Type = "phone"
)

func (t Type) Valid() bool {
	switch t {
	case TypeEmail, TypePhone:
		return true
	default:
		return false
	}
}

type Verification struct {
	id        string
	userID    identity.Identifier
	vtype     Type
	target    string
	codeHash  string
	createdAt time.Time
	expiresAt time.Time
	usedAt    *time.Time
}

func New(id string, userID identity.Identifier, vtype Type, target, codeHash string, createdAt, expiresAt time.Time) (Verification, error) {
	if id == "" || userID.IsZero() || !vtype.Valid() || target == "" || codeHash == "" || createdAt.IsZero() || expiresAt.IsZero() || !expiresAt.After(createdAt) {
		return Verification{}, domain.ErrInvalidVerification
	}
	return Verification{
		id:        id,
		userID:    userID,
		vtype:     vtype,
		target:    target,
		codeHash:  codeHash,
		createdAt: createdAt,
		expiresAt: expiresAt,
	}, nil
}

func (v Verification) ID() string {
	return v.id
}

func (v Verification) UserID() identity.Identifier {
	return v.userID
}

func (v Verification) Type() Type {
	return v.vtype
}

func (v Verification) Target() string {
	return v.target
}

func (v Verification) CodeHash() string {
	return v.codeHash
}

func (v Verification) CreatedAt() time.Time {
	return v.createdAt
}

func (v Verification) ExpiresAt() time.Time {
	return v.expiresAt
}

func (v Verification) IsUsed() bool {
	return v.usedAt != nil
}

func (v Verification) UsedAt() (time.Time, bool) {
	if v.usedAt == nil {
		return time.Time{}, false
	}
	return *v.usedAt, true
}

func (v Verification) IsExpired(at time.Time) bool {
	return !at.Before(v.expiresAt)
}

func (v Verification) IsPending(at time.Time) bool {
	return !v.IsUsed() && !v.IsExpired(at)
}

func (v Verification) Consume(at time.Time) (Verification, error) {
	if at.IsZero() || at.Before(v.createdAt) {
		return Verification{}, domain.ErrInvalidVerification
	}
	if v.IsUsed() {
		return Verification{}, domain.ErrVerificationUsed
	}
	if v.IsExpired(at) {
		return Verification{}, domain.ErrVerificationExpired
	}
	v.usedAt = &at
	return v, nil
}
