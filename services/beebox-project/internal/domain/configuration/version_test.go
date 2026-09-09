package configuration_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

func TestNewVersion_CreatesDraftVersion(t *testing.T) {
	config, err := configuration.New("project-1", "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := configuration.NewVersion("project-1", 1, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Number != 1 {
		t.Fatalf("expected version number %d, got %d", 1, got.Number)
	}

	if got.Status != configuration.StatusDraft {
		t.Fatalf("expected status %q, got %q", configuration.StatusDraft, got.Status)
	}
}

func TestNewVersion_RejectsInvalidInput(t *testing.T) {
	config, err := configuration.New("project-1", "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	otherProjectConfig, err := configuration.New("project-2", "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name      string
		projectID string
		number    int
		config    configuration.Configuration
	}{
		{
			name:      "missing project ID",
			projectID: "",
			number:    1,
			config:    config,
		},
		{
			name:      "zero version number",
			projectID: "project-1",
			number:    0,
			config:    config,
		},
		{
			name:      "negative version number",
			projectID: "project-1",
			number:    -1,
			config:    config,
		},
		{
			name:      "configuration belongs to a different project",
			projectID: "project-1",
			number:    1,
			config:    otherProjectConfig,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := configuration.NewVersion(test.projectID, test.number, test.config)
			if !errors.Is(err, configuration.ErrInvalidVersion) {
				t.Fatalf("expected ErrInvalidVersion, got %v", err)
			}
		})
	}
}

func TestVersionTransition_AllowsValidTransitions(t *testing.T) {
	config, err := configuration.New("project-1", "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name string
		from configuration.LifecycleStatus
		to   configuration.LifecycleStatus
	}{
		{
			name: "draft to validated",
			from: configuration.StatusDraft,
			to:   configuration.StatusValidated,
		},
		{
			name: "validated to published",
			from: configuration.StatusValidated,
			to:   configuration.StatusPublished,
		},
		{
			name: "published to applied",
			from: configuration.StatusPublished,
			to:   configuration.StatusApplied,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version, err := configuration.NewVersion("project-1", 1, config)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			version.Status = test.from

			got, err := version.Transition(test.to)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got.Status != test.to {
				t.Fatalf("expected status %q, got %q", test.to, got.Status)
			}
		})
	}
}

func TestVersionTransition_RejectsInvalidTransitions(t *testing.T) {
	config, err := configuration.New("project-1", "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name string
		from configuration.LifecycleStatus
		to   configuration.LifecycleStatus
	}{
		{
			name: "draft to published",
			from: configuration.StatusDraft,
			to:   configuration.StatusPublished,
		},
		{
			name: "draft to applied",
			from: configuration.StatusDraft,
			to:   configuration.StatusApplied,
		},
		{
			name: "validated to applied",
			from: configuration.StatusValidated,
			to:   configuration.StatusApplied,
		},
		{
			name: "validated back to draft",
			from: configuration.StatusValidated,
			to:   configuration.StatusDraft,
		},
		{
			name: "applied to draft",
			from: configuration.StatusApplied,
			to:   configuration.StatusDraft,
		},
		{
			name: "applied stays applied",
			from: configuration.StatusApplied,
			to:   configuration.StatusApplied,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version, err := configuration.NewVersion("project-1", 1, config)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}
			version.Status = test.from

			_, err = version.Transition(test.to)
			if !errors.Is(err, configuration.ErrInvalidLifecycleTransition) {
				t.Fatalf("expected ErrInvalidLifecycleTransition, got %v", err)
			}
		})
	}
}