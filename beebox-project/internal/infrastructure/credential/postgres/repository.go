package postgres

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

var _ credential.Repository = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, c credential.Credential) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO credentials (id, project_id, environment, public_key, secret_hash, status, created_at, revoked_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (id) DO UPDATE SET
			secret_hash = EXCLUDED.secret_hash,
			status = EXCLUDED.status,
			revoked_at = EXCLUDED.revoked_at
	`, c.ID, c.ProjectID, string(c.Environment), c.PublicKey, c.SecretHash, string(c.Status), c.CreatedAt, c.RevokedAt)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to save credential", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (credential.Credential, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, project_id, environment, public_key, secret_hash, status, created_at, revoked_at
		FROM credentials
		WHERE id = $1
	`, id)
	return scanCredential(row)
}

func (r *Repository) FindByPublicKey(ctx context.Context, publicKey string) (credential.Credential, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, project_id, environment, public_key, secret_hash, status, created_at, revoked_at
		FROM credentials
		WHERE public_key = $1
	`, publicKey)
	return scanCredential(row)
}

func scanCredential(row pgx.Row) (credential.Credential, error) {
	var c credential.Credential
	var env, status string
	if err := row.Scan(&c.ID, &c.ProjectID, &env, &c.PublicKey, &c.SecretHash, &status, &c.CreatedAt, &c.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return credential.Credential{}, apperror.New(apperror.CodeNotFound, "credential not found")
		}
		return credential.Credential{}, apperror.Wrap(apperror.CodeInternal, "failed to find credential", err)
	}
	c.Environment = credential.Environment(env)
	c.Status = credential.Status(status)
	return c, nil
}