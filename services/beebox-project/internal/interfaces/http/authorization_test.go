package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func TestProjectAccess_RequiresAuthentication(t *testing.T) {
	router := NewRouter(project.NewService(memory.NewProjectRepository()), testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: "organization-1"}})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/projects/project-1", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestProjectAccess_RejectsUnauthorizedOrganization(t *testing.T) {
	repo := memory.NewProjectRepository()
	ownerRouter := NewRouter(project.NewService(repo), testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: "organization-1"}})
	doJSON(t, ownerRouter, http.MethodPost, "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})

	otherRouter := NewRouter(project.NewService(repo), testAuthenticator{principal: auth.Principal{UserID: "user-2", OrganizationID: "organization-2"}})
	req := httptest.NewRequestWithContext(context.Background(), http.MethodGet, "/v1/projects/project-1", nil)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	otherRouter.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestProjectCreation_DerivesOrganizationFromPrincipal(t *testing.T) {
	router := newTestRouter()
	rec := doJSON(t, router, http.MethodPost, "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-2"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		OrganizationID string `json:"organization_id"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body.OrganizationID != "organization-1" {
		t.Fatalf("expected organization-1, got %q", body.OrganizationID)
	}
}
