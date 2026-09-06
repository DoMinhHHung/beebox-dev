package memory

import (
	"context"
	"sync"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
)

type Repository struct {
	mu          sync.RWMutex
	byID        map[string]credential.Credential
	byPublicKey map[string]string
}

var _ credential.Repository = (*Repository)(nil)

func New() *Repository {
	return &Repository{
		byID:        make(map[string]credential.Credential),
		byPublicKey: make(map[string]string),
	}
}

func (r *Repository) Save(ctx context.Context, c credential.Credential) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existing, ok := r.byID[c.ID]; ok && existing.PublicKey != c.PublicKey {
		return apperror.New(apperror.CodeInvalidInput, "public key must not change for an existing credential id")
	}

	r.byID[c.ID] = c
	r.byPublicKey[c.PublicKey] = c.ID
	return nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (credential.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.byID[id]
	if !ok {
		return credential.Credential{}, apperror.New(apperror.CodeNotFound, "credential not found")
	}
	return c, nil
}

func (r *Repository) FindByPublicKey(ctx context.Context, publicKey string) (credential.Credential, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byPublicKey[publicKey]
	if !ok {
		return credential.Credential{}, apperror.New(apperror.CodeNotFound, "credential not found")
	}
	return r.byID[id], nil
}
