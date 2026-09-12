package http

import (
	"context"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
)

func NewRouter(projects *project.Service, authenticators ...auth.Authenticator) *http.ServeMux {
	var authenticator auth.Authenticator = unauthenticatedAuthenticator{}
	if len(authenticators) > 0 && authenticators[0] != nil {
		authenticator = authenticators[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)

	h := &projectHandler{service: projects}
	mux.HandleFunc("POST /v1/projects", requireAuthentication(authenticator, h.create))
	mux.HandleFunc("GET /v1/projects/{id}", requireAuthentication(authenticator, h.get))
	mux.HandleFunc("PATCH /v1/projects/{id}", requireAuthentication(authenticator, h.transition))
	mux.HandleFunc("DELETE /v1/projects/{id}", requireAuthentication(authenticator, h.archive))
	return mux
}

type unauthenticatedAuthenticator struct{}

func (unauthenticatedAuthenticator) Authenticate(context.Context, string) (auth.Principal, error) {
	return auth.Principal{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
