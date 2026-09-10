package configuration_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

func TestNew_CreatesConfigurationReference(t *testing.T) {
	got, err := configuration.New(
		"project-1",
		"auth",
		"v1",
		"login",
		"v1",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ProjectID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ProjectID)
	}

	if got.ModuleID != "auth" {
		t.Fatalf("expected module ID %q, got %q", "auth", got.ModuleID)
	}

	if got.ModuleVersion != "v1" {
		t.Fatalf(
			"expected module version %q, got %q",
			"v1",
			got.ModuleVersion,
		)
	}

	if got.CapabilityID != "login" {
		t.Fatalf(
			"expected capability ID %q, got %q",
			"login",
			got.CapabilityID,
		)
	}

	if got.CapabilityVersion != "v1" {
		t.Fatalf(
			"expected capability version %q, got %q",
			"v1",
			got.CapabilityVersion,
		)
	}
}

func TestNew_RejectsMissingReference(t *testing.T) {
	tests := []struct {
		name              string
		projectID         string
		moduleID          string
		moduleVersion     string
		capabilityID      string
		capabilityVersion string
	}{
		{
			name:              "missing project ID",
			projectID:         "",
			moduleID:          "auth",
			moduleVersion:     "v1",
			capabilityID:      "login",
			capabilityVersion: "v1",
		},
		{
			name:              "missing module ID",
			projectID:         "project-1",
			moduleID:          "",
			moduleVersion:     "v1",
			capabilityID:      "login",
			capabilityVersion: "v1",
		},
		{
			name:              "missing module version",
			projectID:         "project-1",
			moduleID:          "auth",
			moduleVersion:     "",
			capabilityID:      "login",
			capabilityVersion: "v1",
		},
		{
			name:              "missing capability ID",
			projectID:         "project-1",
			moduleID:          "auth",
			moduleVersion:     "v1",
			capabilityID:      "",
			capabilityVersion: "v1",
		},
		{
			name:              "missing capability version",
			projectID:         "project-1",
			moduleID:          "auth",
			moduleVersion:     "v1",
			capabilityID:      "login",
			capabilityVersion: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := configuration.New(
				test.projectID,
				test.moduleID,
				test.moduleVersion,
				test.capabilityID,
				test.capabilityVersion,
			)

			if !errors.Is(err, configuration.ErrInvalidConfiguration) {
				t.Fatalf(
					"expected ErrInvalidConfiguration, got %v",
					err,
				)
			}
		})
	}
}

func TestWithDataFields_AddsControlledReferences(t *testing.T) {
	current, err := configuration.New(
		"project-1",
		"auth",
		"v1",
		"login",
		"v1",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	fields := []configuration.DataFieldReference{
		{
			ModuleID:          "auth",
			CapabilityID:      "login",
			CapabilityVersion: "v1",
			ID:                "email",
			Version:           "v1",
		},
	}

	got, err := current.WithDataFields(fields)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(got.DataFields) != 1 {
		t.Fatalf("expected one data field, got %d", len(got.DataFields))
	}

	if got.DataFields[0] != fields[0] {
		t.Fatalf("expected data field reference %#v, got %#v", fields[0], got.DataFields[0])
	}
}

func TestWithDataFields_RejectsMissingReferencePart(t *testing.T) {
	tests := []struct {
		name  string
		field configuration.DataFieldReference
	}{
		{
			name: "missing module ID",
			field: configuration.DataFieldReference{
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
			},
		},
		{
			name: "missing capability ID",
			field: configuration.DataFieldReference{
				ModuleID:          "auth",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
			},
		},
		{
			name: "missing capability version",
			field: configuration.DataFieldReference{
				ModuleID:     "auth",
				CapabilityID: "login",
				ID:           "email",
				Version:      "v1",
			},
		},
		{
			name: "missing field ID",
			field: configuration.DataFieldReference{
				ModuleID:          "auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				Version:           "v1",
			},
		},
		{
			name: "missing field version",
			field: configuration.DataFieldReference{
				ModuleID:          "auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			current, err := configuration.New(
				"project-1",
				"auth",
				"v1",
				"login",
				"v1",
			)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			_, err = current.WithDataFields(
				[]configuration.DataFieldReference{test.field},
			)
			if !errors.Is(err, configuration.ErrInvalidDataFieldReference) {
				t.Fatalf(
					"expected ErrInvalidDataFieldReference, got %v",
					err,
				)
			}
		})
	}
}

func TestWithDataFields_RejectsDuplicateReference(t *testing.T) {
	field := configuration.DataFieldReference{
		ModuleID:          "auth",
		CapabilityID:      "login",
		CapabilityVersion: "v1",
		ID:                "email",
		Version:           "v1",
	}

	current, err := configuration.New(
		"project-1",
		"auth",
		"v1",
		"login",
		"v1",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = current.WithDataFields([]configuration.DataFieldReference{
		field,
		field,
	})
	if !errors.Is(err, configuration.ErrDuplicateDataField) {
		t.Fatalf("expected ErrDuplicateDataField, got %v", err)
	}
}
