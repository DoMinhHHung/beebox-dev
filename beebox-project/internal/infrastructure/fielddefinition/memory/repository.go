package memory

import (
	"context"
	"sync"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
)

type key struct {
	projectID string
	version   int
}

type Repository struct {
	mu      sync.RWMutex
	schemas map[key]fielddefinition.Schema
	latest  map[string]int
}

var _ fielddefinition.Repository = (*Repository)(nil)

func New() *Repository {
	return &Repository{
		schemas: make(map[key]fielddefinition.Schema),
		latest:  make(map[string]int),
	}
}

func (r *Repository) Save(ctx context.Context, s fielddefinition.Schema) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	k := key{projectID: s.ProjectID, version: s.Version}
	if _, exists := r.schemas[k]; exists {
		return apperror.New(apperror.CodeConflict, "schema version already exists")
	}

	if current, ok := r.latest[s.ProjectID]; ok && s.Version <= current {
		return apperror.New(apperror.CodeConflict, "schema version must be greater than current latest version")
	}

	r.schemas[k] = s
	r.latest[s.ProjectID] = s.Version
	return nil
}

func (r *Repository) FindLatest(ctx context.Context, projectID string) (fielddefinition.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	version, ok := r.latest[projectID]
	if !ok {
		return fielddefinition.Schema{}, apperror.New(apperror.CodeNotFound, "no schema found for project: "+projectID)
	}
	return r.schemas[key{projectID: projectID, version: version}], nil
}

func (r *Repository) FindVersion(ctx context.Context, projectID string, version int) (fielddefinition.Schema, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.schemas[key{projectID: projectID, version: version}]
	if !ok {
		return fielddefinition.Schema{}, apperror.New(apperror.CodeNotFound, "schema version not found")
	}
	return s, nil
}
