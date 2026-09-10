package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

type ProjectRepository struct {
	pool *pgxpool.Pool
}

func NewProjectRepository(pool *pgxpool.Pool) *ProjectRepository {
	return &ProjectRepository{pool: pool}
}

var _ project.Repository = (*ProjectRepository)(nil)

func (r *ProjectRepository) Create(ctx context.Context, p domainproject.Project) error {
	const query = `
		INSERT INTO projects (id, organization_id, status, revision)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.pool.Exec(ctx, query, p.ID, p.OrganizationID, string(p.Status), p.Revision)
	if err != nil {
		if isUniqueViolation(err) {
			return project.ErrProjectAlreadyExists
		}
		return err
	}

	return nil
}

func (r *ProjectRepository) Get(ctx context.Context, id string) (domainproject.Project, error) {
	const query = `
		SELECT id, organization_id, status, revision
		FROM projects
		WHERE id = $1
	`

	var p domainproject.Project
	var status string

	err := r.pool.QueryRow(ctx, query, id).Scan(&p.ID, &p.OrganizationID, &status, &p.Revision)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domainproject.Project{}, project.ErrProjectNotFound
		}
		return domainproject.Project{}, err
	}

	p.Status = domainproject.Status(status)
	return p, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p domainproject.Project) error {
	expectedRevision := p.Revision - 1
	const query = `
		UPDATE projects
		SET status = $2, revision = $3, updated_at = now()
		WHERE id = $1 AND revision = $4
	`

	tag, err := r.pool.Exec(ctx, query, p.ID, string(p.Status), p.Revision, expectedRevision)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		_, getErr := r.Get(ctx, p.ID)
		if errors.Is(getErr, project.ErrProjectNotFound) {
			return project.ErrProjectNotFound
		}
		if getErr != nil {
			return getErr
		}
		return project.ErrProjectConflict
	}

	return nil
}

func isUniqueViolation(err error) bool {
	var pgErr interface{ SQLState() string }
	if errors.As(err, &pgErr) {
		return pgErr.SQLState() == "23505"
	}
	return false
}
