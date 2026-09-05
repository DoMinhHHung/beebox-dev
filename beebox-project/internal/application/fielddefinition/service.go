package fielddefinition

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
)

type Service struct {
	repo fielddefinition.Repository
}

func NewService(repo fielddefinition.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Define(ctx context.Context, projectID string, fields []fielddefinition.FieldDefinition) (fielddefinition.Schema, error) {
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

func (s *Service) GetLatest(ctx context.Context, projectID string) (fielddefinition.Schema, error) {
	return s.repo.FindLatest(ctx, projectID)
}

func (s *Service) GetVersion(ctx context.Context, projectID string, version int) (fielddefinition.Schema, error) {
	return s.repo.FindVersion(ctx, projectID, version)
}