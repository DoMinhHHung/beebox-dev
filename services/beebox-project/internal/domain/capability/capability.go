package capability

import "errors"

var ErrInvalidCapability = errors.New("invalid capability")

type Capability struct {
	ModuleID string
	ID       string
}

func New(moduleID string, id string) (Capability, error) {
	if moduleID == "" || id == "" {
		return Capability{}, ErrInvalidCapability
	}

	return Capability{
		ModuleID: moduleID,
		ID:       id,
	}, nil
}
