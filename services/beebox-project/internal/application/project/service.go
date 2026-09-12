package project

import (
	"context"
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, id string, organizationID string) (domainproject.Project, error) {
	p, err := domainproject.New(id, organizationID)
	if err != nil {
		return domainproject.Project{}, apperror.New(apperror.CodeValidation, err.Error())
	}

	if err := s.repo.Create(ctx, p); err != nil {
		if errors.Is(err, ErrProjectAlreadyExists) {
			return domainproject.Project{}, apperror.New(apperror.CodeConflict, "project already exists")
		}
		return domainproject.Project{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to create project", err)
	}

	return p, nil
}

func (s *Service) Get(ctx context.Context, id string) (domainproject.Project, error) {
	if id == "" {
		return domainproject.Project{}, apperror.New(apperror.CodeValidation, "project id is required")
	}

	p, err := s.repo.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return domainproject.Project{}, apperror.New(apperror.CodeNotFound, "project not found")
		}
		return domainproject.Project{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load project", err)
	}

	return p, nil
}

func (s *Service) GetAuthorized(ctx context.Context, id string, organizationID string) (domainproject.Project, error) {
	p, err := s.Get(ctx, id)
	if err != nil {
		return domainproject.Project{}, err
	}
	if organizationID == "" || p.OrganizationID != organizationID {
		return domainproject.Project{}, apperror.New(apperror.CodeForbidden, "forbidden")
	}
	return p, nil
}

func (s *Service) Transition(ctx context.Context, id string, to domainproject.Status) (domainproject.Project, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return domainproject.Project{}, err
	}

	next, err := current.Transition(to)
	if err != nil {
		return domainproject.Project{}, apperror.New(apperror.CodeConflict, err.Error())
	}

	if err := s.repo.Update(ctx, next); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return domainproject.Project{}, apperror.New(apperror.CodeNotFound, "project not found")
		}
		if errors.Is(err, ErrProjectConflict) {
			return domainproject.Project{}, apperror.New(apperror.CodeConflict, "project was updated concurrently")
		}
		return domainproject.Project{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to update project", err)
	}

	return next, nil
}

func (s *Service) TransitionAuthorized(ctx context.Context, id string, to domainproject.Status, organizationID string) (domainproject.Project, error) {
	current, err := s.GetAuthorized(ctx, id, organizationID)
	if err != nil {
		return domainproject.Project{}, err
	}

	next, err := current.Transition(to)
	if err != nil {
		return domainproject.Project{}, apperror.New(apperror.CodeConflict, err.Error())
	}

	if err := s.repo.Update(ctx, next); err != nil {
		if errors.Is(err, ErrProjectNotFound) {
			return domainproject.Project{}, apperror.New(apperror.CodeNotFound, "project not found")
		}
		if errors.Is(err, ErrProjectConflict) {
			return domainproject.Project{}, apperror.New(apperror.CodeConflict, "project was updated concurrently")
		}
		return domainproject.Project{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to update project", err)
	}

	return next, nil
}
