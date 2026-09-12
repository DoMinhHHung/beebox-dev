package http

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

type testAuthenticator struct {
	principal auth.Principal
}

func (a testAuthenticator) Authenticate(context.Context, string) (auth.Principal, error) {
	return a.principal, nil
}

func newTestRouter() *http.ServeMux {
	return NewRouter(project.NewService(memory.NewProjectRepository()), "test-internal-token", testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: "organization-1"}})
}

func doJSON(t *testing.T, router *http.ServeMux, method string, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("failed to marshal request body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequestWithContext(context.Background(), method, path, reader)
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateProject_ReturnsCreated(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreateProject_RejectsInvalidBody(t *testing.T) {
	router := newTestRouter()

	req := httptest.NewRequestWithContext(context.Background(), "POST", "/v1/projects", bytes.NewReader([]byte("{not json")))
	req.Header.Set("Authorization", "Bearer test-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestCreateProject_RejectsDuplicateAsConflict(t *testing.T) {
	router := newTestRouter()
	body := map[string]string{"id": "project-1", "organization_id": "organization-1"}

	doJSON(t, router, "POST", "/v1/projects", body)
	rec := doJSON(t, router, "POST", "/v1/projects", body)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestGetProject_ReturnsProject(t *testing.T) {
	router := newTestRouter()
	doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})

	rec := doJSON(t, router, "GET", "/v1/projects/project-1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetProject_ReturnsNotFound(t *testing.T) {
	router := newTestRouter()

	rec := doJSON(t, router, "GET", "/v1/projects/missing", nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}

func TestTransitionProject_AppliesValidTransition(t *testing.T) {
	router := newTestRouter()
	doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})

	rec := doJSON(t, router, "PATCH", "/v1/projects/project-1", map[string]string{"status": "ACTIVE"})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTransitionProject_RejectsInvalidTransitionAsConflict(t *testing.T) {
	router := newTestRouter()
	doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})

	rec := doJSON(t, router, "PATCH", "/v1/projects/project-1", map[string]string{"status": "ARCHIVED"})
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestTransitionProject_RejectsUnknownStatusAsValidationError(t *testing.T) {
	router := newTestRouter()
	doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})

	rec := doJSON(t, router, "PATCH", "/v1/projects/project-1", map[string]string{"status": "NOT_A_STATUS"})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestArchiveProject_ArchivesActiveProject(t *testing.T) {
	router := newTestRouter()
	doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})
	doJSON(t, router, "PATCH", "/v1/projects/project-1", map[string]string{"status": "ACTIVE"})

	rec := doJSON(t, router, "DELETE", "/v1/projects/project-1", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestArchiveProject_RejectsSecondArchiveAsConflict(t *testing.T) {
	router := newTestRouter()
	doJSON(t, router, "POST", "/v1/projects", map[string]string{"id": "project-1", "organization_id": "organization-1"})
	doJSON(t, router, "PATCH", "/v1/projects/project-1", map[string]string{"status": "ACTIVE"})
	doJSON(t, router, "DELETE", "/v1/projects/project-1", nil)

	rec := doJSON(t, router, "DELETE", "/v1/projects/project-1", nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}
