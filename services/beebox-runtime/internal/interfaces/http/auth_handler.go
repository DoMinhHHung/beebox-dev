package http

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
)

type authHandler struct {
	sessions *authcap.Service
}

type projectSessionResponse struct {
	ProjectID      string `json:"project_id"`
	UserID         string `json:"user_id"`
	SessionID      string `json:"session_id"`
	OrganizationID string `json:"organization_id,omitempty"`
}

func (h *authHandler) currentSession(w http.ResponseWriter, r *http.Request) {
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

	session, err := h.sessions.CurrentSession(r.Context(), project, cfg, r.Header.Get("Authorization"))
	if err != nil {
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, projectSessionResponse{
		ProjectID:      session.ProjectID,
		UserID:         session.UserID,
		SessionID:      session.SessionID,
		OrganizationID: session.OrganizationID,
	})
}
