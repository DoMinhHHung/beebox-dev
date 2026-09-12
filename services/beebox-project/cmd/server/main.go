package main

import (
	"context"
	"log"
	"net/http"
	"os"

	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
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
	configurationRepo := postgres.NewConfigurationRepository(pool)
	configurationService := applicationconfiguration.NewService(configurationRepo, service)
	identityClient := identity.NewClient(cfg.IdentityURL, nil)
	router := interfaceshttp.NewRouterWithConfiguration(service, configurationService, identityClient)

	addr := ":" + cfg.Port
	log.Printf("beebox-project: listening on %s", addr)
	if err := http.ListenAndServe(addr, router); err != nil {
		log.Fatal(err)
	}
}
