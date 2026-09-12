package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/interfaces/http"
)

type stubCreds struct {
	ctx domain.ProjectContext
	err error
}

func (s stubCreds) VerifyPublic(context.Context, string, string) (domain.ProjectContext, error) {
	return s.ctx, s.err
}

type stubConfigs struct {
	cfg domain.AppliedConfiguration
	err error
}

func (s stubConfigs) GetAppliedConfiguration(context.Context, string) (domain.AppliedConfiguration, error) {
	return s.cfg, s.err
}

func TestHealthz_NoCredential(t *testing.T) {
	router := interfaceshttp.NewRouter(projectresolve.NewService(stubCreds{}, stubConfigs{}))
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestProjectReady_RequiresCredential(t *testing.T) {
	router := interfaceshttp.NewRouter(projectresolve.NewService(stubCreds{}, stubConfigs{}))
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/_ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestProjectReady_Success(t *testing.T) {
	router := interfaceshttp.NewRouter(projectresolve.NewService(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: domain.AppliedConfiguration{ProjectID: "p1", AppliedVersion: 3, ModuleID: "beebox-auth", CapabilityID: "password"}},
	))
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/_ready", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "raw-key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["project_id"] != "p1" || body["applied_version"] != float64(3) {
		t.Fatalf("%#v", body)
	}
}

func TestProjectReady_InvalidCredential(t *testing.T) {
	router := interfaceshttp.NewRouter(projectresolve.NewService(
		stubCreds{err: apperror.New(apperror.CodeUnauthenticated, "unauthenticated")},
		stubConfigs{},
	))
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/_ready", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "bad")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestProjectReady_ServiceUnavailable(t *testing.T) {
	router := interfaceshttp.NewRouter(projectresolve.NewService(
		stubCreds{err: apperror.New(apperror.CodeDependencyFailure, "project service unavailable")},
		stubConfigs{},
	))
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/_ready", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "raw")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestProjectReady_NoAppliedConfig(t *testing.T) {
	router := interfaceshttp.NewRouter(projectresolve.NewService(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{err: apperror.New(apperror.CodeNotFound, "configuration not applied")},
	))
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/_ready", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "raw")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestProjectReady_RejectsInternalTokenAsProjectCredentialSemantics(t *testing.T) {
	// Internal token is never accepted as project credential: verifier returns unauthenticated
	router := interfaceshttp.NewRouter(projectresolve.NewService(
		stubCreds{err: apperror.New(apperror.CodeUnauthenticated, "unauthenticated")},
		stubConfigs{},
	))
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/_ready", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "internal-service-token")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}
