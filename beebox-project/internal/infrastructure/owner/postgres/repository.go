package postgres

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/owner"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

var _ owner.Repository = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, o owner.Owner) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO owners (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			password_hash = EXCLUDED.password_hash
	`, o.ID, o.Email, o.PasswordHash, o.CreatedAt)
	if err != nil {
		if infrapostgres.IsUniqueViolation(err) {
			return apperror.New(apperror.CodeConflict, "email already registered")
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to save owner", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (owner.Owner, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM owners
		WHERE id = $1
	`, id)
	return scanOwner(row)
}

func (r *Repository) FindByEmail(ctx context.Context, email string) (owner.Owner, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, email, password_hash, created_at
		FROM owners
		WHERE email = $1
	`, email)
	return scanOwner(row)
}

func scanOwner(row pgx.Row) (owner.Owner, error) {
	var o owner.Owner
	if err := row.Scan(&o.ID, &o.Email, &o.PasswordHash, &o.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return owner.Owner{}, apperror.New(apperror.CodeNotFound, "owner not found")
		}
		return owner.Owner{}, apperror.Wrap(apperror.CodeInternal, "failed to find owner", err)
	}
	return o, nil
}
