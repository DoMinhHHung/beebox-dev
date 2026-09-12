package postgres

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	domainconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

type ConfigurationRepository struct {
	pool *pgxpool.Pool
}

func NewConfigurationRepository(pool *pgxpool.Pool) *ConfigurationRepository {
	return &ConfigurationRepository{pool: pool}
}

var _ applicationconfiguration.Repository = (*ConfigurationRepository)(nil)

func (r *ConfigurationRepository) CreateVersion(ctx context.Context, version domainconfiguration.Version) error {
	fields, err := json.Marshal(version.Configuration.DataFields)
	if err != nil {
		return err
	}
	const query = `
		INSERT INTO project_configuration_versions
		(project_id, version, module_id, module_version, capability_id, capability_version, data_fields, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = r.pool.Exec(ctx, query, version.ProjectID, version.Number, version.Configuration.ModuleID, version.Configuration.ModuleVersion, version.Configuration.CapabilityID, version.Configuration.CapabilityVersion, fields, string(version.Status))
	if err != nil && isUniqueViolation(err) {
		return applicationconfiguration.ErrVersionConflict
	}
	return err
}

func (r *ConfigurationRepository) GetVersion(ctx context.Context, projectID string, number int) (domainconfiguration.Version, error) {
	const query = `
		SELECT module_id, module_version, capability_id, capability_version, data_fields, status
		FROM project_configuration_versions
		WHERE project_id = $1 AND version = $2
	`
	return r.scanVersion(ctx, query, projectID, number)
}

func (r *ConfigurationRepository) GetLatestVersion(ctx context.Context, projectID string) (domainconfiguration.Version, error) {
	const query = `
		SELECT version, module_id, module_version, capability_id, capability_version, data_fields, status
		FROM project_configuration_versions
		WHERE project_id = $1
		ORDER BY version DESC
		LIMIT 1
	`
	var versionNumber int
	var moduleID, moduleVersion, capabilityID, capabilityVersion, status string
	var fields []byte
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&versionNumber, &moduleID, &moduleVersion, &capabilityID, &capabilityVersion, &fields, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainconfiguration.Version{}, applicationconfiguration.ErrVersionNotFound
	}
	if err != nil {
		return domainconfiguration.Version{}, err
	}
	return buildVersion(projectID, versionNumber, moduleID, moduleVersion, capabilityID, capabilityVersion, fields, status)
}

func (r *ConfigurationRepository) scanVersion(ctx context.Context, query, projectID string, number int) (domainconfiguration.Version, error) {
	var moduleID, moduleVersion, capabilityID, capabilityVersion, status string
	var fields []byte
	err := r.pool.QueryRow(ctx, query, projectID, number).Scan(&moduleID, &moduleVersion, &capabilityID, &capabilityVersion, &fields, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainconfiguration.Version{}, applicationconfiguration.ErrVersionNotFound
	}
	if err != nil {
		return domainconfiguration.Version{}, err
	}
	return buildVersion(projectID, number, moduleID, moduleVersion, capabilityID, capabilityVersion, fields, status)
}

func (r *ConfigurationRepository) UpdateVersion(ctx context.Context, version domainconfiguration.Version) error {
	const query = `UPDATE project_configuration_versions SET status = $3 WHERE project_id = $1 AND version = $2`
	tag, err := r.pool.Exec(ctx, query, version.ProjectID, version.Number, string(version.Status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return applicationconfiguration.ErrVersionNotFound
	}
	return nil
}

func (r *ConfigurationRepository) GetRollout(ctx context.Context, projectID string) (domainconfiguration.RolloutState, error) {
	const query = `SELECT desired_version, applied_version FROM project_configuration_rollouts WHERE project_id = $1`
	var desired, applied int
	err := r.pool.QueryRow(ctx, query, projectID).Scan(&desired, &applied)
	if errors.Is(err, pgx.ErrNoRows) {
		return domainconfiguration.RolloutState{}, applicationconfiguration.ErrRolloutNotFound
	}
	if err != nil {
		return domainconfiguration.RolloutState{}, err
	}
	return domainconfiguration.RolloutState{ProjectID: projectID, DesiredVersion: desired, AppliedVersion: applied}, nil
}

func (r *ConfigurationRepository) SaveRollout(ctx context.Context, state domainconfiguration.RolloutState) error {
	const query = `
		INSERT INTO project_configuration_rollouts (project_id, desired_version, applied_version)
		VALUES ($1, $2, $3)
		ON CONFLICT (project_id) DO UPDATE SET desired_version = EXCLUDED.desired_version, applied_version = EXCLUDED.applied_version
	`
	_, err := r.pool.Exec(ctx, query, state.ProjectID, state.DesiredVersion, state.AppliedVersion)
	return err
}

func buildVersion(projectID string, number int, moduleID, moduleVersion, capabilityID, capabilityVersion string, fields []byte, status string) (domainconfiguration.Version, error) {
	config, err := domainconfiguration.New(projectID, moduleID, moduleVersion, capabilityID, capabilityVersion)
	if err != nil {
		return domainconfiguration.Version{}, err
	}
	var references []domainconfiguration.DataFieldReference
	if len(fields) > 0 {
		if err := json.Unmarshal(fields, &references); err != nil {
			return domainconfiguration.Version{}, err
		}
	}
	config, err = config.WithDataFields(references)
	if err != nil {
		return domainconfiguration.Version{}, err
	}
	version, err := domainconfiguration.NewVersion(projectID, number, config)
	if err != nil {
		return domainconfiguration.Version{}, err
	}
	version.Status = domainconfiguration.LifecycleStatus(status)
	return version, nil
}
