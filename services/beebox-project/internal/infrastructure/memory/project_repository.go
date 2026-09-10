package memory

import (
	"context"
	"sync"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

type ProjectRepository struct {
	mu       sync.RWMutex
	projects map[string]domainproject.Project
}

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{projects: make(map[string]domainproject.Project)}
}

var _ project.Repository = (*ProjectRepository)(nil)

func (r *ProjectRepository) Create(ctx context.Context, p domainproject.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.projects[p.ID]; exists {
		return project.ErrProjectAlreadyExists
	}
	r.projects[p.ID] = p
	return nil
}

func (r *ProjectRepository) Get(ctx context.Context, id string) (domainproject.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, exists := r.projects[id]
	if !exists {
		return domainproject.Project{}, project.ErrProjectNotFound
	}
	return p, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p domainproject.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	current, exists := r.projects[p.ID]
	if !exists {
		return project.ErrProjectNotFound
	}
	if current.Revision != p.Revision-1 {
		return project.ErrProjectConflict
	}
	r.projects[p.ID] = p
	return nil
}
