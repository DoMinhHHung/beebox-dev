package infrastructure_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/infrastructure"
)

func TestNewDesiredState_CreatesDesiredState(t *testing.T) {
	got, err := infrastructure.NewDesiredState(
		"project-1",
		"supabase",
		"ap-southeast-1",
		"standard",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ProjectID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ProjectID)
	}

	if got.Provider != "supabase" {
		t.Fatalf("expected provider %q, got %q", "supabase", got.Provider)
	}

	if got.Region != "ap-southeast-1" {
		t.Fatalf("expected region %q, got %q", "ap-southeast-1", got.Region)
	}

	if got.Plan != "standard" {
		t.Fatalf("expected plan %q, got %q", "standard", got.Plan)
	}
}

func TestNewDesiredState_RejectsMissingDesiredStateField(t *testing.T) {
	tests := []struct {
		name     string
		project  string
		provider string
		region   string
		plan     string
	}{
		{
			name:     "missing project ID",
			project:  "",
			provider: "supabase",
			region:   "ap-southeast-1",
			plan:     "standard",
		},
		{
			name:     "missing provider",
			project:  "project-1",
			provider: "",
			region:   "ap-southeast-1",
			plan:     "standard",
		},
		{
			name:     "missing region",
			project:  "project-1",
			provider: "supabase",
			region:   "",
			plan:     "standard",
		},
		{
			name:     "missing plan",
			project:  "project-1",
			provider: "supabase",
			region:   "ap-southeast-1",
			plan:     "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := infrastructure.NewDesiredState(
				test.project,
				test.provider,
				test.region,
				test.plan,
			)

			if !errors.Is(err, infrastructure.ErrInvalidDesiredState) {
				t.Fatalf("expected ErrInvalidDesiredState, got %v", err)
			}
		})
	}
}
