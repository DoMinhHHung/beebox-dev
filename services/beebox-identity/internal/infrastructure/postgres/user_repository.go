package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) Create(ctx context.Context, value user.User) error {
	q := querierFrom(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO users (id, created_at) VALUES ($1, $2)`,
		value.ID().String(), value.CreatedAt().UTC(),
	)
	return err
}

func (r *UserRepository) FindByIdentifier(ctx context.Context, identifier identity.Identifier) (user.User, error) {
	q := querierFrom(ctx, r.pool)
	var id string
	var createdAt time.Time
	err := q.QueryRow(ctx,
		`SELECT id, created_at FROM users WHERE id = $1`,
		identifier.String(),
	).Scan(&id, &createdAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return user.User{}, auth.ErrNotFound
	}
	if err != nil {
		return user.User{}, err
	}
	ident, err := identity.NewIdentifier(id)
	if err != nil {
		return user.User{}, err
	}
	return user.New(ident, createdAt.UTC())
}

var _ auth.UserRepository = (*UserRepository)(nil)
