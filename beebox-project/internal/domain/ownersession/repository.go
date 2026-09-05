package ownersession

import "context"

type Repository interface {
	Save(ctx context.Context, s Session) error
	FindByTokenHash(ctx context.Context, hash string) (Session, error)
}