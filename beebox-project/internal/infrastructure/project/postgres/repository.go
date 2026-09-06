package postgres

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool *pgxpool.Pool
}

var _ project.Repository = (*Repository)(nil)

// New tạo một Repository sử dụng connection pool PostgreSQL được cung cấp.
func New(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

func (r *Repository) Save(ctx context.Context, p project.Project) error {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO projects (id, owner_id, name, tier, created_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (id) DO UPDATE SET
			owner_id = EXCLUDED.owner_id,
			name = EXCLUDED.name,
			tier = EXCLUDED.tier
	`, p.ID, p.OwnerID, p.Name, string(p.Tier), p.CreatedAt)
	if err != nil {
		return apperror.Wrap(apperror.CodeInternal, "failed to save project", err)
	}
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (project.Project, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, owner_id, name, tier, created_at
		FROM projects
		WHERE id = $1
	`, id)

	var p project.Project
	var tier string
	if err := row.Scan(&p.ID, &p.OwnerID, &p.Name, &tier, &p.CreatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return project.Project{}, apperror.New(apperror.CodeNotFound, "project not found")
		}
		return project.Project{}, apperror.Wrap(apperror.CodeInternal, "failed to find project", err)
	}
	p.Tier = project.Tier(tier)
	return p, nil
}
