package beeboxauth_test

import (
	"testing"

	beeboxauth "github.com/DoMinhHHung/beebox-dev/modules/beebox-auth"
)

func TestModules_ReturnsBeeboxAuthModule(t *testing.T) {
	modules := beeboxauth.Modules()

	if len(modules) != 1 {
		t.Fatalf("expected 1 module, got %d", len(modules))
	}

	if modules[0].ID != "beebox-auth" {
		t.Fatalf("expected module ID %q, got %q", "beebox-auth", modules[0].ID)
	}

	if modules[0].Version != "v1" {
		t.Fatalf("expected module version %q, got %q", "v1", modules[0].Version)
	}
}

func TestCapabilities_ReturnsPasswordAndSession(t *testing.T) {
	tests := []struct {
		name     string
		id       string
		version  string
		moduleID string
	}{
		{name: "password", id: "password", version: "v1", moduleID: "beebox-auth"},
		{name: "session", id: "session", version: "v1", moduleID: "beebox-auth"},
	}

	capabilities := beeboxauth.Capabilities()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var found *beeboxauth.CapabilityDefinition
			for i := range capabilities {
				if capabilities[i].ID == test.id {
					found = &capabilities[i]
					break
				}
			}

			if found == nil {
				t.Fatalf("expected capability %q to exist", test.id)
			}

			if found.ModuleID != test.moduleID {
				t.Fatalf("expected capability %q module ID %q, got %q", test.id, test.moduleID, found.ModuleID)
			}

			if found.Version != test.version {
				t.Fatalf("expected capability %q version %q, got %q", test.id, test.version, found.Version)
			}
		})
	}
}

func TestDataFields_ReturnsEmailAndPasswordHash(t *testing.T) {
	tests := []struct {
		name              string
		id                string
		moduleID          string
		capabilityID      string
		capabilityVersion string
		version           string
		changeKind        beeboxauth.ChangeKind
	}{
		{
			name:              "email",
			id:                "email",
			moduleID:          "beebox-auth",
			capabilityID:      "password",
			capabilityVersion: "v1",
			version:           "v1",
			changeKind:        beeboxauth.ChangeAdditive,
		},
		{
			name:              "password_hash",
			id:                "password_hash",
			moduleID:          "beebox-auth",
			capabilityID:      "password",
			capabilityVersion: "v1",
			version:           "v1",
			changeKind:        beeboxauth.ChangeIncompatible,
		},
	}

	dataFields := beeboxauth.DataFields()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var found *beeboxauth.DataFieldDefinition
			for i := range dataFields {
				if dataFields[i].ID == test.id {
					found = &dataFields[i]
					break
				}
			}

			if found == nil {
				t.Fatalf("expected data field %q to exist", test.id)
			}

			if found.ModuleID != test.moduleID {
				t.Fatalf("expected data field %q module ID %q, got %q", test.id, test.moduleID, found.ModuleID)
			}

			if found.CapabilityID != test.capabilityID {
				t.Fatalf("expected data field %q capability ID %q, got %q", test.id, test.capabilityID, found.CapabilityID)
			}

			if found.CapabilityVersion != test.capabilityVersion {
				t.Fatalf("expected data field %q capability version %q, got %q", test.id, test.capabilityVersion, found.CapabilityVersion)
			}

			if found.Version != test.version {
				t.Fatalf("expected data field %q version %q, got %q", test.id, test.version, found.Version)
			}

			if found.ChangeKind != test.changeKind {
				t.Fatalf("expected data field %q change kind %q, got %q", test.id, test.changeKind, found.ChangeKind)
			}
		})
	}
}
