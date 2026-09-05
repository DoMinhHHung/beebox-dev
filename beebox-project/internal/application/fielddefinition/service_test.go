package fielddefinition_test

import (
	"context"
	"testing"

	fielddefinitionApp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/fielddefinition/memory"
)

func field(t *testing.T, name string, kind fielddefinition.FieldKind, required bool) fielddefinition.FieldDefinition {
	t.Helper()
	f, err := fielddefinition.NewFieldDefinition(name, kind, required)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return f
}

func TestService_Define_CreatesInitialVersion(t *testing.T) {
	repo := memory.New()
	svc := fielddefinitionApp.NewService(repo)
	ctx := context.Background()

	schema, err := svc.Define(ctx, "proj-1", []fielddefinition.FieldDefinition{
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
	svc := fielddefinitionApp.NewService(repo)
	ctx := context.Background()

	_, err := svc.Define(ctx, "proj-1", []fielddefinition.FieldDefinition{
		field(t, "fullName", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second, err := svc.Define(ctx, "proj-1", []fielddefinition.FieldDefinition{
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
	svc := fielddefinitionApp.NewService(repo)
	ctx := context.Background()

	_, err := svc.Define(ctx, "proj-1", nil)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetLatestAndGetVersion(t *testing.T) {
	repo := memory.New()
	svc := fielddefinitionApp.NewService(repo)
	ctx := context.Background()

	created, err := svc.Define(ctx, "proj-1", []fielddefinition.FieldDefinition{
		field(t, "email", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	latest, err := svc.GetLatest(ctx, "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if latest.Version != created.Version {
		t.Fatalf("expected version %d, got %d", created.Version, latest.Version)
	}

	byVersion, err := svc.GetVersion(ctx, "proj-1", created.Version)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if byVersion.ProjectID != created.ProjectID {
		t.Fatalf("expected ProjectID %q, got %q", created.ProjectID, byVersion.ProjectID)
	}
}

func TestService_GetLatest_NotFound(t *testing.T) {
	repo := memory.New()
	svc := fielddefinitionApp.NewService(repo)

	_, err := svc.GetLatest(context.Background(), "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}