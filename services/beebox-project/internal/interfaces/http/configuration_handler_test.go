package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func newConfigurationRouter(repo *memory.ProjectRepository, organizationID string) *http.ServeMux {
	projects := project.NewService(repo)
	configurationRepo := memory.NewConfigurationRepository()
	configurations := applicationconfiguration.NewService(configurationRepo, projects, mustTestCatalog())
	return NewRouterWithConfiguration(projects, configurations, testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: organizationID}})
}

func TestConfiguration_UnauthenticatedReturns401(t *testing.T) {
	router := newConfigurationRouter(memory.NewProjectRepository(), "organization-1")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/projects/project-1/configuration", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestConfiguration_CreateAndReadCurrent(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	router := newConfigurationRouter(repo, "organization-1")
	body := map[string]any{"module_id": "auth", "module_version": "v1", "capability_id": "login", "capability_version": "v1"}
	created := doJSON(t, router, http.MethodPut, "/v1/projects/project-1/configuration", body)
	if created.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", created.Code, created.Body.String())
	}
	current := doJSON(t, router, http.MethodGet, "/v1/projects/project-1/configuration", nil)
	if current.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", current.Code, current.Body.String())
	}
}

func TestConfiguration_RejectsUnauthorizedOrganization(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	router := newConfigurationRouter(repo, "organization-2")
	rec := doJSON(t, router, http.MethodPut, "/v1/projects/project-1/configuration", map[string]string{"module_id": "auth", "module_version": "v1", "capability_id": "login", "capability_version": "v1"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestConfiguration_RejectsMalformedAndInvalidBody(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	router := newConfigurationRouter(repo, "organization-1")
	req := httptest.NewRequestWithContext(context.Background(), http.MethodPut, "/v1/projects/project-1/configuration", strings.NewReader("{bad"))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed body, got %d", rec.Code)
	}
	invalid := doJSON(t, router, http.MethodPut, "/v1/projects/project-1/configuration", map[string]string{"module_id": ""})
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for invalid config, got %d", invalid.Code)
	}
}

func mustTestCatalog() catalog.Catalog {
	cat, err := catalog.New(
		[]catalog.ModuleDefinition{{ID: "auth", Version: "v1"}},
		[]catalog.CapabilityDefinition{{ModuleID: "auth", ID: "login", Version: "v1"}},
		nil,
	)
	if err != nil {
		panic(err)
	}
	return cat
}
