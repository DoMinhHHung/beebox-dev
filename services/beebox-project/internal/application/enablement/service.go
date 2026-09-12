package enablement

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/enablement"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

var (
	ErrEnablementNotFound      = errors.New("enablement not found")
	ErrEnablementAlreadyExists = errors.New("enablement already exists")
)

type Repository interface {
	Save(context.Context, domainenablement.Enablement) error
	Update(context.Context, domainenablement.Enablement) error
	Get(context.Context, string, string, string) (domainenablement.Enablement, error)
	ListByProject(context.Context, string) ([]domainenablement.Enablement, error)
	Delete(context.Context, string, string, string) error
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

func (s *Service) ListCatalogModules() []catalog.ModuleDefinition {
	return s.catalog.Modules()
}

func (s *Service) GetCatalogModule(moduleID, version string) (catalog.ModuleDefinition, error) {
	definition, err := s.catalog.Module(moduleID, version)
	if err != nil {
		return catalog.ModuleDefinition{}, mapCatalogError(err)
	}
	return definition, nil
}

func (s *Service) GetCatalogCapability(moduleID, capabilityID, version string) (catalog.CapabilityDefinition, error) {
	definition, err := s.catalog.Capability(moduleID, capabilityID, version)
	if err != nil {
		return catalog.CapabilityDefinition{}, mapCatalogError(err)
	}
	return definition, nil
}

func (s *Service) ListCatalogCapabilities(moduleID string) []catalog.CapabilityDefinition {
	return s.catalog.Capabilities(moduleID)
}

func (s *Service) Enable(
	ctx context.Context,
	projectID string,
	organizationID string,
	moduleID string,
	moduleVersion string,
	capabilityID string,
	capabilityVersion string,
	fields []configuration.DataFieldReference,
) (domainenablement.Enablement, error) {
	if _, err := s.projects.GetAuthorized(ctx, projectID, organizationID); err != nil {
		return domainenablement.Enablement{}, err
	}

	if _, err := s.catalog.Module(moduleID, moduleVersion); err != nil {
		return domainenablement.Enablement{}, mapCatalogError(err)
	}
	if _, err := s.catalog.Capability(moduleID, capabilityID, capabilityVersion); err != nil {
		return domainenablement.Enablement{}, mapCatalogError(err)
	}

	enabled, err := domainenablement.New(projectID, moduleID, moduleVersion, capabilityID, capabilityVersion)
	if err != nil {
		return domainenablement.Enablement{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	enabled, err = enabled.WithDataFields(fields)
	if err != nil {
		return domainenablement.Enablement{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	if err := s.validateFields(fields, moduleID, capabilityID, capabilityVersion); err != nil {
		return domainenablement.Enablement{}, err
	}

	if err := s.repo.Save(ctx, enabled); err != nil {
		if errors.Is(err, ErrEnablementAlreadyExists) {
			return domainenablement.Enablement{}, apperror.New(apperror.CodeConflict, "capability already enabled")
		}
		return domainenablement.Enablement{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to enable capability", err)
	}
	return enabled, nil
}

func (s *Service) UpdateFields(
	ctx context.Context,
	projectID string,
	organizationID string,
	moduleID string,
	capabilityID string,
	fields []configuration.DataFieldReference,
) (domainenablement.Enablement, error) {
	if _, err := s.projects.GetAuthorized(ctx, projectID, organizationID); err != nil {
		return domainenablement.Enablement{}, err
	}

	current, err := s.repo.Get(ctx, projectID, moduleID, capabilityID)
	if err != nil {
		if errors.Is(err, ErrEnablementNotFound) {
			return domainenablement.Enablement{}, apperror.New(apperror.CodeNotFound, "capability not enabled")
		}
		return domainenablement.Enablement{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load enablement", err)
	}

	if err := s.catalog.ValidateDataFieldTransition(current.DataFields, fields); err != nil {
		return domainenablement.Enablement{}, mapCatalogError(err)
	}

	updated, err := current.WithDataFields(fields)
	if err != nil {
		return domainenablement.Enablement{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	if err := s.validateFields(fields, current.ModuleID, current.CapabilityID, current.CapabilityVersion); err != nil {
		return domainenablement.Enablement{}, err
	}

	if err := s.repo.Update(ctx, updated); err != nil {
		if errors.Is(err, ErrEnablementNotFound) {
			return domainenablement.Enablement{}, apperror.New(apperror.CodeNotFound, "capability not enabled")
		}
		return domainenablement.Enablement{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to update enablement", err)
	}
	return updated, nil
}

func (s *Service) Disable(
	ctx context.Context,
	projectID string,
	organizationID string,
	moduleID string,
	capabilityID string,
) error {
	if _, err := s.projects.GetAuthorized(ctx, projectID, organizationID); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, projectID, moduleID, capabilityID); err != nil {
		if errors.Is(err, ErrEnablementNotFound) {
			return apperror.New(apperror.CodeNotFound, "capability not enabled")
		}
		return apperror.Wrap(apperror.CodeDependencyFailure, "failed to disable capability", err)
	}
	return nil
}

func (s *Service) List(
	ctx context.Context,
	projectID string,
	organizationID string,
) ([]domainenablement.Enablement, error) {
	if _, err := s.projects.GetAuthorized(ctx, projectID, organizationID); err != nil {
		return nil, err
	}
	items, err := s.repo.ListByProject(ctx, projectID)
	if err != nil {
		return nil, apperror.Wrap(apperror.CodeDependencyFailure, "failed to list enablements", err)
	}
	if items == nil {
		items = []domainenablement.Enablement{}
	}
	return items, nil
}

func (s *Service) validateFields(
	fields []configuration.DataFieldReference,
	moduleID string,
	capabilityID string,
	capabilityVersion string,
) error {
	for _, field := range fields {
		if field.ModuleID != moduleID || field.CapabilityID != capabilityID || field.CapabilityVersion != capabilityVersion {
			return apperror.New(apperror.CodeValidation, catalog.ErrUnknownDataField.Error())
		}
		if _, err := s.catalog.DataField(field.ModuleID, field.CapabilityID, field.CapabilityVersion, field.ID, field.Version); err != nil {
			return mapCatalogError(err)
		}
	}
	return nil
}

func mapCatalogError(err error) error {
	switch {
	case errors.Is(err, catalog.ErrUnknownModule),
		errors.Is(err, catalog.ErrUnknownCapability),
		errors.Is(err, catalog.ErrUnknownDataField),
		errors.Is(err, catalog.ErrIncompatibleDataFieldChange),
		errors.Is(err, configuration.ErrDuplicateDataField),
		errors.Is(err, domainenablement.ErrDuplicateField):
		return apperror.New(apperror.CodeValidation, err.Error())
	default:
		return apperror.New(apperror.CodeValidation, err.Error())
	}
}
