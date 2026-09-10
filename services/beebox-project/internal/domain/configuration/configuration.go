package configuration

import "errors"

var (
	ErrInvalidConfiguration      = errors.New("invalid configuration")
	ErrInvalidDataFieldReference = errors.New("invalid data field reference")
	ErrDuplicateDataField        = errors.New("duplicate data field")
)

type DataFieldReference struct {
	ModuleID          string
	CapabilityID      string
	CapabilityVersion string
	ID                string
	Version           string
}

type Configuration struct {
	ProjectID         string
	ModuleID          string
	ModuleVersion     string
	CapabilityID      string
	CapabilityVersion string
	DataFields        []DataFieldReference
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

func (c Configuration) WithDataFields(
	fields []DataFieldReference,
) (Configuration, error) {
	seen := make(map[DataFieldReference]struct{}, len(fields))

	for _, field := range fields {
		if field.ModuleID == "" ||
			field.CapabilityID == "" ||
			field.CapabilityVersion == "" ||
			field.ID == "" ||
			field.Version == "" {
			return Configuration{}, ErrInvalidDataFieldReference
		}

		if _, exists := seen[field]; exists {
			return Configuration{}, ErrDuplicateDataField
		}

		seen[field] = struct{}{}
	}

	c.DataFields = append([]DataFieldReference(nil), fields...)

	return c, nil
}
