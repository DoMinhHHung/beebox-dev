package project_test

import (
	"context"
	"sync"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

type fakeRepository struct {
	mu       sync.Mutex
	projects map[string]domainproject.Project
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{projects: make(map[string]domainproject.Project)}
}

var _ project.Repository = (*fakeRepository)(nil)

func (r *fakeRepository) Create(ctx context.Context, p domainproject.Project) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.projects[p.ID]; exists {
		return project.ErrProjectAlreadyExists
	}
	r.projects[p.ID] = p
	return nil
}

func (r *fakeRepository) Get(ctx context.Context, id string) (domainproject.Project, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	p, exists := r.projects[id]
	if !exists {
		return domainproject.Project{}, project.ErrProjectNotFound
	}
	return p, nil
}

func (r *fakeRepository) Update(ctx context.Context, p domainproject.Project) error {
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

func TestCreate_ReturnsProjectOnSuccess(t *testing.T) {
	svc := project.NewService(newFakeRepository())

	got, err := svc.Create(context.Background(), "project-1", "organization-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != domainproject.StatusDraft {
		t.Fatalf("expected status DRAFT, got %q", got.Status)
	}
}

func TestCreate_RejectsInvalidInputAsValidationError(t *testing.T) {
	svc := project.NewService(newFakeRepository())

	_, err := svc.Create(context.Background(), "", "organization-1")
	if apperror.CodeOf(err) != apperror.CodeValidation {
		t.Fatalf("expected CodeValidation, got %v", apperror.CodeOf(err))
	}
}

func TestCreate_RejectsDuplicateAsConflict(t *testing.T) {
	svc := project.NewService(newFakeRepository())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "project-1", "organization-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err := svc.Create(ctx, "project-1", "organization-1")
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %v", apperror.CodeOf(err))
	}
}

func TestGet_ReturnsNotFoundForMissingProject(t *testing.T) {
	svc := project.NewService(newFakeRepository())

	_, err := svc.Get(context.Background(), "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestTransition_AppliesValidTransition(t *testing.T) {
	svc := project.NewService(newFakeRepository())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "project-1", "organization-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := svc.Transition(ctx, "project-1", domainproject.StatusActive)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != domainproject.StatusActive {
		t.Fatalf("expected status ACTIVE, got %q", got.Status)
	}
}

func TestTransition_RejectsInvalidTransitionAsConflict(t *testing.T) {
	svc := project.NewService(newFakeRepository())
	ctx := context.Background()

	if _, err := svc.Create(ctx, "project-1", "organization-1"); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err := svc.Transition(ctx, "project-1", domainproject.StatusArchived)
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %v", apperror.CodeOf(err))
	}
}

func TestTransition_ReturnsNotFoundForMissingProject(t *testing.T) {
	svc := project.NewService(newFakeRepository())

	_, err := svc.Transition(context.Background(), "missing", domainproject.StatusActive)
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}
