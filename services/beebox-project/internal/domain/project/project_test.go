package project_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

func TestNew_CreatesDraftProject(t *testing.T) {
	got, err := project.New("project-1", "organization-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ID)
	}

	if got.OrganizationID != "organization-1" {
		t.Fatalf("expected organization ID %q, got %q", "organization-1", got.OrganizationID)
	}

	if got.Status != project.StatusDraft {
		t.Fatalf("expected status %q, got %q", project.StatusDraft, got.Status)
	}
}

func TestNew_RejectsMissingIdentityOrOwnership(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		organizationID string
	}{
		{
			name:           "missing project ID",
			id:             "",
			organizationID: "organization-1",
		},
		{
			name:           "missing organization ID",
			id:             "project-1",
			organizationID: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := project.New(test.id, test.organizationID)
			if !errors.Is(err, project.ErrInvalidProject) {
				t.Fatalf("expected ErrInvalidProject, got %v", err)
			}
		})
	}
}

func TestProjectTransition_AllowsValidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from project.Status
		to   project.Status
	}{
		{
			name: "draft to active",
			from: project.StatusDraft,
			to:   project.StatusActive,
		},
		{
			name: "active to suspended",
			from: project.StatusActive,
			to:   project.StatusSuspended,
		},
		{
			name: "suspended to active",
			from: project.StatusSuspended,
			to:   project.StatusActive,
		},
		{
			name: "active to archived",
			from: project.StatusActive,
			to:   project.StatusArchived,
		},
		{
			name: "suspended to archived",
			from: project.StatusSuspended,
			to:   project.StatusArchived,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := project.Project{Status: test.from}
			got, err := current.Transition(test.to)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got.Status != test.to {
				t.Fatalf("expected status %q, got %q", test.to, got.Status)
			}
		})
	}
}

func TestProjectTransition_RejectsInvalidTransitions(t *testing.T) {
	tests := []struct {
		name string
		from project.Status
		to   project.Status
	}{
		{
			name: "draft to suspended",
			from: project.StatusDraft,
			to:   project.StatusSuspended,
		},
		{
			name: "draft to archived",
			from: project.StatusDraft,
			to:   project.StatusArchived,
		},
		{
			name: "active to draft",
			from: project.StatusActive,
			to:   project.StatusDraft,
		},
		{
			name: "archived to active",
			from: project.StatusArchived,
			to:   project.StatusActive,
		},
		{
			name: "archived to archived",
			from: project.StatusArchived,
			to:   project.StatusArchived,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current := project.Project{Status: test.from}
			_, err := current.Transition(test.to)
			if !errors.Is(err, project.ErrInvalidTransition) {
				t.Fatalf("expected ErrInvalidTransition, got %v", err)
			}
		})
	}
}
