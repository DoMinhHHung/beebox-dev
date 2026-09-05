package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

var _ fielddefinition.Repository = (*Repository)(nil)

func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, s fielddefinition.Schema) error {
	fieldsJSON, err := json.Marshal(s.Fields)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to marshal schema fields", err)
	}

	_, err = r.pool.Exec(ctx, `
		INSERT INTO field_definition_schemas (project_id, version, fields)
		VALUES ($1, $2, $3)
	`, s.ProjectID, s.Version, string(fieldsJSON))
	if err != nil {
		if infrapostgres.IsUniqueViolation(err) {
			return apperror.New(apperror.CodeConflict, "schema version already exists")
		}
		return apperror.Wrap(apperror.CodeInternal, "failed to save schema", err)
	}
	return nil
}

func (r *Repository) FindLatest(ctx context.Context, projectID string) (fielddefinition.Schema, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT project_id, version, fields
		FROM field_definition_schemas
		WHERE project_id = $1
		ORDER BY version DESC
		LIMIT 1
	`, projectID)
	return scanSchema(row)
}

func (r *Repository) FindVersion(ctx context.Context, projectID string, version int) (fielddefinition.Schema, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT project_id, version, fields
		FROM field_definition_schemas
		WHERE project_id = $1 AND version = $2
	`, projectID, version)
	return scanSchema(row)
}

func scanSchema(row pgx.Row) (fielddefinition.Schema, error) {
	var s fielddefinition.Schema
	var fieldsJSON []byte
	if err := row.Scan(&s.ProjectID, &s.Version, &fieldsJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fielddefinition.Schema{}, apperror.New(apperror.CodeNotFound, "schema not found")
		}
		return fielddefinition.Schema{}, apperror.Wrap(apperror.CodeInternal, "failed to find schema", err)
	}
	if err := json.Unmarshal(fieldsJSON, &s.Fields); err != nil {
		return fielddefinition.Schema{}, apperror.Wrap(apperror.CodeInternal, "failed to unmarshal schema fields", err)
	}
	return s, nil
}