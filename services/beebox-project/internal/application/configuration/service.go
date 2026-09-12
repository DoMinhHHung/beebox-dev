package configuration

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	domainconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

type Repository interface {
	CreateVersion(context.Context, domainconfiguration.Version) error
	GetVersion(context.Context, string, int) (domainconfiguration.Version, error)
	GetLatestVersion(context.Context, string) (domainconfiguration.Version, error)
	UpdateVersion(context.Context, domainconfiguration.Version) error
	GetRollout(context.Context, string) (domainconfiguration.RolloutState, error)
	SaveRollout(context.Context, domainconfiguration.RolloutState) error
}

type ProjectAuthorizer interface {
	GetAuthorized(context.Context, string, string) (domainproject.Project, error)
}

type Service struct {
	repo     Repository
	projects ProjectAuthorizer
	catalog  catalog.Catalog
}

func NewService(repo Repository, projects ProjectAuthorizer, cat catalog.Catalog) *Service {
	return &Service{repo: repo, projects: projects, catalog: cat}
}

func (s *Service) CreateOrUpdate(ctx context.Context, projectID, organizationID string, config domainconfiguration.Configuration) (domainconfiguration.Version, error) {
	if _, err := s.authorize(ctx, projectID, organizationID); err != nil {
		return domainconfiguration.Version{}, err
	}
	if config.ProjectID != projectID {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, domainconfiguration.ErrInvalidConfiguration.Error())
	}
	validated, err := domainconfiguration.New(projectID, config.ModuleID, config.ModuleVersion, config.CapabilityID, config.CapabilityVersion)
	if err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	validated, err = validated.WithDataFields(config.DataFields)
	if err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	if _, err := s.catalog.Module(validated.ModuleID, validated.ModuleVersion); err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	if _, err := s.catalog.Capability(validated.ModuleID, validated.CapabilityID, validated.CapabilityVersion); err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	for _, field := range validated.DataFields {
		if _, err := s.catalog.DataField(field.ModuleID, field.CapabilityID, field.CapabilityVersion, field.ID, field.Version); err != nil {
			return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
		}
	}
	config = validated
	latest, err := s.repo.GetLatestVersion(ctx, projectID)
	if err != nil && !errors.Is(err, ErrVersionNotFound) {
		return domainconfiguration.Version{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load configuration", err)
	}
	var previousFields []domainconfiguration.DataFieldReference
	number := 1
	if err == nil {
		previousFields = latest.Configuration.DataFields
		number = latest.Number + 1
	}
	if err := s.catalog.ValidateDataFieldTransition(previousFields, validated.DataFields); err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	version, err := domainconfiguration.NewVersion(projectID, number, config)
	if err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	if err := s.repo.CreateVersion(ctx, version); err != nil {
		if errors.Is(err, ErrVersionConflict) {
			return domainconfiguration.Version{}, apperror.New(apperror.CodeConflict, "configuration version already exists")
		}
		return domainconfiguration.Version{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to save configuration", err)
	}
	return version, nil
}

func (s *Service) Current(ctx context.Context, projectID, organizationID string) (domainconfiguration.Version, error) {
	if _, err := s.authorize(ctx, projectID, organizationID); err != nil {
		return domainconfiguration.Version{}, err
	}
	version, err := s.repo.GetLatestVersion(ctx, projectID)
	if errors.Is(err, ErrVersionNotFound) {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeNotFound, "configuration not found")
	}
	if err != nil {
		return domainconfiguration.Version{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load configuration", err)
	}
	return version, nil
}

func (s *Service) Version(ctx context.Context, projectID, organizationID string, number int) (domainconfiguration.Version, error) {
	if _, err := s.authorize(ctx, projectID, organizationID); err != nil {
		return domainconfiguration.Version{}, err
	}
	version, err := s.repo.GetVersion(ctx, projectID, number)
	if errors.Is(err, ErrVersionNotFound) {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeNotFound, "configuration version not found")
	}
	if err != nil {
		return domainconfiguration.Version{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load configuration version", err)
	}
	return version, nil
}

func (s *Service) Transition(ctx context.Context, projectID, organizationID string, number int, status domainconfiguration.LifecycleStatus) (domainconfiguration.Version, error) {
	version, err := s.Version(ctx, projectID, organizationID, number)
	if err != nil {
		return domainconfiguration.Version{}, err
	}
	next, err := version.Transition(status)
	if err != nil {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeConflict, err.Error())
	}
	if err := s.repo.UpdateVersion(ctx, next); err != nil {
		if errors.Is(err, ErrVersionNotFound) {
			return domainconfiguration.Version{}, apperror.New(apperror.CodeNotFound, "configuration version not found")
		}
		if errors.Is(err, ErrVersionConflict) {
			return domainconfiguration.Version{}, apperror.New(apperror.CodeConflict, "configuration version was updated concurrently")
		}
		return domainconfiguration.Version{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to update configuration version", err)
	}
	return next, nil
}

func (s *Service) Rollout(ctx context.Context, projectID, organizationID string, number int) (domainconfiguration.RolloutState, error) {
	version, err := s.Version(ctx, projectID, organizationID, number)
	if err != nil {
		return domainconfiguration.RolloutState{}, err
	}
	state, err := s.repo.GetRollout(ctx, projectID)
	if errors.Is(err, ErrRolloutNotFound) {
		state, err = domainconfiguration.NewRolloutState(projectID)
	}
	if err != nil {
		return domainconfiguration.RolloutState{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load rollout", err)
	}
	state, err = state.WithDesiredVersion(version)
	if err != nil {
		return domainconfiguration.RolloutState{}, apperror.New(apperror.CodeConflict, err.Error())
	}
	if err := s.repo.SaveRollout(ctx, state); err != nil {
		return domainconfiguration.RolloutState{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to save rollout", err)
	}
	return state, nil
}

func (s *Service) authorize(ctx context.Context, projectID, organizationID string) (domainproject.Project, error) {
	if projectID == "" {
		return domainproject.Project{}, apperror.New(apperror.CodeValidation, "project id is required")
	}
	return s.projects.GetAuthorized(ctx, projectID, organizationID)
}

var (
	ErrVersionNotFound = errors.New("configuration version not found")
	ErrVersionConflict = errors.New("configuration version conflict")
	ErrRolloutNotFound = errors.New("configuration rollout not found")
)
