package identityclient_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/identityclient"
)

func TestGetSession_Success(t *testing.T) {
	var sawAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawAuth = r.Header.Get("Authorization")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"user_id": "u1", "session_id": "s1", "organization_id": "o1",
		})
	}))
	defer srv.Close()
	client := identityclient.New(srv.URL, time.Second, srv.Client())
	got, err := client.GetSession(context.Background(), "user-token")
	if err != nil {
		t.Fatal(err)
	}
	if got.UserID != "u1" || got.SessionID != "s1" {
		t.Fatalf("%#v", got)
	}
	if sawAuth != "Bearer user-token" {
		t.Fatalf("auth %q", sawAuth)
	}
}

func TestGetSession_UnauthorizedAndUnavailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()
	client := identityclient.New(srv.URL, time.Second, srv.Client())
	_, err := client.GetSession(context.Background(), "bad")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("got %v", err)
	}
	if strings.Contains(err.Error(), "bad") {
		t.Fatal("must not leak token")
	}
}
