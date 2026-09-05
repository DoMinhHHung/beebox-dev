package projectapp

import (
	"context"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
)

type Service struct {
	repo project.Repository
}

func NewService(repo project.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateProject(ctx context.Context, ownerID, name string, tier project.Tier) (project.Project, error) {
	id, err := idgen.New()
	if err != nil {
		return project.Project{}, apperror.Wrap(apperror.CodeInternal, "failed to generate project id", err)
	}

	p, err := project.NewProject(id, ownerID, name, tier, time.Now())
	if err != nil {
		return project.Project{}, err
	}

	if err := s.repo.Save(ctx, p); err != nil {
		return project.Project{}, err
	}

	return p, nil
}