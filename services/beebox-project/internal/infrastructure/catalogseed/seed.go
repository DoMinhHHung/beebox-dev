package catalogseed

import "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"

func Default() (catalog.Catalog, error) {
	return catalog.New(
		[]catalog.ModuleDefinition{
			{ID: "beebox-auth", Version: "v1"},
		},
		[]catalog.CapabilityDefinition{
			{ModuleID: "beebox-auth", ID: "password", Version: "v1"},
			{ModuleID: "beebox-auth", ID: "session", Version: "v1"},
		},
		[]catalog.DataFieldDefinition{
			{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "email", Version: "v1", ChangeKind: catalog.ChangeAdditive},
			{ModuleID: "beebox-auth", CapabilityID: "password", CapabilityVersion: "v1", ID: "password_hash", Version: "v1", ChangeKind: catalog.ChangeIncompatible},
		},
	)
}
