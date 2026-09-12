package http_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
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

type stubIdentity struct {
	session authcap.Session
	err     error
	token   string
}

func (s *stubIdentity) GetSession(_ context.Context, token string) (authcap.Session, error) {
	s.token = token
	return s.session, s.err
}

func sessionCFG() domain.AppliedConfiguration {
	return domain.AppliedConfiguration{
		ProjectID: "p1", AppliedVersion: 1, ModuleID: "beebox-auth", ModuleVersion: "v1",
		CapabilityID: "session", CapabilityVersion: "v1",
	}
}

func newRouter(creds stubCreds, configs stubConfigs, identity *stubIdentity) *http.ServeMux {
	return interfaceshttp.NewRouter(
		projectresolve.NewService(creds, configs),
		authcap.NewService(identity),
	)
}

func TestHealthz_NoCredential(t *testing.T) {
	router := newRouter(stubCreds{}, stubConfigs{}, &stubIdentity{})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestAuthSession_HappyPath(t *testing.T) {
	identity := &stubIdentity{session: authcap.Session{UserID: "u1", SessionID: "s1", OrganizationID: "o1"}}
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: sessionCFG()},
		identity,
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "project-key")
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
	var body map[string]any
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["project_id"] != "p1" || body["user_id"] != "u1" || body["session_id"] != "s1" {
		t.Fatalf("%#v", body)
	}
	if strings.Contains(rec.Body.String(), "user-session") || strings.Contains(rec.Body.String(), "project-key") {
		t.Fatal("must not echo credentials")
	}
	if identity.token != "user-session" {
		t.Fatalf("identity token %q", identity.token)
	}
}

func TestAuthSession_MissingProjectCredential(t *testing.T) {
	router := newRouter(stubCreds{}, stubConfigs{}, &stubIdentity{})
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestAuthSession_InvalidProjectCredential(t *testing.T) {
	router := newRouter(
		stubCreds{err: apperror.New(apperror.CodeUnauthenticated, "unauthenticated")},
		stubConfigs{},
		&stubIdentity{},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "bad")
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestAuthSession_WrongCapability(t *testing.T) {
	cfg := sessionCFG()
	cfg.CapabilityID = "password"
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: cfg},
		&stubIdentity{},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "key")
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("got %d %s", rec.Code, rec.Body.String())
	}
}

func TestAuthSession_MissingUserBearer(t *testing.T) {
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: sessionCFG()},
		&stubIdentity{},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "key")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestAuthSession_IdentityUnauthorized(t *testing.T) {
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: sessionCFG()},
		&stubIdentity{err: apperror.New(apperror.CodeUnauthenticated, "unauthenticated")},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "key")
	req.Header.Set("Authorization", "Bearer bad-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestAuthSession_IdentityUnavailable(t *testing.T) {
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: sessionCFG()},
		&stubIdentity{err: apperror.New(apperror.CodeDependencyFailure, "identity service unavailable")},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "key")
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("got %d", rec.Code)
	}
}

func TestAuthSession_ProjectContextNotOverriddenByUser(t *testing.T) {
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{cfg: sessionCFG()},
		&stubIdentity{session: authcap.Session{UserID: "u-from-b", SessionID: "s1", OrganizationID: "org-b"}},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "key")
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("got %d", rec.Code)
	}
	var body map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &body)
	if body["project_id"] != "p1" {
		t.Fatalf("project must stay p1, got %#v", body["project_id"])
	}
}

func TestAuthSession_NoAppliedConfig(t *testing.T) {
	router := newRouter(
		stubCreds{ctx: domain.ProjectContext{ProjectID: "p1", CredentialID: "c1"}},
		stubConfigs{err: apperror.New(apperror.CodeNotFound, "configuration not applied")},
		&stubIdentity{},
	)
	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "key")
	req.Header.Set("Authorization", "Bearer user-session")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("got %d", rec.Code)
	}
}
