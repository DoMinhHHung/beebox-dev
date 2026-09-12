package user

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type User struct {
	id        identity.Identifier
	createdAt time.Time
}

func New(id identity.Identifier, createdAt time.Time) (User, error) {
	if id.IsZero() || createdAt.IsZero() {
		return User{}, domain.ErrInvalidUser
	}
	return User{
		id:        id,
		createdAt: createdAt,
	}, nil
}

func (u User) ID() identity.Identifier {
	return u.id
}

func (u User) CreatedAt() time.Time {
	return u.createdAt
}
