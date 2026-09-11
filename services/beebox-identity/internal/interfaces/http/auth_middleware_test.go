package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
)

type middlewareSessionRepo struct {
	findValue session.Session
	findErr   error
	findID    string
}

func (f *middlewareSessionRepo) Create(context.Context, session.Session) error {
	return nil
}

func (f *middlewareSessionRepo) FindByID(_ context.Context, id string) (session.Session, error) {
	f.findID = id
	return f.findValue, f.findErr
}

func (f *middlewareSessionRepo) Revoke(context.Context, session.Session) error {
	return nil
}

func TestExtractBearerToken(t *testing.T) {
	cases := []struct {
		name    string
		header  string
		wantErr bool
		token   string
	}{
		{"missing", "", true, ""},
		{"wrong scheme", "Basic abc", true, ""},
		{"empty bearer", "Bearer ", true, ""},
		{"valid", "Bearer tok-123", false, "tok-123"},
		{"case insensitive scheme", "bearer tok-123", false, "tok-123"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			token, err := extractBearerToken(tc.header)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if token != tc.token {
				t.Fatalf("expected %q, got %q", tc.token, token)
			}
		})
	}
}

func TestRequireAuthenticationValidToken(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	token := "opaque-session-token-for-middleware-test-aaaa"
	userID, _ := identity.NewIdentifier("user-1")
	s, _ := session.New(authHashToken(token), userID, now.Add(-time.Hour), now.Add(time.Hour))
	repo := &middlewareSessionRepo{findValue: s}
	authenticator := auth.NewAuthenticateSessionService(repo, &httpFakeClock{now: now})

	var gotUser string
	var gotSession string
	var sawRawToken bool
	next := func(w http.ResponseWriter, r *http.Request) {
		authn, ok := AuthenticationFromContext(r.Context())
		if !ok {
			t.Fatal("expected authentication context")
		}
		gotUser = authn.UserID.String()
		gotSession = authn.SessionID
		if authn.SessionID == token {
			sawRawToken = true
		}
		w.WriteHeader(http.StatusNoContent)
	}

	handler := requireAuthentication(authenticator, next)
	req := httptest.NewRequest(http.MethodPost, "/auth/signout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d body=%s", rec.Code, rec.Body.String())
	}
	if gotUser != "user-1" {
		t.Fatalf("unexpected user %q", gotUser)
	}
	if gotSession != authHashToken(token) {
		t.Fatal("expected storage session id in context")
	}
	if sawRawToken {
		t.Fatal("raw token must not be stored in authentication context")
	}
}

func TestRequireAuthenticationMissingHeader(t *testing.T) {
	authenticator := auth.NewAuthenticateSessionService(&middlewareSessionRepo{findErr: auth.ErrNotFound}, &httpFakeClock{now: time.Now().UTC()})
	handler := requireAuthentication(authenticator, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/signout", nil)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
	assertErrorCode(t, rec, "UNAUTHENTICATED")
}

func TestRequireAuthenticationInvalidToken(t *testing.T) {
	authenticator := auth.NewAuthenticateSessionService(&middlewareSessionRepo{findErr: auth.ErrNotFound}, &httpFakeClock{now: time.Now().UTC()})
	handler := requireAuthentication(authenticator, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/signout", nil)
	req.Header.Set("Authorization", "Bearer missing")
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthenticationExpiredSession(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	token := "token"
	userID, _ := identity.NewIdentifier("user-1")
	s, _ := session.New(authHashToken(token), userID, now.Add(-2*time.Hour), now.Add(-time.Hour))
	authenticator := auth.NewAuthenticateSessionService(&middlewareSessionRepo{findValue: s}, &httpFakeClock{now: now})
	handler := requireAuthentication(authenticator, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/signout", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestRequireAuthenticationRepositoryFailure(t *testing.T) {
	authenticator := auth.NewAuthenticateSessionService(&middlewareSessionRepo{findErr: errors.New("db")}, &httpFakeClock{now: time.Now().UTC()})
	handler := requireAuthentication(authenticator, func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("next must not run")
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/signout", nil)
	req.Header.Set("Authorization", "Bearer token")
	rec := httptest.NewRecorder()
	handler(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("expected 502, got %d body=%s", rec.Code, rec.Body.String())
	}
	var env errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if env.Error.Code != "DEPENDENCY_FAILURE" {
		t.Fatalf("expected DEPENDENCY_FAILURE, got %s", env.Error.Code)
	}
}

func TestProtectedSignOutWithoutToken(t *testing.T) {
	now := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	authenticator := auth.NewAuthenticateSessionService(&middlewareSessionRepo{findErr: auth.ErrNotFound}, &httpFakeClock{now: now})
	revoker := auth.NewRevokeSessionService(&middlewareSessionRepo{}, &httpFakeClock{now: now})
	mux := NewRouter(Dependencies{
		AuthenticateSession: authenticator,
		RevokeSession:       revoker,
	})
	req := httptest.NewRequest(http.MethodPost, "/auth/signout", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rec.Code)
	}
}

func TestPublicSignUpRemainsOpen(t *testing.T) {
	mux := NewRouter(Dependencies{})
	req := httptest.NewRequest(http.MethodPost, "/auth/signup", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code == http.StatusUnauthorized {
		t.Fatal("signup must remain public")
	}
}
