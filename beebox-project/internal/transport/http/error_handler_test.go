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
		c.Error(routeErr)
	})
	return engine
}

func TestErrorHandler_MapsCodesToStatus(t *testing.T) {
	cases := []struct {
		name           string
		code           apperror.Code
		expectedStatus int
	}{
		{"invalid input", apperror.CodeInvalidInput, http.StatusBadRequest},
		{"unauthenticated", apperror.CodeUnauthenticated, http.StatusUnauthorized},
		{"forbidden", apperror.CodeForbidden, http.StatusForbidden},
		{"tenant access denied", apperror.CodeTenantAccessDenied, http.StatusForbidden},
		{"not found", apperror.CodeNotFound, http.StatusNotFound},
		{"conflict", apperror.CodeConflict, http.StatusConflict},
		{"credential invalid", apperror.CodeCredentialInvalid, http.StatusUnauthorized},
		{"credential revoked", apperror.CodeCredentialRevoked, http.StatusUnauthorized},
		{"credential type unsupported", apperror.CodeCredentialTypeUnsupported, http.StatusBadRequest},
		{"internal", apperror.CodeInternal, http.StatusInternalServerError},
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
			if body.Error.Message != "test message" {
				t.Fatalf("expected message %q, got %q", "test message", body.Error.Message)
			}
		})
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