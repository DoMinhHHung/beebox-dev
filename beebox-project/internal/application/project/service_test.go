package projectapp

import (
	"context"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
)

func TestService_CreateProject_Valid(t *testing.T) {
	repo := memory.NewProjectRepository()
	svc := NewService(repo)

	p, err := svc.CreateProject(context.Background(), "owner-1", "My Project", project.TierFree)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID == "" {
		t.Fatal("expected generated ID, got empty string")
	}

	found, err := repo.FindByID(context.Background(), p.ID)
	if err != nil {
		t.Fatalf("expected saved project to be findable, got error: %v", err)
	}
	if found.Name != "My Project" {
		t.Fatalf("expected name %q, got %q", "My Project", found.Name)
	}
}

func TestService_CreateProject_InvalidName(t *testing.T) {
	repo := memory.NewProjectRepository()
	svc := NewService(repo)

	_, err := svc.CreateProject(context.Background(), "owner-1", "", project.TierFree)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %s", apperror.CodeOf(err))
	}
}
