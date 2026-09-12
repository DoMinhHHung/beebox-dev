package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func newConfigurationRouter(repo *memory.ProjectRepository, organizationID string) *http.ServeMux {
	projects := project.NewService(repo)
	configurationRepo := memory.NewConfigurationRepository()
	configurations := applicationconfiguration.NewService(configurationRepo, projects, mustTestCatalog())
	return NewRouterWithConfiguration(projects, configurations, testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: organizationID}}, "test-internal-token")
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

func TestConfiguration_ApplyPublishedVersion(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	router := newConfigurationRouter(repo, "organization-1")
	body := map[string]any{"module_id": "auth", "module_version": "v1", "capability_id": "login", "capability_version": "v1"}
	created := doJSON(t, router, http.MethodPut, "/v1/projects/project-1/configuration", body)
	if created.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", created.Code, created.Body.String())
	}
	for _, status := range []string{"VALIDATED", "PUBLISHED"} {
		rec := doJSON(t, router, http.MethodPatch, "/v1/projects/project-1/configuration/versions/1", map[string]string{"status": status})
		if rec.Code != http.StatusOK {
			t.Fatalf("transition %s: %d %s", status, rec.Code, rec.Body.String())
		}
	}
	applied := doJSON(t, router, http.MethodPost, "/v1/projects/project-1/configuration/versions/1/apply", nil)
	if applied.Code != http.StatusOK {
		t.Fatalf("apply: %d %s", applied.Code, applied.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(applied.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["status"] != "APPLIED" {
		t.Fatalf("expected APPLIED, got %#v", payload["status"])
	}

	again := doJSON(t, router, http.MethodPost, "/v1/projects/project-1/configuration/versions/1/apply", nil)
	if again.Code != http.StatusOK {
		t.Fatalf("idempotent apply: %d %s", again.Code, again.Body.String())
	}
}

func TestConfiguration_ApplyRejectsUnauthorizedAndUnpublished(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	owner := newConfigurationRouter(repo, "organization-1")
	body := map[string]any{"module_id": "auth", "module_version": "v1", "capability_id": "login", "capability_version": "v1"}
	if rec := doJSON(t, owner, http.MethodPut, "/v1/projects/project-1/configuration", body); rec.Code != http.StatusCreated {
		t.Fatalf("create: %d", rec.Code)
	}

	draftApply := doJSON(t, owner, http.MethodPost, "/v1/projects/project-1/configuration/versions/1/apply", nil)
	if draftApply.Code != http.StatusConflict {
		t.Fatalf("draft apply expected 409, got %d %s", draftApply.Code, draftApply.Body.String())
	}

	other := newConfigurationRouter(repo, "organization-2")
	forbidden := doJSON(t, other, http.MethodPost, "/v1/projects/project-1/configuration/versions/1/apply", nil)
	if forbidden.Code != http.StatusForbidden && forbidden.Code != http.StatusNotFound {
		t.Fatalf("expected 403/404 for wrong org, got %d %s", forbidden.Code, forbidden.Body.String())
	}

	missing := doJSON(t, owner, http.MethodPost, "/v1/projects/project-1/configuration/versions/99/apply", nil)
	if missing.Code != http.StatusNotFound {
		t.Fatalf("missing version expected 404, got %d %s", missing.Code, missing.Body.String())
	}
}

func newConfigurationRouterWithToken(repo *memory.ProjectRepository, organizationID, internalToken string) *http.ServeMux {
	projects := project.NewService(repo)
	configurationRepo := memory.NewConfigurationRepository()
	configurations := applicationconfiguration.NewService(configurationRepo, projects, mustTestCatalog())
	return NewRouterWithConfiguration(projects, configurations, testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: organizationID}}, internalToken)
}

func TestInternalAppliedConfiguration_RequiresInternalToken(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	token := "runtime-internal-token"
	router := newConfigurationRouterWithToken(repo, "organization-1", token)

	missing := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/project-1/applied-configuration", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, missing)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token expected 401, got %d", rec.Code)
	}

	wrong := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/project-1/applied-configuration", nil)
	wrong.Header.Set("Authorization", "Bearer wrong-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, wrong)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong token expected 401, got %d", rec.Code)
	}

	scheme := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/project-1/applied-configuration", nil)
	scheme.Header.Set("Authorization", "Token "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, scheme)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong scheme expected 401, got %d", rec.Code)
	}

	// Developer identity session must not authenticate internal endpoint
	identity := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/project-1/applied-configuration", nil)
	identity.Header.Set("Authorization", "Bearer test-token")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, identity)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("developer session expected 401 on internal route, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), token) {
		t.Fatal("response must not contain internal token")
	}
}

func TestInternalAppliedConfiguration_ReturnsSnapshot(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.Transition(context.Background(), "project-1", domainproject.StatusActive); err != nil {
		t.Fatal(err)
	}

	configurationRepo := memory.NewConfigurationRepository()
	configurations := applicationconfiguration.NewService(configurationRepo, projects, mustTestCatalog())
	token := "runtime-internal-token"
	router := NewRouterWithConfiguration(projects, configurations, testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: "organization-1"}}, token)

	body := map[string]any{"module_id": "auth", "module_version": "v1", "capability_id": "login", "capability_version": "v1"}
	if rec := doJSON(t, router, http.MethodPut, "/v1/projects/project-1/configuration", body); rec.Code != http.StatusCreated {
		t.Fatalf("create: %d %s", rec.Code, rec.Body.String())
	}
	for _, status := range []string{"VALIDATED", "PUBLISHED"} {
		if rec := doJSON(t, router, http.MethodPatch, "/v1/projects/project-1/configuration/versions/1", map[string]string{"status": status}); rec.Code != http.StatusOK {
			t.Fatalf("transition %s: %d %s", status, rec.Code, rec.Body.String())
		}
	}
	if rec := doJSON(t, router, http.MethodPost, "/v1/projects/project-1/configuration/versions/1/apply", nil); rec.Code != http.StatusOK {
		t.Fatalf("apply: %d %s", rec.Code, rec.Body.String())
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/project-1/applied-configuration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", rec.Code, rec.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatal(err)
	}
	if payload["project_id"] != "project-1" {
		t.Fatalf("project_id %#v", payload["project_id"])
	}
	if payload["project_status"] != "ACTIVE" {
		t.Fatalf("project_status %#v", payload["project_status"])
	}
	if payload["applied_version"] != float64(1) {
		t.Fatalf("applied_version %#v", payload["applied_version"])
	}
	if payload["module_id"] != "auth" || payload["capability_id"] != "login" {
		t.Fatalf("unexpected payload %#v", payload)
	}
	if strings.Contains(rec.Body.String(), token) {
		t.Fatal("response must not contain internal token")
	}
}

func TestInternalAppliedConfiguration_NotFoundCases(t *testing.T) {
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	configurationRepo := memory.NewConfigurationRepository()
	configurations := applicationconfiguration.NewService(configurationRepo, projects, mustTestCatalog())
	token := "runtime-internal-token"
	router := NewRouterWithConfiguration(projects, configurations, testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: "organization-1"}}, token)

	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/missing/applied-configuration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("missing project expected 404, got %d %s", rec.Code, rec.Body.String())
	}

	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.Transition(context.Background(), "project-1", domainproject.StatusActive); err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/internal/v1/projects/project-1/applied-configuration", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("no applied config expected 404, got %d %s", rec.Code, rec.Body.String())
	}
}
