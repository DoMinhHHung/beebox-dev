package postgres

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/ownersession"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

var _ ownersession.Repository = (*Repository)(nil)

// New tạo một repository sử dụng connection pool PostgreSQL được cung cấp.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, s ownersession.Session) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO owner_sessions (id, owner_id, token_hash, created_at, expires_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (id) DO UPDATE SET
			revoked_at = EXCLUDED.revoked_at
	`, s.ID, s.OwnerID, s.TokenHash, s.CreatedAt, s.ExpiresAt, s.RevokedAt)
	if err != nil {
		if infrapostgres.IsUniqueViolation(err) {
			return apperror.New(apperror.CodeConflict, "session token collision")
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to save session", err)
	}
	return nil
}

func (r *Repository) FindByTokenHash(ctx context.Context, hash string) (ownersession.Session, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, token_hash, created_at, expires_at, revoked_at
		FROM owner_sessions
		WHERE token_hash = $1
	`, hash)

	var s ownersession.Session
	if err := row.Scan(&s.ID, &s.OwnerID, &s.TokenHash, &s.CreatedAt, &s.ExpiresAt, &s.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ownersession.Session{}, apperror.New(apperror.CodeNotFound, "session not found")
		}
		return ownersession.Session{}, apperror.Wrap(apperror.CodeInternal, "failed to find session", err)
	}
	return s, nil
}
