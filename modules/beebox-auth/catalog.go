package beeboxauth

type ChangeKind string

const (
	ChangeAdditive     ChangeKind = "ADDITIVE"
	ChangeIncompatible ChangeKind = "INCOMPATIBLE"
)

type ModuleDefinition struct {
	ID      string
	Version string
}

type CapabilityDefinition struct {
	ModuleID string
	ID       string
	Version  string
}

type DataFieldDefinition struct {
	ModuleID          string
	CapabilityID      string
	CapabilityVersion string
	ID                string
	Version           string
	ChangeKind        ChangeKind
}

func Modules() []ModuleDefinition {
	return []ModuleDefinition{
		{ID: "beebox-auth", Version: "v1"},
	}
}

func Capabilities() []CapabilityDefinition {
	return []CapabilityDefinition{
		{ModuleID: "beebox-auth", ID: "password", Version: "v1"},
		{ModuleID: "beebox-auth", ID: "session", Version: "v1"},
	}
}

func DataFields() []DataFieldDefinition {
	return []DataFieldDefinition{
		{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "email", Version: "v1", ChangeKind: ChangeAdditive},
		{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "password_hash", Version: "v1", ChangeKind: ChangeIncompatible},
	}
}