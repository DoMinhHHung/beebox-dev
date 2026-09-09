package catalog_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
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
		t.Fatalf(
			"expected module version %q, got %q",
			"v1",
			definition.Version,
		)
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
		nil,
	)
	if !errors.Is(err, catalog.ErrDuplicateModule) {
		t.Fatalf("expected ErrDuplicateModule, got %v", err)
	}
}

func TestModule_RejectsUnknownIdentifierOrVersion(t *testing.T) {
	tests := []struct {
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
		nil,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	for _, test := range tests {
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
		nil,
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
		t.Fatalf(
			"expected capability ID %q, got %q",
			"login",
			definition.ID,
		)
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
				nil,
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
		nil,
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
		nil,
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
		nil,
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

func TestNewAndLookup_ReturnsRegisteredDataField(t *testing.T) {
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	definition, err := got.DataField(
		"beebox-auth",
		"login",
		"v1",
		"email",
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

	if definition.CapabilityID != "login" {
		t.Fatalf(
			"expected capability ID %q, got %q",
			"login",
			definition.CapabilityID,
		)
	}

	if definition.CapabilityVersion != "v1" {
		t.Fatalf(
			"expected capability version %q, got %q",
			"v1",
			definition.CapabilityVersion,
		)
	}

	if definition.ID != "email" {
		t.Fatalf("expected field ID %q, got %q", "email", definition.ID)
	}

	if definition.Version != "v1" {
		t.Fatalf(
			"expected field version %q, got %q",
			"v1",
			definition.Version,
		)
	}

	if definition.ChangeKind != catalog.ChangeAdditive {
		t.Fatalf(
			"expected change kind %q, got %q",
			catalog.ChangeAdditive,
			definition.ChangeKind,
		)
	}
}

func TestNew_RejectsInvalidDataFieldDefinition(t *testing.T) {
	tests := []struct {
		name       string
		definition catalog.DataFieldDefinition
	}{
		{
			name: "missing module ID",
			definition: catalog.DataFieldDefinition{
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
		{
			name: "missing capability ID",
			definition: catalog.DataFieldDefinition{
				ModuleID:          "beebox-auth",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
		{
			name: "missing capability version",
			definition: catalog.DataFieldDefinition{
				ModuleID:     "beebox-auth",
				CapabilityID: "login",
				ID:           "email",
				Version:      "v1",
				ChangeKind:   catalog.ChangeAdditive,
			},
		},
		{
			name: "missing field ID",
			definition: catalog.DataFieldDefinition{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
		{
			name: "missing field version",
			definition: catalog.DataFieldDefinition{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
		{
			name: "invalid change kind",
			definition: catalog.DataFieldDefinition{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        "UNKNOWN",
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
				[]catalog.CapabilityDefinition{
					{
						ModuleID: "beebox-auth",
						ID:       "login",
						Version:  "v1",
					},
				},
				[]catalog.DataFieldDefinition{test.definition},
			)

			if !errors.Is(err, catalog.ErrInvalidDataFieldDefinition) {
				t.Fatalf(
					"expected ErrInvalidDataFieldDefinition, got %v",
					err,
				)
			}
		})
	}
}

func TestNew_RejectsDataFieldForUnknownCapability(t *testing.T) {
	_, err := catalog.New(
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "logout",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
	)
	if !errors.Is(err, catalog.ErrUnknownCapability) {
		t.Fatalf("expected ErrUnknownCapability, got %v", err)
	}
}

func TestNew_RejectsDuplicateDataField(t *testing.T) {
	definition := catalog.DataFieldDefinition{
		ModuleID:          "beebox-auth",
		CapabilityID:      "login",
		CapabilityVersion: "v1",
		ID:                "email",
		Version:           "v1",
		ChangeKind:        catalog.ChangeAdditive,
	}

	_, err := catalog.New(
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
		[]catalog.DataFieldDefinition{definition, definition},
	)
	if !errors.Is(err, catalog.ErrDuplicateDataField) {
		t.Fatalf("expected ErrDuplicateDataField, got %v", err)
	}
}

func TestDataField_RejectsUnknownCustomFieldOrVersion(t *testing.T) {
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name    string
		fieldID string
		version string
	}{
		{
			name:    "custom field",
			fieldID: "custom_field",
			version: "v1",
		},
		{
			name:    "version mismatch",
			fieldID: "email",
			version: "v2",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := catalogValue.DataField(
				"beebox-auth",
				"login",
				"v1",
				test.fieldID,
				test.version,
			)
			if !errors.Is(err, catalog.ErrUnknownDataField) {
				t.Fatalf("expected ErrUnknownDataField, got %v", err)
			}
		})
	}
}

func TestValidateDataFieldTransition_AllowsAdditiveField(t *testing.T) {
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "display_name",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	current := []configuration.DataFieldReference{
		{
			ModuleID:          "beebox-auth",
			CapabilityID:      "login",
			CapabilityVersion: "v1",
			ID:                "email",
			Version:           "v1",
		},
	}

	next := append(
		[]configuration.DataFieldReference(nil),
		current...,
	)
	next = append(next, configuration.DataFieldReference{
		ModuleID:          "beebox-auth",
		CapabilityID:      "login",
		CapabilityVersion: "v1",
		ID:                "display_name",
		Version:           "v1",
	})

	if err := catalogValue.ValidateDataFieldTransition(current, next); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestValidateDataFieldTransition_RejectsIncompatibleField(t *testing.T) {
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "legacy_id",
				Version:           "v1",
				ChangeKind:        catalog.ChangeIncompatible,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	current := []configuration.DataFieldReference{
		{
			ModuleID:          "beebox-auth",
			CapabilityID:      "login",
			CapabilityVersion: "v1",
			ID:                "email",
			Version:           "v1",
		},
	}

	next := append(
		[]configuration.DataFieldReference(nil),
		current...,
	)
	next = append(next, configuration.DataFieldReference{
		ModuleID:          "beebox-auth",
		CapabilityID:      "login",
		CapabilityVersion: "v1",
		ID:                "legacy_id",
		Version:           "v1",
	})

	err = catalogValue.ValidateDataFieldTransition(current, next)
	if !errors.Is(err, catalog.ErrIncompatibleDataFieldChange) {
		t.Fatalf(
			"expected ErrIncompatibleDataFieldChange, got %v",
			err,
		)
	}
}

func TestValidateDataFieldTransition_RejectsDuplicateField(t *testing.T) {
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	field := configuration.DataFieldReference{
		ModuleID:          "beebox-auth",
		CapabilityID:      "login",
		CapabilityVersion: "v1",
		ID:                "email",
		Version:           "v1",
	}

	err = catalogValue.ValidateDataFieldTransition(
		nil,
		[]configuration.DataFieldReference{field, field},
	)
	if !errors.Is(err, configuration.ErrDuplicateDataField) {
		t.Fatalf("expected ErrDuplicateDataField, got %v", err)
	}
}

func TestValidateDataFieldTransition_RejectsRemovedField(t *testing.T) {
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
		[]catalog.DataFieldDefinition{
			{
				ModuleID:          "beebox-auth",
				CapabilityID:      "login",
				CapabilityVersion: "v1",
				ID:                "email",
				Version:           "v1",
				ChangeKind:        catalog.ChangeAdditive,
			},
		},
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	field := configuration.DataFieldReference{
		ModuleID:          "beebox-auth",
		CapabilityID:      "login",
		CapabilityVersion: "v1",
		ID:                "email",
		Version:           "v1",
	}

	err = catalogValue.ValidateDataFieldTransition(
		[]configuration.DataFieldReference{field},
		nil,
	)
	if !errors.Is(err, catalog.ErrIncompatibleDataFieldChange) {
		t.Fatalf(
			"expected ErrIncompatibleDataFieldChange, got %v",
			err,
		)
	}
}
