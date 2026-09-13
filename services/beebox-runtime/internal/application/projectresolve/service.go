package projectresolve

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type CredentialVerifier interface {
	VerifyPublic(ctx context.Context, projectID, rawCredential string) (domain.ProjectContext, error)
}

type AppliedConfigurationLoader interface {
	GetAppliedConfiguration(ctx context.Context, projectID string) (domain.AppliedConfiguration, error)
}

type Service struct {
	credentials CredentialVerifier
	configs     AppliedConfigurationLoader
}

func NewService(credentials CredentialVerifier, configs AppliedConfigurationLoader) *Service {
	return &Service{credentials: credentials, configs: configs}
}

func (s *Service) Resolve(ctx context.Context, pathProjectID, rawCredential string) (domain.ProjectContext, error) {
	if pathProjectID == "" {
		return domain.ProjectContext{}, apperror.New(apperror.CodeValidation, "project id is required")
	}
	if rawCredential == "" {
		return domain.ProjectContext{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	verified, err := s.credentials.VerifyPublic(ctx, pathProjectID, rawCredential)
	if err != nil {
		return domain.ProjectContext{}, err
	}
	if verified.ProjectID != pathProjectID {
		return domain.ProjectContext{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	return verified, nil
}

func (s *Service) LoadAppliedConfiguration(ctx context.Context, projectID string) (domain.AppliedConfiguration, error) {
	if projectID == "" {
		return domain.AppliedConfiguration{}, apperror.New(apperror.CodeValidation, "project id is required")
	}
	return s.configs.GetAppliedConfiguration(ctx, projectID)
}
