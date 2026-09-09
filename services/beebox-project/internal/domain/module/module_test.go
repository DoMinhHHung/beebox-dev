package module_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/module"
)

func TestNew_CreatesModule(t *testing.T) {
	got, err := module.New("project-1", "auth")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ProjectID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ProjectID)
	}

	if got.ID != "auth" {
		t.Fatalf("expected module ID %q, got %q", "auth", got.ID)
	}
}

func TestNew_RejectsMissingProjectID(t *testing.T) {
	_, err := module.New("", "auth")
	if !errors.Is(err, module.ErrInvalidModule) {
		t.Fatalf("expected ErrInvalidModule, got %v", err)
	}
}

func TestNew_RejectsMissingModuleID(t *testing.T) {
	_, err := module.New("project-1", "")
	if !errors.Is(err, module.ErrInvalidModule) {
		t.Fatalf("expected ErrInvalidModule, got %v", err)
	}
}
