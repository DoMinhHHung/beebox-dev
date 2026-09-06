package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
	projectmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
	"github.com/gin-gonic/gin"
)

func signUpAndSignIn(t *testing.T, engine *gin.Engine, email, password string) (ownerID, token string) {
	t.Helper()

	signUpBody := `{"email":"` + email + `","password":"` + password + `"}`
	signUpReq := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(signUpBody))
	signUpReq.Header.Set("Content-Type", "application/json")
	signUpRec := httptest.NewRecorder()
	engine.ServeHTTP(signUpRec, signUpReq)
	if signUpRec.Code != http.StatusCreated {
		t.Fatalf("expected sign-up to succeed, got %d, body: %s", signUpRec.Code, signUpRec.Body.String())
	}
	var signUpResp signUpResponse
	if err := json.Unmarshal(signUpRec.Body.Bytes(), &signUpResp); err != nil {
		t.Fatalf("failed to unmarshal sign-up response: %v", err)
	}

	signInBody := `{"email":"` + email + `","password":"` + password + `"}`
	signInReq := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(signInBody))
	signInReq.Header.Set("Content-Type", "application/json")
	signInRec := httptest.NewRecorder()
	engine.ServeHTTP(signInRec, signInReq)
	if signInRec.Code != http.StatusOK {
		t.Fatalf("expected sign-in to succeed, got %d, body: %s", signInRec.Code, signInRec.Body.String())
	}
	var signInResp signInResponse
	if err := json.Unmarshal(signInRec.Body.Bytes(), &signInResp); err != nil {
		t.Fatalf("failed to unmarshal sign-in response: %v", err)
	}

	return signUpResp.ID, signInResp.Token
}

func newProjectTestEngine() *gin.Engine {
	authSvc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	projectSvc := projectapp.NewService(projectmemory.NewProjectRepository())

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())

	authGroup := engine.Group("/auth")
	RegisterAuthRoutes(authGroup, authSvc)

	dashboard := engine.Group("/dashboard")
	dashboard.Use(RequireOwnerSession(authSvc))
	RegisterProjectRoutes(dashboard, projectSvc)

	return engine
}

func TestCreateProject_Valid_UsesOwnerIDFromToken(t *testing.T) {
	engine := newProjectTestEngine()
	ownerID, token := signUpAndSignIn(t, engine, "minh@example.com", "password123")

	body := `{"name":"My Project","tier":"pro"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp createProjectResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.OwnerID != ownerID {
		t.Fatalf("expected owner_id %q (from authenticated token), got %q", ownerID, resp.OwnerID)
	}
	if resp.Tier != "pro" {
		t.Fatalf("expected tier %q, got %q", "pro", resp.Tier)
	}
}

func TestCreateProject_DefaultTier(t *testing.T) {
	engine := newProjectTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh2@example.com", "password123")

	body := `{"name":"My Project"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp createProjectResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Tier != "free" {
		t.Fatalf("expected default tier %q, got %q", "free", resp.Tier)
	}
}

func TestCreateProject_NoToken(t *testing.T) {
	engine := newProjectTestEngine()

	body := `{"name":"My Project"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestCreateProject_InvalidBody(t *testing.T) {
	engine := newProjectTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh3@example.com", "password123")

	body := `{"name":""}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}