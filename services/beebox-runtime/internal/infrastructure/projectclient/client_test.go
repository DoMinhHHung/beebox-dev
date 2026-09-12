package projectclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/projectclient"
)

func TestVerifyPublic_SuccessAndAuthHeader(t *testing.T) {
	var sawAuth, sawBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		sawBody = string(buf[:n])
		_ = json.NewEncoder(w).Encode(map[string]string{
			"project_id":    "p1",
			"credential_id": "c1",
			"kind":          "PUBLIC",
		})
	}))
	defer srv.Close()

	client := projectclient.New(srv.URL, "internal-token", time.Second, srv.Client())
	got, err := client.VerifyPublic(context.Background(), "p1", "raw-public")
	if err != nil {
		t.Fatal(err)
	}
	if got.ProjectID != "p1" || got.CredentialID != "c1" {
		t.Fatalf("%#v", got)
	}
	if sawAuth != "Bearer internal-token" {
		t.Fatalf("auth %q", sawAuth)
	}
	if !strings.Contains(sawBody, "raw-public") {
		t.Fatalf("body %q", sawBody)
	}
}

func TestVerifyPublic_InvalidCredential(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	client := projectclient.New(srv.URL, "tok", time.Second, srv.Client())
	_, err := client.VerifyPublic(context.Background(), "p1", "bad")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("got %v", err)
	}
	if strings.Contains(err.Error(), "bad") {
		t.Fatal("must not leak credential")
	}
}

func TestVerifyPublic_ProjectMismatchInResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]string{
			"project_id":    "other",
			"credential_id": "c1",
			"kind":          "PUBLIC",
		})
	}))
	defer srv.Close()
	client := projectclient.New(srv.URL, "tok", time.Second, srv.Client())
	_, err := client.VerifyPublic(context.Background(), "p1", "raw")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("got %v", err)
	}
}

func TestVerifyPublic_Unavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer srv.Close()
	client := projectclient.New(srv.URL, "tok", time.Second, srv.Client())
	_, err := client.VerifyPublic(context.Background(), "p1", "raw")
	if apperror.CodeOf(err) != apperror.CodeDependencyFailure {
		t.Fatalf("got %v", err)
	}
}

func TestGetAppliedConfiguration_Success(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"project_id":         "p1",
			"project_status":     "ACTIVE",
			"applied_version":    2,
			"module_id":          "beebox-auth",
			"module_version":     "v1",
			"capability_id":      "password",
			"capability_version": "v1",
			"data_fields":        []any{},
		})
	}))
	defer srv.Close()
	client := projectclient.New(srv.URL, "internal-token", time.Second, srv.Client())
	got, err := client.GetAppliedConfiguration(context.Background(), "p1")
	if err != nil {
		t.Fatal(err)
	}
	if got.AppliedVersion != 2 || got.ModuleID != "beebox-auth" {
		t.Fatalf("%#v", got)
	}
	if sawAuth != "Bearer internal-token" {
		t.Fatalf("auth %q", sawAuth)
	}
}

func TestGetAppliedConfiguration_NotFoundAndForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()
	client := projectclient.New(srv.URL, "tok", time.Second, srv.Client())
	_, err := client.GetAppliedConfiguration(context.Background(), "p1")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("got %v", err)
	}
}

func TestGetAppliedConfiguration_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("{bad"))
	}))
	defer srv.Close()
	client := projectclient.New(srv.URL, "tok", time.Second, srv.Client())
	_, err := client.GetAppliedConfiguration(context.Background(), "p1")
	if apperror.CodeOf(err) != apperror.CodeDependencyFailure {
		t.Fatalf("got %v", err)
	}
}
