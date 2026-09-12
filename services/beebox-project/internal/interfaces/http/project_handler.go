package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
)

const maxJSONBodyBytes = 1 << 20

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dest); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return apperror.New(apperror.CodePayloadTooLarge, "request body too large")
		}
		return apperror.New(apperror.CodeValidation, "invalid request body")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return apperror.New(apperror.CodeValidation, "invalid request body")
	}
	return nil
}

type projectHandler struct {
	service *project.Service
}

type projectResponse struct {
	ID             string `json:"id"`
	OrganizationID string `json:"organization_id"`
	Status         string `json:"status"`
}

func toProjectResponse(p domainproject.Project) projectResponse {
	return projectResponse{ID: p.ID, OrganizationID: p.OrganizationID, Status: string(p.Status)}
}

type createProjectRequest struct {
	ID string `json:"id"`
}

func (h *projectHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createProjectRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	p, err := h.service.Create(r.Context(), req.ID, principal.OrganizationID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toProjectResponse(p))
}

func (h *projectHandler) get(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	p, err := h.service.GetAuthorized(r.Context(), r.PathValue("id"), principal.OrganizationID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

type transitionRequest struct {
	Status string `json:"status"`
}

var knownStatuses = map[string]domainproject.Status{
	string(domainproject.StatusDraft):     domainproject.StatusDraft,
	string(domainproject.StatusActive):    domainproject.StatusActive,
	string(domainproject.StatusSuspended): domainproject.StatusSuspended,
	string(domainproject.StatusArchived):  domainproject.StatusArchived,
}

func parseStatus(raw string) (domainproject.Status, error) {
	status, ok := knownStatuses[raw]
	if !ok {
		return "", apperror.New(apperror.CodeValidation, "unknown project status")
	}
	return status, nil
}

func (h *projectHandler) transition(w http.ResponseWriter, r *http.Request) {
	var req transitionRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	status, err := parseStatus(req.Status)
	if err != nil {
		writeError(w, err)
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	p, err := h.service.TransitionAuthorized(r.Context(), r.PathValue("id"), status, principal.OrganizationID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}

func (h *projectHandler) archive(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	p, err := h.service.TransitionAuthorized(r.Context(), r.PathValue("id"), domainproject.StatusArchived, principal.OrganizationID)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toProjectResponse(p))
}
