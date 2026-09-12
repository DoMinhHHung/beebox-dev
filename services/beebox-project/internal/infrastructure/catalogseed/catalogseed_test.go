package catalogseed_test

import (
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/catalogseed"
)

func TestDefault_ExposesBeeboxAuthModule(t *testing.T) {
	got, err := catalogseed.Default()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	definition, err := got.Module("beebox-auth", "v1")
	if err != nil {
		t.Fatalf("expected module lookup to succeed, got %v", err)
	}

	if definition.ID != "beebox-auth" {
		t.Fatalf("expected module ID %q, got %q", "beebox-auth", definition.ID)
	}

	if definition.Version != "v1" {
		t.Fatalf("expected module version %q, got %q", "v1", definition.Version)
	}
}

func TestDefault_ExposesPasswordAndSessionCapabilities(t *testing.T) {
	got, err := catalogseed.Default()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name string
		id   string
	}{
		{name: "password", id: "password"},
		{name: "session", id: "session"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition, err := got.Capability("beebox-auth", test.id, "v1")
			if err != nil {
				t.Fatalf("expected capability lookup to succeed, got %v", err)
			}

			if definition.ModuleID != "beebox-auth" {
				t.Fatalf("expected capability module ID %q, got %q", "beebox-auth", definition.ModuleID)
			}

			if definition.ID != test.id {
				t.Fatalf("expected capability ID %q, got %q", test.id, definition.ID)
			}

			if definition.Version != "v1" {
				t.Fatalf("expected capability version %q, got %q", "v1", definition.Version)
			}
		})
	}
}

func TestDefault_ExposesEmailAndPasswordHashDataFields(t *testing.T) {
	got, err := catalogseed.Default()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	tests := []struct {
		name       string
		id         string
		changeKind catalog.ChangeKind
	}{
		{name: "email", id: "email", changeKind: catalog.ChangeAdditive},
		{name: "password_hash", id: "password_hash", changeKind: catalog.ChangeIncompatible},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			definition, err := got.DataField("beebox-auth", "password", "v1", test.id, "v1")
			if err != nil {
				t.Fatalf("expected data field lookup to succeed, got %v", err)
			}

			if definition.ID != test.id {
				t.Fatalf("expected data field ID %q, got %q", test.id, definition.ID)
			}

			if definition.ChangeKind != test.changeKind {
				t.Fatalf("expected data field %q change kind %q, got %q", test.id, test.changeKind, definition.ChangeKind)
			}
		})
	}
}
