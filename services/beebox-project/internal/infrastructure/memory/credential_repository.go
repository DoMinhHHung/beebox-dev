package memory

import (
	"context"
	"sync"

	applicationcredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/credential"
	domaincredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/credential"
)

type CredentialRepository struct {
	mu          sync.Mutex
	byProjectID map[string]map[string]domaincredential.Credential
}

func NewCredentialRepository() *CredentialRepository {
	return &CredentialRepository{
		byProjectID: make(map[string]map[string]domaincredential.Credential),
	}
}

var _ applicationcredential.Repository = (*CredentialRepository)(nil)

func (r *CredentialRepository) Create(_ context.Context, credential domaincredential.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	credential.Value = ""
	if credential.SecretHash == "" {
		return applicationcredential.ErrCredentialAlreadyExists
	}

	for _, projectCreds := range r.byProjectID {
		for _, existing := range projectCreds {
			if existing.SecretHash == credential.SecretHash {
				return applicationcredential.ErrCredentialAlreadyExists
			}
		}
	}

	projectCreds := r.byProjectID[credential.ProjectID]
	if projectCreds == nil {
		projectCreds = make(map[string]domaincredential.Credential)
		r.byProjectID[credential.ProjectID] = projectCreds
	}
	if _, exists := projectCreds[credential.ID]; exists {
		return applicationcredential.ErrCredentialAlreadyExists
	}
	projectCreds[credential.ID] = credential
	return nil
}

func (r *CredentialRepository) FindByID(_ context.Context, projectID, id string) (domaincredential.Credential, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	credential, ok := r.byProjectID[projectID][id]
	if !ok {
		return domaincredential.Credential{}, applicationcredential.ErrCredentialNotFound
	}
	credential.Value = ""
	return credential, nil
}

func (r *CredentialRepository) FindByProjectAndHash(_ context.Context, projectID string, kind domaincredential.Kind, secretHash string) (domaincredential.Credential, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, credential := range r.byProjectID[projectID] {
		if credential.Kind == kind && credential.SecretHash == secretHash {
			credential.Value = ""
			return credential, nil
		}
	}
	return domaincredential.Credential{}, applicationcredential.ErrCredentialNotFound
}

func (r *CredentialRepository) Update(_ context.Context, credential domaincredential.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	projectCreds := r.byProjectID[credential.ProjectID]
	if projectCreds == nil {
		return applicationcredential.ErrCredentialNotFound
	}
	if _, ok := projectCreds[credential.ID]; !ok {
		return applicationcredential.ErrCredentialNotFound
	}
	credential.Value = ""
	projectCreds[credential.ID] = credential
	return nil
}
