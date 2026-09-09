package catalog

import "errors"

var (
	ErrInvalidModuleDefinition     = errors.New("invalid module definition")
	ErrInvalidCapabilityDefinition = errors.New("invalid capability definition")
	ErrDuplicateModule             = errors.New("duplicate module")
	ErrDuplicateCapability         = errors.New("duplicate capability")
	ErrUnknownModule               = errors.New("unknown module")
	ErrUnknownCapability           = errors.New("unknown capability")
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

type moduleKey struct {
	ID      string
	Version string
}

type capabilityKey struct {
	ModuleID string
	ID       string
	Version  string
}

type Catalog struct {
	modules      map[moduleKey]ModuleDefinition
	capabilities map[capabilityKey]CapabilityDefinition
}

func New(
	modules []ModuleDefinition,
	capabilities []CapabilityDefinition,
) (Catalog, error) {
	catalog := Catalog{
		modules:      make(map[moduleKey]ModuleDefinition, len(modules)),
		capabilities: make(map[capabilityKey]CapabilityDefinition, len(capabilities)),
	}

	for _, definition := range modules {
		if definition.ID == "" || definition.Version == "" {
			return Catalog{}, ErrInvalidModuleDefinition
		}

		key := moduleKey{
			ID:      definition.ID,
			Version: definition.Version,
		}

		if _, exists := catalog.modules[key]; exists {
			return Catalog{}, ErrDuplicateModule
		}

		catalog.modules[key] = definition
	}

	for _, definition := range capabilities {
		if definition.ModuleID == "" ||
			definition.ID == "" ||
			definition.Version == "" {
			return Catalog{}, ErrInvalidCapabilityDefinition
		}

		moduleExists := false
		for key := range catalog.modules {
			if key.ID == definition.ModuleID {
				moduleExists = true
				break
			}
		}

		if !moduleExists {
			return Catalog{}, ErrUnknownModule
		}

		key := capabilityKey{
			ModuleID: definition.ModuleID,
			ID:       definition.ID,
			Version:  definition.Version,
		}

		if _, exists := catalog.capabilities[key]; exists {
			return Catalog{}, ErrDuplicateCapability
		}

		catalog.capabilities[key] = definition
	}

	return catalog, nil
}

func (c Catalog) Module(id string, version string) (ModuleDefinition, error) {
	if id == "" || version == "" {
		return ModuleDefinition{}, ErrUnknownModule
	}

	key := moduleKey{
		ID:      id,
		Version: version,
	}

	definition, exists := c.modules[key]
	if !exists {
		return ModuleDefinition{}, ErrUnknownModule
	}

	return definition, nil
}

func (c Catalog) Capability(
	moduleID string,
	id string,
	version string,
) (CapabilityDefinition, error) {
	if moduleID == "" || id == "" || version == "" {
		return CapabilityDefinition{}, ErrUnknownCapability
	}

	key := capabilityKey{
		ModuleID: moduleID,
		ID:       id,
		Version:  version,
	}

	definition, exists := c.capabilities[key]
	if !exists {
		return CapabilityDefinition{}, ErrUnknownCapability
	}

	return definition, nil
}
