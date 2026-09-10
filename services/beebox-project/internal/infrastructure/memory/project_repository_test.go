package memory_test

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func TestCreateAndGet_RoundTrips(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()
	p := domainproject.Project{ID: "project-1", OrganizationID: "organization-1", Status: domainproject.StatusDraft}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repo.Get(ctx, "project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != p {
		t.Fatalf("expected %+v, got %+v", p, got)
	}
}

func TestCreate_RejectsDuplicateID(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()
	p := domainproject.Project{ID: "project-1", OrganizationID: "organization-1", Status: domainproject.StatusDraft}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	err := repo.Create(ctx, p)
	if !errors.Is(err, project.ErrProjectAlreadyExists) {
		t.Fatalf("expected ErrProjectAlreadyExists, got %v", err)
	}
}

func TestGet_ReturnsNotFoundForMissingID(t *testing.T) {
	repo := memory.NewProjectRepository()

	_, err := repo.Get(context.Background(), "missing")
	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestUpdate_ReturnsNotFoundForMissingID(t *testing.T) {
	repo := memory.NewProjectRepository()

	err := repo.Update(context.Background(), domainproject.Project{ID: "missing"})
	if !errors.Is(err, project.ErrProjectNotFound) {
		t.Fatalf("expected ErrProjectNotFound, got %v", err)
	}
}

func TestUpdate_PersistsChange(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()
	p := domainproject.Project{ID: "project-1", OrganizationID: "organization-1", Status: domainproject.StatusDraft}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	p.Status = domainproject.StatusActive
	if err := repo.Update(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repo.Get(ctx, "project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != domainproject.StatusActive {
		t.Fatalf("expected status ACTIVE, got %q", got.Status)
	}
}

func TestRepository_IsSafeForConcurrentUse(t *testing.T) {
	repo := memory.NewProjectRepository()
	ctx := context.Background()

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id := "project-concurrent"
			_ = repo.Create(ctx, domainproject.Project{ID: id, OrganizationID: "organization-1", Status: domainproject.StatusDraft})
			_, _ = repo.Get(ctx, id)
		}()
	}
	wg.Wait()
}
