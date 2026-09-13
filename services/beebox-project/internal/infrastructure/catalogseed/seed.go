package catalogseed

import (
	beeboxauth "github.com/DoMinhHHung/beebox-dev/modules/beebox-auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/catalog"
)

func Default() (catalog.Catalog, error) {
	return catalog.New(
		beeboxauth.Modules(),
		beeboxauth.Capabilities(),
		beeboxauth.DataFields(),
	)
}
