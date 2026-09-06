package main

import (
	"context"
	"github.com/joho/godotenv"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/auth"
	credentialapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	fielddefinitionapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/fielddefinition"
	projectapp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/config"
	credentialpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/postgres"
	fielddefpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/fielddefinition/postgres"
	ownerpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/postgres"
	sessionpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/postgres"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	projectpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/postgres"
	httpapi "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/transport/http"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuration error: %v", err)
	}

	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL must be set to run beebox-project")
	}

	if err := postgres.RunMigrations(cfg.DatabaseURL); err != nil {
		log.Fatalf("migration error: %v", err)
	}

	ctx, cancelPool := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	cancelPool()
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	ownerRepo := ownerpg.New(pool)
	sessionRepo := sessionpg.New(pool)
	projectRepo := projectpg.New(pool)
	credentialRepo := credentialpg.New(pool)
	fieldRepo := fielddefpg.New(pool)

	deps := httpapi.Dependencies{
		AuthService:            auth.NewService(ownerRepo, sessionRepo, cfg.OwnerSessionTTL),
		ProjectService:         projectapp.NewService(projectRepo),
		CredentialService:      credentialapp.NewService(credentialRepo, projectRepo),
		FieldDefinitionService: fielddefinitionapp.NewService(fieldRepo, projectRepo),
	}

	engine := httpapi.New(deps)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTPPort,
		Handler: engine,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	log.Printf("listening on :%s", cfg.HTTPPort)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("shutdown error: %v", err)
	}

	log.Println("shutdown complete")
}
