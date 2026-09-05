package credential

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
)

type Service struct {
	repo credential.Repository
}

func NewService(repo credential.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, projectID string, env credential.Environment) (credential.Metadata, string, error) {
	c, secret, err := credential.NewCredential(projectID, env)
	if err != nil {
		return credential.Metadata{}, "", err
	}
	if err := s.repo.Save(ctx, c); err != nil {
		return credential.Metadata{}, "", err
	}
	return c.Metadata(), secret, nil
}

func (s *Service) Rotate(ctx context.Context, id string) (credential.Metadata, string, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
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

func (s *Service) Revoke(ctx context.Context, id string) (credential.Metadata, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
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

func (s *Service) GetMetadata(ctx context.Context, id string) (credential.Metadata, error) {
	c, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return credential.Metadata{}, err
	}
	return c.Metadata(), nil
}