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

const (
	ModuleID      = "beebox-auth"
	ModuleVersion = "v1"

	CapabilityPassword        = "password"
	CapabilityPasswordVersion = "v1"

	CapabilitySession        = "session"
	CapabilitySessionVersion = "v1"

	FieldEmail           = "email"
	FieldEmailVersion    = "v1"
	FieldPasswordHash    = "password_hash"
	FieldPasswordHashVer = "v1"
)

func DefaultDefinitions() (
	modules []ModuleDefinition,
	capabilities []CapabilityDefinition,
	dataFields []DataFieldDefinition,
) {
	modules = []ModuleDefinition{
		{ID: ModuleID, Version: ModuleVersion},
	}
	capabilities = []CapabilityDefinition{
		{ModuleID: ModuleID, ID: CapabilityPassword, Version: CapabilityPasswordVersion},
		{ModuleID: ModuleID, ID: CapabilitySession, Version: CapabilitySessionVersion},
	}
	dataFields = []DataFieldDefinition{
		{
			ModuleID: ModuleID, CapabilityID: CapabilityPassword, CapabilityVersion: CapabilityPasswordVersion,
			ID: FieldEmail, Version: FieldEmailVersion, ChangeKind: ChangeAdditive,
		},
		{
			ModuleID: ModuleID, CapabilityID: CapabilityPassword, CapabilityVersion: CapabilityPasswordVersion,
			ID: FieldPasswordHash, Version: FieldPasswordHashVer, ChangeKind: ChangeIncompatible,
		},
	}
	return modules, capabilities, dataFields
}
