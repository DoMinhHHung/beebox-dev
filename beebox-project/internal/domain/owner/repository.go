package owner

import "context"

type Repository interface {
	Save(ctx context.Context, o Owner) error
	FindByID(ctx context.Context, id string) (Owner, error)
	FindByEmail(ctx context.Context, email string) (Owner, error)
}