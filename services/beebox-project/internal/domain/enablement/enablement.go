package enablement

import (
	"errors"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

var (
	ErrInvalidEnablement = errors.New("invalid enablement")
	ErrDuplicateField    = errors.New("duplicate data field")
)

type Enablement struct {
	ProjectID         string
	ModuleID          string
	ModuleVersion     string
	CapabilityID      string
	CapabilityVersion string
	DataFields        []configuration.DataFieldReference
}

func New(
	projectID string,
	moduleID string,
	moduleVersion string,
	capabilityID string,
	capabilityVersion string,
) (Enablement, error) {
	if projectID == "" ||
		moduleID == "" ||
		moduleVersion == "" ||
		capabilityID == "" ||
		capabilityVersion == "" {
		return Enablement{}, ErrInvalidEnablement
	}

	return Enablement{
		ProjectID:         projectID,
		ModuleID:          moduleID,
		ModuleVersion:     moduleVersion,
		CapabilityID:      capabilityID,
		CapabilityVersion: capabilityVersion,
	}, nil
}

func (e Enablement) WithDataFields(fields []configuration.DataFieldReference) (Enablement, error) {
	seen := make(map[configuration.DataFieldReference]struct{}, len(fields))

	for _, field := range fields {
		if field.ModuleID == "" ||
			field.CapabilityID == "" ||
			field.CapabilityVersion == "" ||
			field.ID == "" ||
			field.Version == "" {
			return Enablement{}, configuration.ErrInvalidDataFieldReference
		}

		if _, exists := seen[field]; exists {
			return Enablement{}, ErrDuplicateField
		}

		seen[field] = struct{}{}
	}

	e.DataFields = append([]configuration.DataFieldReference(nil), fields...)
	return e, nil
}
