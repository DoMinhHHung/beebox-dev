package http

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationcredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/credential"
)

type credentialHandler struct {
	service *applicationcredential.Service
}

type issuePublicCredentialRequest struct {
	Name string `json:"name"`
}

type issuePublicCredentialResponse struct {
	ProjectID  string `json:"project_id"`
	ID         string `json:"id"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	Status     string `json:"status"`
	Credential string `json:"credential"`
}

type verifyPublicCredentialRequest struct {
	Credential string `json:"credential"`
}

type verifyPublicCredentialResponse struct {
	ProjectID    string `json:"project_id"`
	CredentialID string `json:"credential_id"`
	Kind         string `json:"kind"`
}

func (h *credentialHandler) issuePublic(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	var req issuePublicCredentialRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	issued, err := h.service.IssuePublic(r.Context(), r.PathValue("id"), principal.OrganizationID, req.Name)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, issuePublicCredentialResponse{
		ProjectID:  issued.Credential.ProjectID,
		ID:         issued.Credential.ID,
		Name:       issued.Credential.Name,
		Kind:       string(issued.Credential.Kind),
		Status:     string(issued.Credential.Status),
		Credential: issued.Secret,
	})
}

func (h *credentialHandler) verifyPublic(w http.ResponseWriter, r *http.Request) {
	var req verifyPublicCredentialRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	verified, err := h.service.VerifyPublic(r.Context(), r.PathValue("id"), req.Credential)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, verifyPublicCredentialResponse{
		ProjectID:    verified.ProjectID,
		CredentialID: verified.CredentialID,
		Kind:         string(verified.Kind),
	})
}
