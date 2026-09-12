package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	applicationenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/enablement"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/enablement"
)

type EnablementRepository struct {
	pool *pgxpool.Pool
}

func NewEnablementRepository(pool *pgxpool.Pool) *EnablementRepository {
	return &EnablementRepository{pool: pool}
}

var _ applicationenablement.Repository = (*EnablementRepository)(nil)

func (r *EnablementRepository) Save(ctx context.Context, item domainenablement.Enablement) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insertCapability = `
		INSERT INTO project_capabilities
		(project_id, module_id, module_version, capability_id, capability_version)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = tx.Exec(ctx, insertCapability, item.ProjectID, item.ModuleID, item.ModuleVersion, item.CapabilityID, item.CapabilityVersion)
	if err != nil {
		if isUniqueViolation(err) {
			return applicationenablement.ErrEnablementAlreadyExists
		}
		return err
	}

	if err := insertEnablementFields(ctx, tx, item); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *EnablementRepository) Update(ctx context.Context, item domainenablement.Enablement) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const check = `
		SELECT 1 FROM project_capabilities
		WHERE project_id = $1 AND module_id = $2 AND capability_id = $3
	`
	var one int
	err = tx.QueryRow(ctx, check, item.ProjectID, item.ModuleID, item.CapabilityID).Scan(&one)
	if errors.Is(err, pgx.ErrNoRows) {
		return applicationenablement.ErrEnablementNotFound
	}
	if err != nil {
		return err
	}

	const deleteFields = `
		DELETE FROM project_capability_fields
		WHERE project_id = $1 AND module_id = $2 AND capability_id = $3
	`
	if _, err := tx.Exec(ctx, deleteFields, item.ProjectID, item.ModuleID, item.CapabilityID); err != nil {
		return err
	}

	const updateCapability = `
		UPDATE project_capabilities
		SET module_version = $4, capability_version = $5
		WHERE project_id = $1 AND module_id = $2 AND capability_id = $3
	`
	if _, err := tx.Exec(ctx, updateCapability, item.ProjectID, item.ModuleID, item.CapabilityID, item.ModuleVersion, item.CapabilityVersion); err != nil {
		return err
	}

	if err := insertEnablementFields(ctx, tx, item); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *EnablementRepository) Get(ctx context.Context, projectID, moduleID, capabilityID string) (domainenablement.Enablement, error) {
	const query = `
		SELECT module_version, capability_version
		FROM project_capabilities
		WHERE project_id = $1 AND module_id = $2 AND capability_id = $3
	`
	var moduleVersion, capabilityVersion string
	err := r.pool.QueryRow(ctx, query, projectID, moduleID, capabilityID).Scan(&moduleVersion, &capabilityVersion)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainenablement.Enablement{}, applicationenablement.ErrEnablementNotFound
	}
	if err != nil {
		return domainenablement.Enablement{}, err
	}

	fields, err := r.loadEnablementFields(ctx, projectID, moduleID, capabilityID)
	if err != nil {
		return domainenablement.Enablement{}, err
	}

	item, err := domainenablement.New(projectID, moduleID, moduleVersion, capabilityID, capabilityVersion)
	if err != nil {
		return domainenablement.Enablement{}, err
	}
	return item.WithDataFields(fields)
}

func (r *EnablementRepository) ListByProject(ctx context.Context, projectID string) ([]domainenablement.Enablement, error) {
	const query = `
		SELECT module_id, module_version, capability_id, capability_version
		FROM project_capabilities
		WHERE project_id = $1
		ORDER BY module_id, capability_id
	`
	rows, err := r.pool.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []domainenablement.Enablement
	for rows.Next() {
		var moduleID, moduleVersion, capabilityID, capabilityVersion string
		if err := rows.Scan(&moduleID, &moduleVersion, &capabilityID, &capabilityVersion); err != nil {
			return nil, err
		}
		fields, err := r.loadEnablementFields(ctx, projectID, moduleID, capabilityID)
		if err != nil {
			return nil, err
		}
		item, err := domainenablement.New(projectID, moduleID, moduleVersion, capabilityID, capabilityVersion)
		if err != nil {
			return nil, err
		}
		item, err = item.WithDataFields(fields)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *EnablementRepository) Delete(ctx context.Context, projectID, moduleID, capabilityID string) error {
	const query = `
		DELETE FROM project_capabilities
		WHERE project_id = $1 AND module_id = $2 AND capability_id = $3
	`
	tag, err := r.pool.Exec(ctx, query, projectID, moduleID, capabilityID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return applicationenablement.ErrEnablementNotFound
	}
	return nil
}

func (r *EnablementRepository) loadEnablementFields(ctx context.Context, projectID, moduleID, capabilityID string) ([]configuration.DataFieldReference, error) {
	const query = `
		SELECT field_id, field_version, capability_version
		FROM project_capability_fields
		WHERE project_id = $1 AND module_id = $2 AND capability_id = $3
		ORDER BY field_id, field_version
	`
	rows, err := r.pool.Query(ctx, query, projectID, moduleID, capabilityID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var fields []configuration.DataFieldReference
	for rows.Next() {
		var fieldID, fieldVersion, capabilityVersion string
		if err := rows.Scan(&fieldID, &fieldVersion, &capabilityVersion); err != nil {
			return nil, err
		}
		fields = append(fields, configuration.DataFieldReference{
			ModuleID:          moduleID,
			CapabilityID:      capabilityID,
			CapabilityVersion: capabilityVersion,
			ID:                fieldID,
			Version:           fieldVersion,
		})
	}
	return fields, rows.Err()
}

func insertEnablementFields(ctx context.Context, tx pgx.Tx, item domainenablement.Enablement) error {
	const query = `
		INSERT INTO project_capability_fields
		(project_id, module_id, capability_id, field_id, field_version, capability_version)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, field := range item.DataFields {
		if _, err := tx.Exec(ctx, query, item.ProjectID, item.ModuleID, item.CapabilityID, field.ID, field.Version, field.CapabilityVersion); err != nil {
			return err
		}
	}
	return nil
}
