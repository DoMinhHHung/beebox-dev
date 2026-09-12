package authcap

import (
	"context"
	"strings"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/capability"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type SessionReader interface {
	GetSession(ctx context.Context, bearerToken string) (Session, error)
}

type Session struct {
	UserID         string
	SessionID      string
	OrganizationID string
	ExpiresAt      *time.Time
}

type ProjectSession struct {
	ProjectID      string
	UserID         string
	SessionID      string
	OrganizationID string
	ExpiresAt      *time.Time
}

type Service struct {
	identity SessionReader
}

func NewService(identity SessionReader) *Service {
	return &Service{identity: identity}
}

func (s *Service) CurrentSession(ctx context.Context, project domain.ProjectContext, cfg domain.AppliedConfiguration, authorizationHeader string) (ProjectSession, error) {
	if err := capability.RequireSessionAuth(cfg); err != nil {
		return ProjectSession{}, err
	}

	token, err := parseBearer(authorizationHeader)
	if err != nil {
		return ProjectSession{}, err
	}

	session, err := s.identity.GetSession(ctx, token)
	if err != nil {
		return ProjectSession{}, err
	}

	return ProjectSession{
		ProjectID:      project.ProjectID,
		UserID:         session.UserID,
		SessionID:      session.SessionID,
		OrganizationID: session.OrganizationID,
		ExpiresAt:      session.ExpiresAt,
	}, nil
}

func parseBearer(header string) (string, error) {
	header = strings.TrimSpace(header)
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
		return "", apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	return strings.TrimSpace(parts[1]), nil
}
