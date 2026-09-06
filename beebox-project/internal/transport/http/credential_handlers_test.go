package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	credentialapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	credentialmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/memory"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
	projectmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
	"github.com/gin-gonic/gin"
)

func newCredentialTestEngine() *gin.Engine {
	authSvc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	projects := projectmemory.NewProjectRepository()
	projectSvc := projectapp.NewService(projects)
	credentialSvc := credentialapp.NewService(credentialmemory.New(), projects)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())

	authGroup := engine.Group("/auth")
	RegisterAuthRoutes(authGroup, authSvc)

	dashboard := engine.Group("/dashboard")
	dashboard.Use(RequireOwnerSession(authSvc))
	RegisterProjectRoutes(dashboard, projectSvc)
	RegisterCredentialRoutes(dashboard, credentialSvc)

	return engine
}

func createProjectViaHTTP(t *testing.T, engine *gin.Engine, token string) createProjectResponse {
	t.Helper()
	body := `{"name":"Test Project"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected project creation to succeed, got %d, body: %s", rec.Code, rec.Body.String())
	}
	var resp createProjectResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal project response: %v", err)
	}
	return resp
}

func TestCreateCredential_Valid_ReturnsSecretOnce(t *testing.T) {
	engine := newCredentialTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	body := `{"environment":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects/"+proj.ID+"/credentials", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}

	var resp credentialWithSecretResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.SecretKey == "" {
		t.Fatal("expected non-empty secret_key on create")
	}
	if resp.PublicKey == "" {
		t.Fatal("expected non-empty public_key")
	}
}

func TestCreateCredential_WrongOwnerDenied(t *testing.T) {
	engine := newCredentialTestEngine()
	_, ownerOneToken := signUpAndSignIn(t, engine, "owner1@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, ownerOneToken)

	_, ownerTwoToken := signUpAndSignIn(t, engine, "owner2@example.com", "password123")

	body := `{"environment":"test"}`
	req := httptest.NewRequest(http.MethodPost, "/dashboard/projects/"+proj.ID+"/credentials", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+ownerTwoToken)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestGetCredential_DoesNotLeakSecret(t *testing.T) {
	engine := newCredentialTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh2@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	createBody := `{"environment":"test"}`
	createReq := httptest.NewRequest(http.MethodPost, "/dashboard/projects/"+proj.ID+"/credentials", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	engine.ServeHTTP(createRec, createReq)
	var created credentialWithSecretResponse
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}

	getReq := httptest.NewRequest(http.MethodGet, "/dashboard/credentials/"+created.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	engine.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, getRec.Code, getRec.Body.String())
	}
	if strings.Contains(getRec.Body.String(), "secret_key") {
		t.Fatalf("response must not contain secret_key field at all, got: %s", getRec.Body.String())
	}
}

func TestRotateCredential_ReturnsNewSecret_DifferentFromOld(t *testing.T) {
	engine := newCredentialTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh3@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	createBody := `{"environment":"live"}`
	createReq := httptest.NewRequest(http.MethodPost, "/dashboard/projects/"+proj.ID+"/credentials", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	engine.ServeHTTP(createRec, createReq)
	var created credentialWithSecretResponse
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}

	rotateReq := httptest.NewRequest(http.MethodPost, "/dashboard/credentials/"+created.ID+"/rotate", nil)
	rotateReq.Header.Set("Authorization", "Bearer "+token)
	rotateRec := httptest.NewRecorder()
	engine.ServeHTTP(rotateRec, rotateReq)

	if rotateRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rotateRec.Code, rotateRec.Body.String())
	}

	var rotated credentialWithSecretResponse
	if err := json.Unmarshal(rotateRec.Body.Bytes(), &rotated); err != nil {
		t.Fatalf("failed to unmarshal rotate response: %v", err)
	}
	if rotated.SecretKey == created.SecretKey {
		t.Fatal("expected a new secret_key after rotate")
	}
	if rotated.PublicKey != created.PublicKey {
		t.Fatal("expected public_key to remain unchanged after rotate")
	}
}

func TestRevokeCredential_ThenGetShowsRevokedStatus(t *testing.T) {
	engine := newCredentialTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh4@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	createBody := `{"environment":"test"}`
	createReq := httptest.NewRequest(http.MethodPost, "/dashboard/projects/"+proj.ID+"/credentials", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createRec := httptest.NewRecorder()
	engine.ServeHTTP(createRec, createReq)
	var created credentialWithSecretResponse
	if err := json.Unmarshal(createRec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to unmarshal create response: %v", err)
	}

	revokeReq := httptest.NewRequest(http.MethodPost, "/dashboard/credentials/"+created.ID+"/revoke", nil)
	revokeReq.Header.Set("Authorization", "Bearer "+token)
	revokeRec := httptest.NewRecorder()
	engine.ServeHTTP(revokeRec, revokeReq)
	if revokeRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, revokeRec.Code, revokeRec.Body.String())
	}

	getReq := httptest.NewRequest(http.MethodGet, "/dashboard/credentials/"+created.ID, nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getRec := httptest.NewRecorder()
	engine.ServeHTTP(getRec, getReq)

	var got credentialMetadataResponse
	if err := json.Unmarshal(getRec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to unmarshal get response: %v", err)
	}
	if got.Status != "REVOKED" {
		t.Fatalf("expected status REVOKED, got %q", got.Status)
	}
}