package catalog_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
)

func TestNewAndLookup_ReturnsRegisteredModule(t *testing.T) {
	got, err := catalog.New(
		[]catalog.ModuleDefinition{
			{
				ID:      "beebox-auth",
				Version: "v1",
			},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	definition, err := got.Module("beebox-auth", "v1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if definition.ID != "beebox-auth" {
		t.Fatalf("expected module ID %q, got %q", "beebox-auth", definition.ID)
	}

	if definition.Version != "v1" {
		t.Fatalf("expected module version %q, got %q", "v1", definition.Version)
	}
}

func TestNew_RejectsInvalidModuleDefinition(t *testing.T) {
	tests := []struct {
		name       string
		definition catalog.ModuleDefinition
	}{
		{
			name: "missing ID",
			definition: catalog.ModuleDefinition{
				Version: "v1",
			},
		},
		{
			name: "missing version",
			definition: catalog.ModuleDefinition{
				ID: "beebox-auth",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := catalog.New(
				[]catalog.ModuleDefinition{test.definition},
				nil,
			)

			if !errors.Is(err, catalog.ErrInvalidModuleDefinition) {
				t.Fatalf(
					"expected ErrInvalidModuleDefinition, got %v",
					err,
				)
			}
		})
	}
}

func TestNew_RejectsDuplicateModule(t *testing.T) {
	definition := catalog.ModuleDefinition{
		ID:      "beebox-auth",
		Version: "v1",
	}

	_, err := catalog.New(
		[]catalog.ModuleDefinition{definition, definition},
		nil,
	)

	if !errors.Is(err, catalog.ErrDuplicateModule) {
		t.Fatalf("expected ErrDuplicateModule, got %v", err)
	}
}

func TestModule_RejectsUnknownIdentifierOrVersion(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		version string
	}{
		{
			name:    "unknown identifier",
			id:      "beebox-payment",
			version: "v1",
		},
		{
			name:    "unknown version",
			id:      "beebox-auth",
			version: "v2",
		},
	}

	catalogValue, err := catalog.New(
		[]catalog.ModuleDefinition{
			{
				ID:      "beebox-auth",
				Version: "v1",
			},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			_, err := catalogValue.Module(test.id, test.version)
			if !errors.Is(err, catalog.ErrUnknownModule) {
				t.Fatalf("expected ErrUnknownModule, got %v", err)
			}
		})
	}
}

func TestNewAndLookup_ReturnsRegisteredCapability(t *testing.T) {
	got, err := catalog.New(
		[]catalog.ModuleDefinition{
			{
				ID:      "beebox-auth",
				Version: "v1",
			},
		},
		[]catalog.CapabilityDefinition{
			{
				ModuleID: "beebox-auth",
				ID:       "login",
				Version:  "v1",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	definition, err := got.Capability(
		"beebox-auth",
		"login",
		"v1",
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if definition.ModuleID != "beebox-auth" {
		t.Fatalf(
			"expected module ID %q, got %q",
			"beebox-auth",
			definition.ModuleID,
		)
	}

	if definition.ID != "login" {
		t.Fatalf("expected capability ID %q, got %q", "login", definition.ID)
	}

	if definition.Version != "v1" {
		t.Fatalf(
			"expected capability version %q, got %q",
			"v1",
			definition.Version,
		)
	}
}

func TestNew_RejectsInvalidCapabilityDefinition(t *testing.T) {
	tests := []struct {
		name       string
		definition catalog.CapabilityDefinition
	}{
		{
			name: "missing module ID",
			definition: catalog.CapabilityDefinition{
				ID:      "login",
				Version: "v1",
			},
		},
		{
			name: "missing capability ID",
			definition: catalog.CapabilityDefinition{
				ModuleID: "beebox-auth",
				Version:  "v1",
			},
		},
		{
			name: "missing capability version",
			definition: catalog.CapabilityDefinition{
				ModuleID: "beebox-auth",
				ID:       "login",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := catalog.New(
				[]catalog.ModuleDefinition{
					{
						ID:      "beebox-auth",
						Version: "v1",
					},
				},
				[]catalog.CapabilityDefinition{test.definition},
			)

			if !errors.Is(err, catalog.ErrInvalidCapabilityDefinition) {
				t.Fatalf(
					"expected ErrInvalidCapabilityDefinition, got %v",
					err,
				)
			}
		})
	}
}

func TestNew_RejectsCapabilityForUnknownModule(t *testing.T) {
	_, err := catalog.New(
		[]catalog.ModuleDefinition{
			{
				ID:      "beebox-auth",
				Version: "v1",
			},
		},
		[]catalog.CapabilityDefinition{
			{
				ModuleID: "beebox-payment",
				ID:       "charge",
				Version:  "v1",
			},
		},
	)

	if !errors.Is(err, catalog.ErrUnknownModule) {
		t.Fatalf("expected ErrUnknownModule, got %v", err)
	}
}

func TestNew_RejectsDuplicateCapability(t *testing.T) {
	definition := catalog.CapabilityDefinition{
		ModuleID: "beebox-auth",
		ID:       "login",
		Version:  "v1",
	}

	_, err := catalog.New(
		[]catalog.ModuleDefinition{
			{
				ID:      "beebox-auth",
				Version: "v1",
			},
		},
		[]catalog.CapabilityDefinition{definition, definition},
	)

	if !errors.Is(err, catalog.ErrDuplicateCapability) {
		t.Fatalf("expected ErrDuplicateCapability, got %v", err)
	}
}

func TestCapability_RejectsUnknownModuleCapabilityOrVersion(t *testing.T) {
	catalogValue, err := catalog.New(
		[]catalog.ModuleDefinition{
			{
				ID:      "beebox-auth",
				Version: "v1",
			},
		},
		[]catalog.CapabilityDefinition{
			{
				ModuleID: "beebox-auth",
				ID:       "login",
				Version:  "v1",
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name     string
		moduleID string
		id       string
		version  string
	}{
		{
			name:     "unknown module",
			moduleID: "beebox-payment",
			id:       "login",
			version:  "v1",
		},
		{
			name:     "unknown capability",
			moduleID: "beebox-auth",
			id:       "logout",
			version:  "v1",
		},
		{
			name:     "unknown version",
			moduleID: "beebox-auth",
			id:       "login",
			version:  "v2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := catalogValue.Capability(
				test.moduleID,
				test.id,
				test.version,
			)

			if !errors.Is(err, catalog.ErrUnknownCapability) {
				t.Fatalf(
					"expected ErrUnknownCapability, got %v",
					err,
				)
			}
		})
	}
}
