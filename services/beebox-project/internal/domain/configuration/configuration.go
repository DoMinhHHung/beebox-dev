package configuration

import "errors"

var ErrInvalidConfiguration = errors.New("invalid configuration")

type Configuration struct {
	ProjectID         string
	ModuleID          string
	ModuleVersion     string
	CapabilityID      string
	CapabilityVersion string
}

func New(
	projectID string,
	moduleID string,
	moduleVersion string,
	capabilityID string,
	capabilityVersion string,
) (Configuration, error) {
	if projectID == "" ||
		moduleID == "" ||
		moduleVersion == "" ||
		capabilityID == "" ||
		capabilityVersion == "" {
		return Configuration{}, ErrInvalidConfiguration
	}

	return Configuration{
		ProjectID:         projectID,
		ModuleID:          moduleID,
		ModuleVersion:     moduleVersion,
		CapabilityID:      capabilityID,
		CapabilityVersion: capabilityVersion,
	}, nil
}
