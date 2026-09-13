package http

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	applicationcredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func newCredentialRouter(t *testing.T, organizationID, internalToken string) (*http.ServeMux, *project.Service, *applicationcredential.Service) {
	t.Helper()
	repo := memory.NewProjectRepository()
	projects := project.NewService(repo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.Transition(context.Background(), "project-1", domainproject.StatusActive); err != nil {
		t.Fatal(err)
	}
	credentials := applicationcredential.NewService(memory.NewCredentialRepository(), projects)
	router := NewRouterWithServices(
		projects,
		nil,
		nil,
		credentials,
		testAuthenticator{principal: auth.Principal{UserID: "user-1", OrganizationID: organizationID}},
		internalToken,
	)
	return router, projects, credentials
}

func TestCredential_IssuePublicAndVerifyInternal(t *testing.T) {
	token := "runtime-internal-token"
	router, _, _ := newCredentialRouter(t, "organization-1", token)

	issued := doJSON(t, router, http.MethodPost, "/v1/projects/project-1/credentials/public", map[string]string{"name": "frontend"})
	if issued.Code != http.StatusCreated {
		t.Fatalf("issue: %d %s", issued.Code, issued.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(issued.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	raw, _ := body["credential"].(string)
	if raw == "" {
		t.Fatal("expected raw credential in issue response")
	}
	if body["kind"] != "PUBLIC" {
		t.Fatalf("kind %#v", body["kind"])
	}
	if strings.Contains(issued.Body.String(), "secret_hash") {
		t.Fatal("must not expose secret_hash")
	}

	req := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/v1/projects/project-1/credentials/verify", strings.NewReader(`{"credential":"`+raw+`"}`))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("verify: %d %s", rec.Code, rec.Body.String())
	}

	wrong := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/v1/projects/project-1/credentials/verify", strings.NewReader(`{"credential":"bad"}`))
	wrong.Header.Set("Authorization", "Bearer "+token)
	wrong.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, wrong)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("wrong credential expected 401, got %d", rec.Code)
	}
}

func TestCredential_InternalVerifyAuthBoundary(t *testing.T) {
	token := "runtime-internal-token"
	router, _, svc := newCredentialRouter(t, "organization-1", token)
	issued, err := svc.IssuePublic(context.Background(), "project-1", "organization-1", "frontend")
	if err != nil {
		t.Fatal(err)
	}

	missing := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/v1/projects/project-1/credentials/verify", strings.NewReader(`{"credential":"`+issued.Secret+`"}`))
	missing.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, missing)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("missing token expected 401, got %d", rec.Code)
	}

	developer := httptest.NewRequestWithContext(context.Background(), http.MethodPost, "/internal/v1/projects/project-1/credentials/verify", strings.NewReader(`{"credential":"`+issued.Secret+`"}`))
	developer.Header.Set("Authorization", "Bearer test-token")
	developer.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	router.ServeHTTP(rec, developer)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("developer session expected 401, got %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), token) || strings.Contains(rec.Body.String(), issued.Secret) {
		t.Fatal("response must not leak tokens or credentials")
	}
}

func TestCredential_IssueRejectsWrongOrg(t *testing.T) {
	router, _, _ := newCredentialRouter(t, "organization-2", "runtime-internal-token")
	rec := doJSON(t, router, http.MethodPost, "/v1/projects/project-1/credentials/public", map[string]string{"name": "frontend"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d %s", rec.Code, rec.Body.String())
	}
}
