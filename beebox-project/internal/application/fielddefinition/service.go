package fielddefinition

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
)

type ProjectReader interface {
	FindByID(ctx context.Context, id string) (project.Project, error)
}

type Service struct {
	repo     fielddefinition.Repository
	projects ProjectReader
}

// NewService tạo một Service với repository định nghĩa trường và trình đọc dự án được cung cấp.
func NewService(repo fielddefinition.Repository, projects ProjectReader) *Service {
	return &Service{repo: repo, projects: projects}
}

func (s *Service) Define(ctx context.Context, actorOwnerID, projectID string, fields []fielddefinition.FieldDefinition) (fielddefinition.Schema, error) {
	if err := s.checkOwnership(ctx, actorOwnerID, projectID); err != nil {
		return fielddefinition.Schema{}, err
	}

	existing, err := s.repo.FindLatest(ctx, projectID)
	if err != nil {
		if apperror.CodeOf(err) != apperror.CodeNotFound {
			return fielddefinition.Schema{}, err
		}
		return s.createInitial(ctx, projectID, fields)
	}
	return s.createNext(ctx, existing, fields)
}

func (s *Service) createInitial(ctx context.Context, projectID string, fields []fielddefinition.FieldDefinition) (fielddefinition.Schema, error) {
	schema, err := fielddefinition.NewSchema(projectID, fields)
	if err != nil {
		return fielddefinition.Schema{}, err
	}
	if err := s.repo.Save(ctx, schema); err != nil {
		return fielddefinition.Schema{}, err
	}
	return schema, nil
}

func (s *Service) createNext(ctx context.Context, current fielddefinition.Schema, fields []fielddefinition.FieldDefinition) (fielddefinition.Schema, error) {
	next, err := current.NextVersion(fields)
	if err != nil {
		return fielddefinition.Schema{}, err
	}
	if err := s.repo.Save(ctx, next); err != nil {
		return fielddefinition.Schema{}, err
	}
	return next, nil
}

func (s *Service) GetLatest(ctx context.Context, actorOwnerID, projectID string) (fielddefinition.Schema, error) {
	if err := s.checkOwnership(ctx, actorOwnerID, projectID); err != nil {
		return fielddefinition.Schema{}, err
	}
	return s.repo.FindLatest(ctx, projectID)
}

func (s *Service) GetVersion(ctx context.Context, actorOwnerID, projectID string, version int) (fielddefinition.Schema, error) {
	if err := s.checkOwnership(ctx, actorOwnerID, projectID); err != nil {
		return fielddefinition.Schema{}, err
	}
	return s.repo.FindVersion(ctx, projectID, version)
}

func (s *Service) checkOwnership(ctx context.Context, actorOwnerID, projectID string) error {
	p, err := s.projects.FindByID(ctx, projectID)
	if err != nil {
		return err
	}
	if p.OwnerID != actorOwnerID {
		return apperror.New(apperror.CodeTenantAccessDenied, "project does not belong to the authenticated owner")
	}
	return nil
}
