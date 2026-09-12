package http

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
)

func NewRouter(resolve *projectresolve.Service, sessions *authcap.Service) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", handleHealthz)

	auth := &authHandler{sessions: sessions}
	mux.HandleFunc("GET /v1/p/{project_id}/auth/session", requireProject(resolve, auth.currentSession))
	mux.HandleFunc("GET /v1/p/{project_id}/_ready", requireProject(resolve, handleProjectReady))
	return mux
}

func handleHealthz(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func handleProjectReady(w http.ResponseWriter, r *http.Request) {
	project, ok := projectContextFrom(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternal, "internal error"))
		return
	}
	cfg, ok := appliedConfigurationFrom(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeInternal, "internal error"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"project_id":      project.ProjectID,
		"credential_id":   project.CredentialID,
		"applied_version": cfg.AppliedVersion,
		"module_id":       cfg.ModuleID,
		"capability_id":   cfg.CapabilityID,
	})
}
