package http

import (
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func TestHealthz_ReturnsOK(t *testing.T) {
	router := NewRouter(project.NewService(memory.NewProjectRepository()))

	req := httptest.NewRequest("GET", "/healthz", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != 200 {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
