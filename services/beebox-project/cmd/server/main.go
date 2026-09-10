package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/postgres"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/interfaces/http"
)

func main() {
	ctx := context.Background()

	pool, err := postgres.NewPool(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("beebox-project: failed to connect to database: %v", err)
	}
	defer pool.Close()

	repo := postgres.NewProjectRepository(pool)
	service := project.NewService(repo)
	router := interfaceshttp.NewRouter(service)

	log.Println("beebox-project: listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
