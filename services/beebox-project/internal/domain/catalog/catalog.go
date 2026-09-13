package catalog

import (
	"errors"

	beeboxauth "github.com/DoMinhHHung/beebox-dev/modules/beebox-auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

var (
	ErrInvalidModuleDefinition     = errors.New("invalid module definition")
	ErrInvalidCapabilityDefinition = errors.New("invalid capability definition")
	ErrInvalidDataFieldDefinition  = errors.New("invalid data field definition")
	ErrDuplicateModule             = errors.New("duplicate module")
	ErrDuplicateCapability         = errors.New("duplicate capability")
	ErrDuplicateDataField          = errors.New("duplicate data field")
	ErrUnknownModule               = errors.New("unknown module")
	ErrUnknownCapability           = errors.New("unknown capability")
	ErrUnknownDataField            = errors.New("unknown data field")
	ErrIncompatibleDataFieldChange = errors.New("incompatible data field change")
)

type ChangeKind = beeboxauth.ChangeKind

const (
	ChangeAdditive     = beeboxauth.ChangeAdditive
	ChangeIncompatible = beeboxauth.ChangeIncompatible
)

type ModuleDefinition = beeboxauth.ModuleDefinition

type CapabilityDefinition = beeboxauth.CapabilityDefinition

type DataFieldDefinition = beeboxauth.DataFieldDefinition

type moduleKey struct {
	ID      string
	Version string
}

type capabilityKey struct {
	ModuleID string
	ID       string
	Version  string
}

type dataFieldKey struct {
	ModuleID          string
	CapabilityID      string
	CapabilityVersion string
	ID                string
	Version           string
}

type Catalog struct {
	modules      map[moduleKey]ModuleDefinition
	capabilities map[capabilityKey]CapabilityDefinition
	dataFields   map[dataFieldKey]DataFieldDefinition
}

func New(
	modules []ModuleDefinition,
	capabilities []CapabilityDefinition,
	dataFields []DataFieldDefinition,
) (Catalog, error) {
	catalog := Catalog{
		modules:      make(map[moduleKey]ModuleDefinition, len(modules)),
		capabilities: make(map[capabilityKey]CapabilityDefinition, len(capabilities)),
		dataFields:   make(map[dataFieldKey]DataFieldDefinition, len(dataFields)),
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

		if !hasModuleID(catalog.modules, definition.ModuleID) {
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

	for _, definition := range dataFields {
		if definition.ModuleID == "" ||
			definition.CapabilityID == "" ||
			definition.CapabilityVersion == "" ||
			definition.ID == "" ||
			definition.Version == "" {
			return Catalog{}, ErrInvalidDataFieldDefinition
		}

		if definition.ChangeKind != ChangeAdditive &&
			definition.ChangeKind != ChangeIncompatible {
			return Catalog{}, ErrInvalidDataFieldDefinition
		}

		capabilityKeyValue := capabilityKey{
			ModuleID: definition.ModuleID,
			ID:       definition.CapabilityID,
			Version:  definition.CapabilityVersion,
		}

		if _, exists := catalog.capabilities[capabilityKeyValue]; !exists {
			return Catalog{}, ErrUnknownCapability
		}

		key := dataFieldKey{
			ModuleID:          definition.ModuleID,
			CapabilityID:      definition.CapabilityID,
			CapabilityVersion: definition.CapabilityVersion,
			ID:                definition.ID,
			Version:           definition.Version,
		}

		if _, exists := catalog.dataFields[key]; exists {
			return Catalog{}, ErrDuplicateDataField
		}

		catalog.dataFields[key] = definition
	}

	return catalog, nil
}

func hasModuleID(
	modules map[moduleKey]ModuleDefinition,
	id string,
) bool {
	for key := range modules {
		if key.ID == id {
			return true
		}
	}

	return false
}

func (c Catalog) Module(
	id string,
	version string,
) (ModuleDefinition, error) {
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

func (c Catalog) DataField(
	moduleID string,
	capabilityID string,
	capabilityVersion string,
	id string,
	version string,
) (DataFieldDefinition, error) {
	key := dataFieldKey{
		ModuleID:          moduleID,
		CapabilityID:      capabilityID,
		CapabilityVersion: capabilityVersion,
		ID:                id,
		Version:           version,
	}

	definition, exists := c.dataFields[key]
	if !exists {
		return DataFieldDefinition{}, ErrUnknownDataField
	}

	return definition, nil
}

func (c Catalog) ValidateDataFieldTransition(
	current []configuration.DataFieldReference,
	next []configuration.DataFieldReference,
) error {
	currentSet := make(map[configuration.DataFieldReference]struct{}, len(current))

	for _, field := range current {
		currentSet[field] = struct{}{}
	}

	nextSet := make(map[configuration.DataFieldReference]struct{}, len(next))

	for _, field := range next {
		if _, exists := nextSet[field]; exists {
			return configuration.ErrDuplicateDataField
		}

		if _, err := c.DataField(
			field.ModuleID,
			field.CapabilityID,
			field.CapabilityVersion,
			field.ID,
			field.Version,
		); err != nil {
			return err
		}

		nextSet[field] = struct{}{}

		if _, exists := currentSet[field]; exists {
			continue
		}

		definition, err := c.DataField(
			field.ModuleID,
			field.CapabilityID,
			field.CapabilityVersion,
			field.ID,
			field.Version,
		)
		if err != nil {
			return err
		}

		if definition.ChangeKind == ChangeIncompatible {
			return ErrIncompatibleDataFieldChange
		}
	}

	for field := range currentSet {
		if _, exists := nextSet[field]; !exists {
			return ErrIncompatibleDataFieldChange
		}
	}

	return nil
}

func (c Catalog) Modules() []ModuleDefinition {
	out := make([]ModuleDefinition, 0, len(c.modules))
	for _, definition := range c.modules {
		out = append(out, definition)
	}
	return out
}

func (c Catalog) Capabilities(moduleID string) []CapabilityDefinition {
	out := make([]CapabilityDefinition, 0)
	for _, definition := range c.capabilities {
		if definition.ModuleID == moduleID {
			out = append(out, definition)
		}
	}
	return out
}

func (c Catalog) DataFields(moduleID string, capabilityID string, capabilityVersion string) []DataFieldDefinition {
	out := make([]DataFieldDefinition, 0)
	for _, definition := range c.dataFields {
		if definition.ModuleID == moduleID && definition.CapabilityID == capabilityID && definition.CapabilityVersion == capabilityVersion {
			out = append(out, definition)
		}
	}
	return out
}
