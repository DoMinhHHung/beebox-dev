package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
)

func TestProjectRepository_SaveAndFindByID(t *testing.T) {
	repo := NewProjectRepository()
	p, err := project.NewProject("id-1", "owner-1", "My Project", project.TierFree, time.Now())
	if err != nil {
		t.Fatalf("unexpected error building project: %v", err)
	}

	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("expected no error saving, got %v", err)
	}

	found, err := repo.FindByID(context.Background(), "id-1")
	if err != nil {
		t.Fatalf("expected to find project, got error: %v", err)
	}
	if found.Name != "My Project" {
		t.Fatalf("expected name %q, got %q", "My Project", found.Name)
	}
}

func TestProjectRepository_FindByID_NotFound(t *testing.T) {
	repo := NewProjectRepository()

	_, err := repo.FindByID(context.Background(), "missing")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", apperror.CodeOf(err))
	}
}
