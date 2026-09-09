package configuration_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

func TestNew_CreatesConfigurationReference(t *testing.T) {
	got, err := configuration.New("project-1", "auth", "login")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ProjectID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ProjectID)
	}

	if got.ModuleID != "auth" {
		t.Fatalf("expected module ID %q, got %q", "auth", got.ModuleID)
	}

	if got.CapabilityID != "login" {
		t.Fatalf("expected capability ID %q, got %q", "login", got.CapabilityID)
	}
}

func TestNew_RejectsMissingReference(t *testing.T) {
	tests := []struct {
		name         string
		projectID    string
		moduleID     string
		capabilityID string
	}{
		{
			name:         "missing project ID",
			projectID:    "",
			moduleID:     "auth",
			capabilityID: "login",
		},
		{
			name:         "missing module ID",
			projectID:    "project-1",
			moduleID:     "",
			capabilityID: "login",
		},
		{
			name:         "missing capability ID",
			projectID:    "project-1",
			moduleID:     "auth",
			capabilityID: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := configuration.New(
				test.projectID,
				test.moduleID,
				test.capabilityID,
			)

			if !errors.Is(err, configuration.ErrInvalidConfiguration) {
				t.Fatalf("expected ErrInvalidConfiguration, got %v", err)
			}
		})
	}
}
