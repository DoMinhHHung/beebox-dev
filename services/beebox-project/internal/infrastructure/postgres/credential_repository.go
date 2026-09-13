package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	applicationcredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/credential"
	domaincredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/credential"
)

type CredentialRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

var _ applicationcredential.Repository = (*CredentialRepository)(nil)

func (r *CredentialRepository) Create(ctx context.Context, credential domaincredential.Credential) error {
	const query = `
		INSERT INTO project_credentials (project_id, id, name, kind, secret_hash, status, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.pool.Exec(
		ctx,
		query,
		credential.ProjectID,
		credential.ID,
		credential.Name,
		string(credential.Kind),
		credential.SecretHash,
		string(credential.Status),
		credential.ExpiresAt,
	)
	if err != nil && isUniqueViolation(err) {
		return applicationcredential.ErrCredentialAlreadyExists
	}
	return err
}

func (r *CredentialRepository) FindByID(ctx context.Context, projectID, id string) (domaincredential.Credential, error) {
	const query = `
		SELECT project_id, id, name, kind, secret_hash, status, expires_at
		FROM project_credentials
		WHERE project_id = $1 AND id = $2
	`
	return r.scanOne(ctx, query, projectID, id)
}

func (r *CredentialRepository) FindByProjectAndHash(ctx context.Context, projectID string, kind domaincredential.Kind, secretHash string) (domaincredential.Credential, error) {
	const query = `
		SELECT project_id, id, name, kind, secret_hash, status, expires_at
		FROM project_credentials
		WHERE project_id = $1 AND kind = $2 AND secret_hash = $3
	`
	return r.scanOne(ctx, query, projectID, string(kind), secretHash)
}

func (r *CredentialRepository) Update(ctx context.Context, credential domaincredential.Credential) error {
	const query = `
		UPDATE project_credentials
		SET name = $3, kind = $4, secret_hash = $5, status = $6, expires_at = $7
		WHERE project_id = $1 AND id = $2
	`
	tag, err := r.pool.Exec(
		ctx,
		query,
		credential.ProjectID,
		credential.ID,
		credential.Name,
		string(credential.Kind),
		credential.SecretHash,
		string(credential.Status),
		credential.ExpiresAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return applicationcredential.ErrCredentialNotFound
	}
	return nil
}

func (r *CredentialRepository) scanOne(ctx context.Context, query string, args ...any) (domaincredential.Credential, error) {
	var (
		projectID  string
		id         string
		name       string
		kind       string
		secretHash string
		status     string
		expiresAt  *time.Time
	)
	err := r.pool.QueryRow(ctx, query, args...).Scan(&projectID, &id, &name, &kind, &secretHash, &status, &expiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domaincredential.Credential{}, applicationcredential.ErrCredentialNotFound
	}
	if err != nil {
		return domaincredential.Credential{}, err
	}
	return domaincredential.Credential{
		ProjectID:  projectID,
		ID:         id,
		Name:       name,
		Kind:       domaincredential.Kind(kind),
		SecretHash: secretHash,
		Status:     domaincredential.Status(status),
		ExpiresAt:  expiresAt,
	}, nil
}
