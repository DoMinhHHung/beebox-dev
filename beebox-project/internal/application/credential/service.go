package credential

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
)

type ProjectReader interface {
	FindByID(ctx context.Context, id string) (project.Project, error)
}

type Service struct {
	repo     credential.Repository
	projects ProjectReader
}

// NewService tạo một dịch vụ quản lý credential với repository credential và bộ đọc project được cung cấp.
func NewService(repo credential.Repository, projects ProjectReader) *Service {
	return &Service{repo: repo, projects: projects}
}

func (s *Service) Create(ctx context.Context, actorOwnerID, projectID string, env credential.Environment) (credential.Metadata, string, error) {
	if err := s.checkOwnership(ctx, actorOwnerID, projectID); err != nil {
		return credential.Metadata{}, "", err
	}

	c, secret, err := credential.NewCredential(projectID, env)
	if err != nil {
		return credential.Metadata{}, "", err
	}
	if err := s.repo.Save(ctx, c); err != nil {
		return credential.Metadata{}, "", err
	}
	return c.Metadata(), secret, nil
}

func (s *Service) Rotate(ctx context.Context, actorOwnerID, id string) (credential.Metadata, string, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return credential.Metadata{}, "", err
	}
	if err := s.checkOwnership(ctx, actorOwnerID, c.ProjectID); err != nil {
		return credential.Metadata{}, "", err
	}

	rotated, secret, err := c.Rotate()
	if err != nil {
		return credential.Metadata{}, "", err
	}
	if err := s.repo.Save(ctx, rotated); err != nil {
		return credential.Metadata{}, "", err
	}
	return rotated.Metadata(), secret, nil
}

func (s *Service) Revoke(ctx context.Context, actorOwnerID, id string) (credential.Metadata, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return credential.Metadata{}, err
	}
	if err := s.checkOwnership(ctx, actorOwnerID, c.ProjectID); err != nil {
		return credential.Metadata{}, err
	}

	revoked, err := c.Revoke()
	if err != nil {
		return credential.Metadata{}, err
	}
	if err := s.repo.Save(ctx, revoked); err != nil {
		return credential.Metadata{}, err
	}
	return revoked.Metadata(), nil
}

func (s *Service) GetMetadata(ctx context.Context, actorOwnerID, id string) (credential.Metadata, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return credential.Metadata{}, err
	}
	if err := s.checkOwnership(ctx, actorOwnerID, c.ProjectID); err != nil {
		return credential.Metadata{}, err
	}
	return c.Metadata(), nil
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
