package http_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/identityclient"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/projectclient"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/interfaces/http"
)

func TestIntegration_AuthSessionAgainstFakeServices(t *testing.T) {
	projectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer internal-token" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/internal/v1/projects/p1/credentials/verify":
			_ = json.NewEncoder(w).Encode(map[string]string{
				"project_id": "p1", "credential_id": "c1", "kind": "PUBLIC",
			})
		case r.Method == http.MethodGet && r.URL.Path == "/internal/v1/projects/p1/applied-configuration":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"project_id": "p1", "project_status": "ACTIVE", "applied_version": 1,
				"module_id": "beebox-auth", "module_version": "v1",
				"capability_id": "session", "capability_version": "v1",
				"data_fields": []any{},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer projectSrv.Close()

	identitySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer user-session" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{
			"user_id": "u1", "session_id": "s1", "organization_id": "o1",
		})
	}))
	defer identitySrv.Close()

	projectHTTP := projectclient.New(projectSrv.URL, "internal-token", time.Second, projectSrv.Client())
	identityHTTP := identityclient.New(identitySrv.URL, time.Second, identitySrv.Client())
	router := interfaceshttp.NewRouter(
		projectresolve.NewService(projectHTTP, projectHTTP),
		authcap.NewService(identityHTTP),
	)

	req := httptest.NewRequest(http.MethodGet, "/v1/p/p1/auth/session", nil)
	req.Header.Set("X-BeeBox-Project-Credential", "public-key")
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
	if body["project_id"] != "p1" || body["user_id"] != "u1" {
		t.Fatalf("%#v", body)
	}
}

func TestIntegration_WrongProjectCredentialRejected(t *testing.T) {
	projectSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer projectSrv.Close()
	identitySrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("identity must not be called")
	}))
	defer identitySrv.Close()

	projectHTTP := projectclient.New(projectSrv.URL, "internal-token", time.Second, projectSrv.Client())
	identityHTTP := identityclient.New(identitySrv.URL, time.Second, identitySrv.Client())
	router := interfaceshttp.NewRouter(
		projectresolve.NewService(projectHTTP, projectHTTP),
		authcap.NewService(identityHTTP),
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
