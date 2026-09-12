package credential

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	domaincredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/credential"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

var (
	ErrCredentialNotFound      = errors.New("credential not found")
	ErrCredentialAlreadyExists = errors.New("credential already exists")
)

type Repository interface {
	Create(context.Context, domaincredential.Credential) error
	FindByID(context.Context, string, string) (domaincredential.Credential, error)
	FindByProjectAndHash(context.Context, string, domaincredential.Kind, string) (domaincredential.Credential, error)
	Update(context.Context, domaincredential.Credential) error
}

type ProjectAuthorizer interface {
	Get(context.Context, string) (domainproject.Project, error)
	GetAuthorized(context.Context, string, string) (domainproject.Project, error)
}

type Service struct {
	repo     Repository
	projects ProjectAuthorizer
	now      func() time.Time
}

func NewService(repo Repository, projects ProjectAuthorizer) *Service {
	return &Service{
		repo:     repo,
		projects: projects,
		now:      time.Now,
	}
}

type IssuedPublicCredential struct {
	Credential domaincredential.Credential
	Secret     string
}

type VerifiedPublicCredential struct {
	ProjectID    string
	CredentialID string
	Kind         domaincredential.Kind
}

func (s *Service) IssuePublic(ctx context.Context, projectID, organizationID, name string) (IssuedPublicCredential, error) {
	if name == "" {
		return IssuedPublicCredential{}, apperror.New(apperror.CodeValidation, "credential name is required")
	}

	project, err := s.projects.GetAuthorized(ctx, projectID, organizationID)
	if err != nil {
		return IssuedPublicCredential{}, err
	}
	if project.Status != domainproject.StatusActive {
		return IssuedPublicCredential{}, apperror.New(apperror.CodeForbidden, "project is not active")
	}

	id, err := newCredentialID()
	if err != nil {
		return IssuedPublicCredential{}, apperror.Wrap(apperror.CodeInternal, "failed to issue credential", err)
	}

	issued, err := domaincredential.New(projectID, id, name, domaincredential.KindPublic)
	if err != nil {
		return IssuedPublicCredential{}, apperror.New(apperror.CodeValidation, err.Error())
	}

	stored := issued.Credential
	stored.Value = ""

	if err := s.repo.Create(ctx, stored); err != nil {
		if errors.Is(err, ErrCredentialAlreadyExists) {
			return IssuedPublicCredential{}, apperror.New(apperror.CodeConflict, "credential already exists")
		}
		return IssuedPublicCredential{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to store credential", err)
	}

	return IssuedPublicCredential{
		Credential: stored,
		Secret:     issued.Secret,
	}, nil
}

func (s *Service) VerifyPublic(ctx context.Context, projectID, raw string) (VerifiedPublicCredential, error) {
	if projectID == "" || raw == "" {
		return VerifiedPublicCredential{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	project, err := s.projects.Get(ctx, projectID)
	if err != nil {
		if apperror.IsCode(err, apperror.CodeNotFound) {
			return VerifiedPublicCredential{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
		}
		return VerifiedPublicCredential{}, err
	}
	if project.Status != domainproject.StatusActive {
		return VerifiedPublicCredential{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	hash := domaincredential.HashSecret(raw)
	found, err := s.repo.FindByProjectAndHash(ctx, projectID, domaincredential.KindPublic, hash)
	if errors.Is(err, ErrCredentialNotFound) {
		return VerifiedPublicCredential{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	if err != nil {
		return VerifiedPublicCredential{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to verify credential", err)
	}

	if found.Kind != domaincredential.KindPublic {
		return VerifiedPublicCredential{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	if found.Status != domaincredential.StatusActive || found.IsExpired(s.now()) || !found.Matches(raw) {
		return VerifiedPublicCredential{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	return VerifiedPublicCredential{
		ProjectID:    found.ProjectID,
		CredentialID: found.ID,
		Kind:         found.Kind,
	}, nil
}

func (s *Service) Revoke(ctx context.Context, projectID, organizationID, credentialID string) (domaincredential.Credential, error) {
	if _, err := s.projects.GetAuthorized(ctx, projectID, organizationID); err != nil {
		return domaincredential.Credential{}, err
	}

	found, err := s.repo.FindByID(ctx, projectID, credentialID)
	if errors.Is(err, ErrCredentialNotFound) {
		return domaincredential.Credential{}, apperror.New(apperror.CodeNotFound, "credential not found")
	}
	if err != nil {
		return domaincredential.Credential{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to load credential", err)
	}

	revoked, err := found.Revoke()
	if err != nil {
		return domaincredential.Credential{}, apperror.New(apperror.CodeConflict, err.Error())
	}
	revoked.Value = ""

	if err := s.repo.Update(ctx, revoked); err != nil {
		if errors.Is(err, ErrCredentialNotFound) {
			return domaincredential.Credential{}, apperror.New(apperror.CodeNotFound, "credential not found")
		}
		return domaincredential.Credential{}, apperror.Wrap(apperror.CodeDependencyFailure, "failed to revoke credential", err)
	}

	return revoked, nil
}

func newCredentialID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
