package capability_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/capability"
)

func TestNew_CreatesCapability(t *testing.T) {
	got, err := capability.New("auth", "login")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ModuleID != "auth" {
		t.Fatalf("expected module ID %q, got %q", "auth", got.ModuleID)
	}

	if got.ID != "login" {
		t.Fatalf("expected capability ID %q, got %q", "login", got.ID)
	}
}

func TestNew_RejectsMissingModuleID(t *testing.T) {
	_, err := capability.New("", "login")
	if !errors.Is(err, capability.ErrInvalidCapability) {
		t.Fatalf("expected ErrInvalidCapability, got %v", err)
	}
}

func TestNew_RejectsMissingCapabilityID(t *testing.T) {
	_, err := capability.New("auth", "")
	if !errors.Is(err, capability.ErrInvalidCapability) {
		t.Fatalf("expected ErrInvalidCapability, got %v", err)
	}
}
