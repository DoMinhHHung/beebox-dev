package http

import (
	"context"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
)

func NewRouter(projects *project.Service, authenticators ...auth.Authenticator) *http.ServeMux {
	return newRouter(projects, nil, authenticators...)
}

func NewRouterWithConfiguration(projects *project.Service, configurations *applicationconfiguration.Service, authenticator auth.Authenticator) *http.ServeMux {
	return newRouter(projects, configurations, authenticator)
}

func newRouter(projects *project.Service, configurations *applicationconfiguration.Service, authenticators ...auth.Authenticator) *http.ServeMux {
	var authenticator auth.Authenticator = unauthenticatedAuthenticator{}
	if len(authenticators) > 0 && authenticators[0] != nil {
		authenticator = authenticators[0]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)

	projectHandler := &projectHandler{service: projects}
	mux.HandleFunc("POST /v1/projects", requireAuthentication(authenticator, projectHandler.create))
	mux.HandleFunc("GET /v1/projects/{id}", requireAuthentication(authenticator, projectHandler.get))
	mux.HandleFunc("PATCH /v1/projects/{id}", requireAuthentication(authenticator, projectHandler.transition))
	mux.HandleFunc("DELETE /v1/projects/{id}", requireAuthentication(authenticator, projectHandler.archive))

	if configurations != nil {
		configurationHandler := &configurationHandler{service: configurations}
		mux.HandleFunc("PUT /v1/projects/{id}/configuration", requireAuthentication(authenticator, configurationHandler.createOrUpdate))
		mux.HandleFunc("GET /v1/projects/{id}/configuration", requireAuthentication(authenticator, configurationHandler.current))
		mux.HandleFunc("GET /v1/projects/{id}/configuration/versions/{version}", requireAuthentication(authenticator, configurationHandler.version))
		mux.HandleFunc("PATCH /v1/projects/{id}/configuration/versions/{version}", requireAuthentication(authenticator, configurationHandler.transition))
		mux.HandleFunc("POST /v1/projects/{id}/configuration/versions/{version}/rollout", requireAuthentication(authenticator, configurationHandler.rollout))
	}
	return mux
}

type unauthenticatedAuthenticator struct{}

func (unauthenticatedAuthenticator) Authenticate(context.Context, string) (auth.Principal, error) {
	return auth.Principal{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
