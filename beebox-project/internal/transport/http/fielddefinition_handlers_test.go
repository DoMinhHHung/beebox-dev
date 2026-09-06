package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	fielddefinitionapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	fielddefinitionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/fielddefinition/memory"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
	projectmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
	"github.com/gin-gonic/gin"
)

func newFieldDefinitionTestEngine() *gin.Engine {
	authSvc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	projects := projectmemory.NewProjectRepository()
	projectSvc := projectapp.NewService(projects)
	fieldSvc := fielddefinitionapp.NewService(fielddefinitionmemory.New(), projects)

	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())

	authGroup := engine.Group("/auth")
	RegisterAuthRoutes(authGroup, authSvc)

	dashboard := engine.Group("/dashboard")
	dashboard.Use(RequireOwnerSession(authSvc))
	RegisterProjectRoutes(dashboard, projectSvc)
	RegisterFieldDefinitionRoutes(dashboard, fieldSvc)

	return engine
}

func defineFieldsViaHTTP(t *testing.T, engine *gin.Engine, token, projectID, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodPut, "/dashboard/projects/"+projectID+"/fields", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestDefineFields_CreatesInitialVersion(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	body := `{"fields":[{"name":"fullName","kind":"STRING","required":true}]}`
	rec := defineFieldsViaHTTP(t, engine, token, proj.ID, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp schemaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Version != 1 {
		t.Fatalf("expected version 1, got %d", resp.Version)
	}
}

func TestDefineFields_SecondCallCreatesNextVersion(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh2@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	firstBody := `{"fields":[{"name":"fullName","kind":"STRING","required":true}]}`
	firstRec := defineFieldsViaHTTP(t, engine, token, proj.ID, firstBody)
	if firstRec.Code != http.StatusOK {
		t.Fatalf("expected first define to succeed, got %d, body: %s", firstRec.Code, firstRec.Body.String())
	}

	secondBody := `{"fields":[{"name":"fullName","kind":"STRING","required":true},{"name":"isVerify","kind":"BOOLEAN","required":true}]}`
	secondRec := defineFieldsViaHTTP(t, engine, token, proj.ID, secondBody)
	if secondRec.Code != http.StatusOK {
		t.Fatalf("expected second define to succeed, got %d, body: %s", secondRec.Code, secondRec.Body.String())
	}

	var resp schemaResponse
	if err := json.Unmarshal(secondRec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Version != 2 {
		t.Fatalf("expected version 2, got %d", resp.Version)
	}
}

func TestDefineFields_WrongOwnerDenied(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, ownerOneToken := signUpAndSignIn(t, engine, "owner1@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, ownerOneToken)

	_, ownerTwoToken := signUpAndSignIn(t, engine, "owner2@example.com", "password123")

	body := `{"fields":[{"name":"fullName","kind":"STRING","required":true}]}`
	rec := defineFieldsViaHTTP(t, engine, ownerTwoToken, proj.ID, body)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusForbidden, rec.Code, rec.Body.String())
	}
}

func TestDefineFields_InvalidKind(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh3@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	body := `{"fields":[{"name":"birthDate","kind":"DATE","required":true}]}`
	rec := defineFieldsViaHTTP(t, engine, token, proj.ID, body)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}

func TestGetLatestFields_ReturnsMostRecentVersion(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh4@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	defineFieldsViaHTTP(t, engine, token, proj.ID, `{"fields":[{"name":"email","kind":"STRING","required":true}]}`)
	defineFieldsViaHTTP(t, engine, token, proj.ID, `{"fields":[{"name":"email","kind":"STRING","required":true},{"name":"age","kind":"NUMBER","required":false}]}`)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/projects/"+proj.ID+"/fields", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp schemaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Version != 2 || len(resp.Fields) != 2 {
		t.Fatalf("expected version 2 with 2 fields, got version %d with %d fields", resp.Version, len(resp.Fields))
	}
}

func TestGetVersionFields_ReturnsSpecificVersion(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh5@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	defineFieldsViaHTTP(t, engine, token, proj.ID, `{"fields":[{"name":"email","kind":"STRING","required":true}]}`)
	defineFieldsViaHTTP(t, engine, token, proj.ID, `{"fields":[{"name":"email","kind":"STRING","required":true},{"name":"age","kind":"NUMBER","required":false}]}`)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/projects/"+proj.ID+"/fields/1", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp schemaResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Version != 1 || len(resp.Fields) != 1 {
		t.Fatalf("expected version 1 with 1 field, got version %d with %d fields", resp.Version, len(resp.Fields))
	}
}

func TestGetVersionFields_InvalidVersionParam(t *testing.T) {
	engine := newFieldDefinitionTestEngine()
	_, token := signUpAndSignIn(t, engine, "minh6@example.com", "password123")
	proj := createProjectViaHTTP(t, engine, token)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/projects/"+proj.ID+"/fields/abc", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusBadRequest, rec.Code, rec.Body.String())
	}
}
