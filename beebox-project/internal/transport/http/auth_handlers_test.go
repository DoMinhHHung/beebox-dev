package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
	"github.com/gin-gonic/gin"
)

func newAuthRoutesEngine(svc *auth.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())
	authGroup := engine.Group("/auth")
	RegisterAuthRoutes(authGroup, svc)
	return engine
}

func TestSignUp_Valid(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	body := `{"email":"minh@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusCreated, rec.Code, rec.Body.String())
	}
	if strings.Contains(rec.Body.String(), "password") {
		t.Fatalf("response must not leak password/password_hash, got: %s", rec.Body.String())
	}

	var resp signUpResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Email != "minh@example.com" {
		t.Fatalf("expected email %q, got %q", "minh@example.com", resp.Email)
	}
}

func TestSignUp_DuplicateEmail(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	body := `{"email":"minh@example.com","password":"password123"}`

	req1 := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	rec1 := httptest.NewRecorder()
	engine.ServeHTTP(rec1, req1)
	if rec1.Code != http.StatusCreated {
		t.Fatalf("expected first sign-up to succeed, got %d", rec1.Code)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")
	rec2 := httptest.NewRecorder()
	engine.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusConflict {
		t.Fatalf("expected status %d, got %d", http.StatusConflict, rec2.Code)
	}
}

func TestSignUp_InvalidBody(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	req := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader("not-json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rec.Code)
	}
}

func TestSignIn_Valid_ReturnsToken(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	signUpBody := `{"email":"minh@example.com","password":"password123"}`
	signUpReq := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(signUpBody))
	signUpReq.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(httptest.NewRecorder(), signUpReq)

	signInBody := `{"email":"minh@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(signInBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}

	var resp signInResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestSignIn_WrongPassword(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	signUpBody := `{"email":"minh@example.com","password":"password123"}`
	signUpReq := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(signUpBody))
	signUpReq.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(httptest.NewRecorder(), signUpReq)

	signInBody := `{"email":"minh@example.com","password":"wrong-password"}`
	req := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(signInBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestSignOut_WithValidToken(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	signUpBody := `{"email":"minh@example.com","password":"password123"}`
	signUpReq := httptest.NewRequest(http.MethodPost, "/auth/sign-up", strings.NewReader(signUpBody))
	signUpReq.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(httptest.NewRecorder(), signUpReq)

	signInBody := `{"email":"minh@example.com","password":"password123"}`
	signInReq := httptest.NewRequest(http.MethodPost, "/auth/sign-in", strings.NewReader(signInBody))
	signInReq.Header.Set("Content-Type", "application/json")
	signInRec := httptest.NewRecorder()
	engine.ServeHTTP(signInRec, signInReq)

	var signInResp signInResponse
	if err := json.Unmarshal(signInRec.Body.Bytes(), &signInResp); err != nil {
		t.Fatalf("failed to unmarshal sign-in response: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/auth/sign-out", nil)
	req.Header.Set("Authorization", "Bearer "+signInResp.Token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestSignOut_NoHeader(t *testing.T) {
	svc := auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
	engine := newAuthRoutesEngine(svc)

	req := httptest.NewRequest(http.MethodPost, "/auth/sign-out", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}
