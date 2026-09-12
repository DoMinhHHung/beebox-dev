package configuration_test

import (
	"context"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func newService(t *testing.T) (*applicationconfiguration.Service, *project.Service) {
	t.Helper()
	projectsRepo := memory.NewProjectRepository()
	projects := project.NewService(projectsRepo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	return applicationconfiguration.NewService(memory.NewConfigurationRepository(), projects), projects
}

func validConfiguration(t *testing.T, projectID string) domainconfiguration.Configuration {
	t.Helper()
	config, err := domainconfiguration.New(projectID, "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("create config: %v", err)
	}
	return config
}

func TestCreateOrUpdate_CreatesVersions(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	first, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil || first.Number != 1 {
		t.Fatalf("expected version 1, got %#v, %v", first, err)
	}
	second, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil || second.Number != 2 {
		t.Fatalf("expected version 2, got %#v, %v", second, err)
	}
}

func TestCreateOrUpdate_RejectsInvalidConfiguration(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.CreateOrUpdate(context.Background(), "project-1", "organization-1", domainconfiguration.Configuration{ProjectID: "project-1"})
	if apperror.CodeOf(err) != apperror.CodeValidation {
		t.Fatalf("expected validation, got %v", apperror.CodeOf(err))
	}
}

func TestLifecycleAndRollout(t *testing.T) {
	svc, _ := newService(t)
	ctx := context.Background()
	version, err := svc.CreateOrUpdate(ctx, "project-1", "organization-1", validConfiguration(t, "project-1"))
	if err != nil {
		t.Fatal(err)
	}
	version, err = svc.Transition(ctx, "project-1", "organization-1", version.Number, domainconfiguration.StatusValidated)
	if err != nil {
		t.Fatal(err)
	}
	version, err = svc.Transition(ctx, "project-1", "organization-1", version.Number, domainconfiguration.StatusPublished)
	if err != nil {
		t.Fatal(err)
	}
	state, err := svc.Rollout(ctx, "project-1", "organization-1", version.Number)
	if err != nil || state.DesiredVersion != version.Number {
		t.Fatalf("unexpected rollout %#v, %v", state, err)
	}
}

func TestAccessAndNotFound(t *testing.T) {
	svc, _ := newService(t)
	_, err := svc.Current(context.Background(), "missing", "organization-1")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected project not found, got %v", apperror.CodeOf(err))
	}
	_, err = svc.Current(context.Background(), "project-1", "organization-2")
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("expected forbidden, got %v", apperror.CodeOf(err))
	}
}
