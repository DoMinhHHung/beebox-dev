package domain

type ProjectContext struct {
	ProjectID    string
	CredentialID string
}

type DataField struct {
	ModuleID          string
	CapabilityID      string
	CapabilityVersion string
	ID                string
	Version           string
}

type AppliedConfiguration struct {
	ProjectID         string
	ProjectStatus     string
	AppliedVersion    int
	ModuleID          string
	ModuleVersion     string
	CapabilityID      string
	CapabilityVersion string
	DataFields        []DataField
}
