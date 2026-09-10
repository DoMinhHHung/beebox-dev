package main

import (
	"log"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
	interfaceshttp "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/interfaces/http"
)

func main() {
	repo := memory.NewProjectRepository()
	service := project.NewService(repo)
	router := interfaceshttp.NewRouter(service)

	log.Println("beebox-project: listening on :8080")
	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatal(err)
	}
}
