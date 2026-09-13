package http

import (
	"net/http"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
)

const projectCredentialHeader = "X-BeeBox-Project-Credential"

func requireProject(resolve *projectresolve.Service, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		projectID := r.PathValue("project_id")
		raw := strings.TrimSpace(r.Header.Get(projectCredentialHeader))
		if raw == "" {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
			return
		}

		project, err := resolve.Resolve(r.Context(), projectID, raw)
		if err != nil {
			writeError(w, err)
			return
		}

		cfg, err := resolve.LoadAppliedConfiguration(r.Context(), project.ProjectID)
		if err != nil {
			writeError(w, err)
			return
		}

		ctx := withProjectContext(r.Context(), project)
		ctx = withAppliedConfiguration(ctx, cfg)
		next(w, r.WithContext(ctx))
	}
}
