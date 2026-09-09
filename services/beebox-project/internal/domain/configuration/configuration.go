package configuration

import "errors"

var ErrInvalidConfiguration = errors.New("invalid configuration")

type Configuration struct {
	ProjectID    string
	ModuleID     string
	CapabilityID string
}

func New(projectID string, moduleID string, capabilityID string) (Configuration, error) {
	if projectID == "" || moduleID == "" || capabilityID == "" {
		return Configuration{}, ErrInvalidConfiguration
	}

	return Configuration{
		ProjectID:    projectID,
		ModuleID:     moduleID,
		CapabilityID: capabilityID,
	}, nil
}
