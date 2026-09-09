package configuration_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

func newTestVersion(
	t *testing.T,
	projectID string,
	number int,
	status configuration.LifecycleStatus,
) configuration.Version {
	t.Helper()

	config, err := configuration.New(projectID, "auth", "v1", "login", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	version, err := configuration.NewVersion(projectID, number, config)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	version.Status = status
	return version
}

func TestNewRolloutState_RejectsMissingProjectID(t *testing.T) {
	_, err := configuration.NewRolloutState("")
	if !errors.Is(err, configuration.ErrInvalidRolloutState) {
		t.Fatalf("expected ErrInvalidRolloutState, got %v", err)
	}
}

func TestWithDesiredVersion_AcceptsPublishedOrAppliedVersion(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name   string
		status configuration.LifecycleStatus
	}{
		{name: "published version", status: configuration.StatusPublished},
		{name: "applied version", status: configuration.StatusApplied},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := newTestVersion(t, "project-1", 3, test.status)

			got, err := state.WithDesiredVersion(version)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if got.DesiredVersion != 3 {
				t.Fatalf("expected desired version %d, got %d", 3, got.DesiredVersion)
			}
		})
	}
}

func TestWithDesiredVersion_RejectsInactivatableVersion(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name   string
		status configuration.LifecycleStatus
	}{
		{name: "draft version", status: configuration.StatusDraft},
		{name: "validated version", status: configuration.StatusValidated},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := newTestVersion(t, "project-1", 1, test.status)

			_, err := state.WithDesiredVersion(version)
			if !errors.Is(err, configuration.ErrInvalidLifecycleTransition) {
				t.Fatalf("expected ErrInvalidLifecycleTransition, got %v", err)
			}
		})
	}
}

func TestWithDesiredVersion_RejectsVersionFromAnotherProject(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	version := newTestVersion(t, "project-2", 1, configuration.StatusPublished)

	_, err = state.WithDesiredVersion(version)
	if !errors.Is(err, configuration.ErrInvalidRolloutState) {
		t.Fatalf("expected ErrInvalidRolloutState, got %v", err)
	}
}

func TestWithDesiredVersion_AllowsMovingBackForRollback(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	v2 := newTestVersion(t, "project-1", 2, configuration.StatusPublished)
	state, err = state.WithDesiredVersion(v2)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	v1 := newTestVersion(t, "project-1", 1, configuration.StatusApplied)
	got, err := state.WithDesiredVersion(v1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.DesiredVersion != 1 {
		t.Fatalf("expected desired version %d, got %d", 1, got.DesiredVersion)
	}
}

func TestWithAppliedVersion_AcceptsAppliedVersion(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	version := newTestVersion(t, "project-1", 1, configuration.StatusApplied)

	got, err := state.WithAppliedVersion(version)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.AppliedVersion != 1 {
		t.Fatalf("expected applied version %d, got %d", 1, got.AppliedVersion)
	}
}

func TestWithAppliedVersion_RejectsNonAppliedVersion(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name   string
		status configuration.LifecycleStatus
	}{
		{name: "draft version", status: configuration.StatusDraft},
		{name: "validated version", status: configuration.StatusValidated},
		{name: "published version", status: configuration.StatusPublished},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			version := newTestVersion(t, "project-1", 1, test.status)

			_, err := state.WithAppliedVersion(version)
			if !errors.Is(err, configuration.ErrInvalidLifecycleTransition) {
				t.Fatalf("expected ErrInvalidLifecycleTransition, got %v", err)
			}
		})
	}
}

func TestWithAppliedVersion_RejectsVersionFromAnotherProject(t *testing.T) {
	state, err := configuration.NewRolloutState("project-1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	version := newTestVersion(t, "project-2", 1, configuration.StatusApplied)

	_, err = state.WithAppliedVersion(version)
	if !errors.Is(err, configuration.ErrInvalidRolloutState) {
		t.Fatalf("expected ErrInvalidRolloutState, got %v", err)
	}
}