package http

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
)

func NewRouter(projects *project.Service) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /healthz", handleHealthz)

	h := &projectHandler{service: projects}
	mux.HandleFunc("POST /v1/projects", h.create)
	mux.HandleFunc("GET /v1/projects/{id}", h.get)
	mux.HandleFunc("PATCH /v1/projects/{id}", h.transition)
	mux.HandleFunc("DELETE /v1/projects/{id}", h.archive)

	return mux
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
