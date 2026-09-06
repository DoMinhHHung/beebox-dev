package fielddefinition_test

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	fielddefinitionApp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/fielddefinition/memory"
	projectmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
)

func field(t *testing.T, name string, kind fielddefinition.FieldKind, required bool) fielddefinition.FieldDefinition {
	t.Helper()
	f, err := fielddefinition.NewFieldDefinition(name, kind, required)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return f
}

func mustFieldDefTestProject(t *testing.T, repo *projectmemory.ProjectRepository, id, ownerID string) project.Project {
	t.Helper()
	p, err := project.NewProject(id, ownerID, "Test Project", project.TierFree, time.Now())
	if err != nil {
		t.Fatalf("unexpected error building project: %v", err)
	}
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("unexpected error saving project: %v", err)
	}
	return p
}

func TestService_Define_CreatesInitialVersion(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	mustFieldDefTestProject(t, projects, "proj-1", "owner-1")
	svc := fielddefinitionApp.NewService(repo, projects)
	ctx := context.Background()

	schema, err := svc.Define(ctx, "owner-1", "proj-1", []fielddefinition.FieldDefinition{
		field(t, "fullName", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if schema.Version != fielddefinition.InitialVersion {
		t.Fatalf("expected initial version, got %d", schema.Version)
	}
}

func TestService_Define_CreatesNextVersionOnSecondCall(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	mustFieldDefTestProject(t, projects, "proj-1", "owner-1")
	svc := fielddefinitionApp.NewService(repo, projects)
	ctx := context.Background()

	_, err := svc.Define(ctx, "owner-1", "proj-1", []fielddefinition.FieldDefinition{
		field(t, "fullName", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := svc.Define(ctx, "owner-1", "proj-1", []fielddefinition.FieldDefinition{
		field(t, "fullName", fielddefinition.FieldKindString, true),
		field(t, "isVerify", fielddefinition.FieldKindBoolean, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if second.Version != fielddefinition.InitialVersion+1 {
		t.Fatalf("expected version %d, got %d", fielddefinition.InitialVersion+1, second.Version)
	}
}

func TestService_Define_InvalidFieldsPropagatesError(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	mustFieldDefTestProject(t, projects, "proj-1", "owner-1")
	svc := fielddefinitionApp.NewService(repo, projects)
	ctx := context.Background()

	_, err := svc.Define(ctx, "owner-1", "proj-1", nil)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestService_Define_WrongOwnerDenied(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	mustFieldDefTestProject(t, projects, "proj-1", "owner-1")
	svc := fielddefinitionApp.NewService(repo, projects)
	ctx := context.Background()

	_, err := svc.Define(ctx, "owner-2", "proj-1", []fielddefinition.FieldDefinition{
		field(t, "fullName", fielddefinition.FieldKindString, true),
	})
	if apperror.CodeOf(err) != apperror.CodeTenantAccessDenied {
		t.Fatalf("expected CodeTenantAccessDenied, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetLatestAndGetVersion(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	mustFieldDefTestProject(t, projects, "proj-1", "owner-1")
	svc := fielddefinitionApp.NewService(repo, projects)
	ctx := context.Background()

	created, err := svc.Define(ctx, "owner-1", "proj-1", []fielddefinition.FieldDefinition{
		field(t, "email", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest, err := svc.GetLatest(ctx, "owner-1", "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Version != created.Version {
		t.Fatalf("expected version %d, got %d", created.Version, latest.Version)
	}

	byVersion, err := svc.GetVersion(ctx, "owner-1", "proj-1", created.Version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if byVersion.ProjectID != created.ProjectID {
		t.Fatalf("expected ProjectID %q, got %q", created.ProjectID, byVersion.ProjectID)
	}
}

func TestService_GetLatest_WrongOwnerDenied(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	mustFieldDefTestProject(t, projects, "proj-1", "owner-1")
	svc := fielddefinitionApp.NewService(repo, projects)
	ctx := context.Background()

	if _, err := svc.Define(ctx, "owner-1", "proj-1", []fielddefinition.FieldDefinition{
		field(t, "email", fielddefinition.FieldKindString, true),
	}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err := svc.GetLatest(ctx, "owner-2", "proj-1")
	if apperror.CodeOf(err) != apperror.CodeTenantAccessDenied {
		t.Fatalf("expected CodeTenantAccessDenied, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetLatest_NotFound(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	svc := fielddefinitionApp.NewService(repo, projects)

	_, err := svc.GetLatest(context.Background(), "owner-1", "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}
