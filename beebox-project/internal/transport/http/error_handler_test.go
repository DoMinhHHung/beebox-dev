package httpapi

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/gin-gonic/gin"
)

func newTestEngine(routeErr error) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(ErrorHandler())
	engine.GET("/trigger", func(c *gin.Context) {
		_ = c.Error(routeErr)
	})
	return engine
}

func TestErrorHandler_MapsCodesToStatus(t *testing.T) {
	cases := []struct {
		name            string
		code            apperror.Code
		expectedStatus  int
		expectedMessage string
	}{
		{"invalid input", apperror.CodeInvalidInput, http.StatusBadRequest, "test message"},
		{"unauthenticated", apperror.CodeUnauthenticated, http.StatusUnauthorized, "test message"},
		{"forbidden", apperror.CodeForbidden, http.StatusForbidden, "test message"},
		{"tenant access denied", apperror.CodeTenantAccessDenied, http.StatusForbidden, "test message"},
		{"not found", apperror.CodeNotFound, http.StatusNotFound, "test message"},
		{"conflict", apperror.CodeConflict, http.StatusConflict, "test message"},
		{"credential invalid", apperror.CodeCredentialInvalid, http.StatusUnauthorized, "test message"},
		{"credential revoked", apperror.CodeCredentialRevoked, http.StatusUnauthorized, "test message"},
		{"credential type unsupported", apperror.CodeCredentialTypeUnsupported, http.StatusBadRequest, "test message"},
		{"internal", apperror.CodeInternal, http.StatusInternalServerError, "internal error"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			appErr := apperror.New(tc.code, "test message")
			engine := newTestEngine(appErr)

			req := httptest.NewRequest(http.MethodGet, "/trigger", nil)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			if rec.Code != tc.expectedStatus {
				t.Fatalf("expected status %d, got %d", tc.expectedStatus, rec.Code)
			}

			var body errorBody
			if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
				t.Fatalf("failed to unmarshal response body: %v", err)
			}

			if body.Error.Code != tc.code {
				t.Fatalf("expected code %s, got %s", tc.code, body.Error.Code)
			}
			if body.Error.Message != tc.expectedMessage {
				t.Fatalf("expected message %q, got %q", tc.expectedMessage, body.Error.Message)
			}
		})
	}
}

func TestErrorHandler_InternalAppErrorDoesNotLeakMessage(t *testing.T) {
	appErr := apperror.New(apperror.CodeInternal, "pq: secret connection string at 10.0.0.5:5432")
	engine := newTestEngine(appErr)

	req := httptest.NewRequest(http.MethodGet, "/trigger", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if body.Error.Code != apperror.CodeInternal {
		t.Fatalf("expected code %s, got %s", apperror.CodeInternal, body.Error.Code)
	}
	if body.Error.Message != "internal error" {
		t.Fatalf("expected generic message, got %q", body.Error.Message)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.5") || strings.Contains(rec.Body.String(), "secret") {
		t.Fatal("response leaked internal error detail")
	}
}

func TestErrorHandler_UnknownErrorDoesNotLeakDetails(t *testing.T) {
	rawErr := errors.New("pq: connection refused at 10.0.0.5:5432")
	engine := newTestEngine(rawErr)

	req := httptest.NewRequest(http.MethodGet, "/trigger", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status %d, got %d", http.StatusInternalServerError, rec.Code)
	}

	var body errorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response body: %v", err)
	}

	if body.Error.Code != apperror.CodeInternal {
		t.Fatalf("expected code %s, got %s", apperror.CodeInternal, body.Error.Code)
	}
	if body.Error.Message != "internal error" {
		t.Fatalf("expected generic message, got %q", body.Error.Message)
	}
	if strings.Contains(rec.Body.String(), "10.0.0.5") {
		t.Fatal("response leaked internal error detail")
	}
}
