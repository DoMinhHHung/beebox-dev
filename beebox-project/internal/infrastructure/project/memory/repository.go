package memory

import (
	"context"
	"sync"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
)

type ProjectRepository struct {
	mu    sync.RWMutex
	store map[string]project.Project
}

var _ project.Repository = (*ProjectRepository)(nil)

func NewProjectRepository() *ProjectRepository {
	return &ProjectRepository{store: make(map[string]project.Project)}
}

func (r *ProjectRepository) Save(ctx context.Context, p project.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[p.ID] = p
	return nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (project.Project, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	p, ok := r.store[id]
	if !ok {
		return project.Project{}, apperror.New(apperror.CodeNotFound, "project not found")
	}
	return p, nil
}