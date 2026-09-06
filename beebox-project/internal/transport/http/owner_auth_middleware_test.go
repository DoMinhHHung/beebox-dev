package httpapi

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	ownermemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/memory"
	sessionmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/memory"
	"github.com/gin-gonic/gin"
)

func newAuthTestService() *auth.Service {
	return auth.NewService(ownermemory.NewOwnerRepository(), sessionmemory.NewSessionRepository(), time.Hour)
}

func newProtectedEngine(svc *auth.Service, handlerCallCount *int) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())
	engine.GET("/protected", RequireOwnerSession(svc), func(c *gin.Context) {
		*handlerCallCount++
		ownerID, _ := OwnerIDFromContext(c)
		c.JSON(http.StatusOK, gin.H{"owner_id": ownerID})
	})
	return engine
}

func TestRequireOwnerSession_ValidToken_SetsOwnerID(t *testing.T) {
	svc := newAuthTestService()
	o, err := svc.SignUp(newTestCtx(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error signing up: %v", err)
	}
	token, err := svc.SignIn(newTestCtx(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error signing in: %v", err)
	}

	callCount := 0
	engine := newProtectedEngine(svc, &callCount)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d, body: %s", http.StatusOK, rec.Code, rec.Body.String())
	}
	if callCount != 1 {
		t.Fatalf("expected handler to be called once, got %d", callCount)
	}
	expectedBody := `{"owner_id":"` + o.ID + `"}`
	if rec.Body.String() != expectedBody {
		t.Fatalf("expected body %q, got %q", expectedBody, rec.Body.String())
	}
}

func TestRequireOwnerSession_MissingHeader(t *testing.T) {
	svc := newAuthTestService()
	callCount := 0
	engine := newProtectedEngine(svc, &callCount)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if callCount != 0 {
		t.Fatal("expected handler NOT to be called when Authorization header is missing")
	}
}

func TestRequireOwnerSession_MalformedHeader(t *testing.T) {
	svc := newAuthTestService()
	callCount := 0
	engine := newProtectedEngine(svc, &callCount)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token abc123")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if callCount != 0 {
		t.Fatal("expected handler NOT to be called when Authorization header is malformed")
	}
}

func TestRequireOwnerSession_InvalidToken(t *testing.T) {
	svc := newAuthTestService()
	callCount := 0
	engine := newProtectedEngine(svc, &callCount)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer not-a-real-token")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if callCount != 0 {
		t.Fatal("expected handler NOT to be called when token is invalid")
	}
}

func TestRequireOwnerSession_RevokedToken(t *testing.T) {
	svc := newAuthTestService()
	if _, err := svc.SignUp(newTestCtx(), "minh@example.com", "password123"); err != nil {
		t.Fatalf("unexpected error signing up: %v", err)
	}
	token, err := svc.SignIn(newTestCtx(), "minh@example.com", "password123")
	if err != nil {
		t.Fatalf("unexpected error signing in: %v", err)
	}
	if err := svc.SignOut(newTestCtx(), token); err != nil {
		t.Fatalf("unexpected error signing out: %v", err)
	}

	callCount := 0
	engine := newProtectedEngine(svc, &callCount)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if callCount != 0 {
		t.Fatal("expected handler NOT to be called when token is revoked")
	}
}
