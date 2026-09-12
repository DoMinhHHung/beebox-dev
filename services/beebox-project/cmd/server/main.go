package main

import (
	"context"
	"log"
	"net/http"
	"os"

	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	applicationenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/enablement"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/catalogseed"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/config"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/postgres"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/interfaces/http"
)

func main() {
	cfg, err := config.Load(os.LookupEnv)
	if err != nil {
		log.Fatalf("beebox-project: invalid configuration: %v", err)
	}

	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("beebox-project: failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewProjectRepository(pool)
	service := project.NewService(repo)
	beeCatalog, err := catalogseed.Default()
	if err != nil {
		log.Fatalf("beebox-project: invalid catalog seed: %v", err)
	}
	configurationRepo := postgres.NewConfigurationRepository(pool)
	configurationService := applicationconfiguration.NewService(configurationRepo, service, beeCatalog)
	enablementRepo := postgres.NewEnablementRepository(pool)
	enablementService := applicationenablement.NewService(enablementRepo, service, beeCatalog)

	identityClient := identity.NewClient(cfg.IdentityURL, nil)
	router := interfaceshttp.NewRouterWithServices(service, configurationService, enablementService, identityClient)

	addr := ":" + cfg.Port
	log.Printf("beebox-project: listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
