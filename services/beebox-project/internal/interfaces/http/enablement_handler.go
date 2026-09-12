package http

import (
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/enablement"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
	domainenablement "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/enablement"
)

type enablementHandler struct {
	service *applicationenablement.Service
}

type enablementFieldRequest struct {
	ModuleID          string `json:"module_id"`
	CapabilityID      string `json:"capability_id"`
	CapabilityVersion string `json:"capability_version"`
	ID                string `json:"id"`
	Version           string `json:"version"`
}

type enableRequest struct {
	ModuleVersion     string                   `json:"module_version"`
	CapabilityVersion string                   `json:"capability_version"`
	DataFields        []enablementFieldRequest `json:"data_fields"`
}

type patchFieldsRequest struct {
	DataFields []enablementFieldRequest `json:"data_fields"`
}

type enablementResponse struct {
	ProjectID         string                   `json:"project_id"`
	ModuleID          string                   `json:"module_id"`
	ModuleVersion     string                   `json:"module_version"`
	CapabilityID      string                   `json:"capability_id"`
	CapabilityVersion string                   `json:"capability_version"`
	DataFields        []enablementFieldRequest `json:"data_fields"`
}

func (h *enablementHandler) listCatalogModules(w http.ResponseWriter, _ *http.Request) {
	modules := h.service.ListCatalogModules()
	out := make([]map[string]string, 0, len(modules))
	for _, module := range modules {
		out = append(out, map[string]string{"id": module.ID, "version": module.Version})
	}
	writeJSON(w, http.StatusOK, map[string]any{"modules": out})
}

func (h *enablementHandler) getCatalogModule(w http.ResponseWriter, r *http.Request) {
	moduleID := r.PathValue("module")
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "v1"
	}
	definition, err := h.service.GetCatalogModule(moduleID, version)
	if err != nil {
		writeError(w, err)
		return
	}
	capabilities := h.service.ListCatalogCapabilities(moduleID)
	caps := make([]map[string]string, 0, len(capabilities))
	for _, capability := range capabilities {
		caps = append(caps, map[string]string{"id": capability.ID, "version": capability.Version})
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"id":           definition.ID,
		"version":      definition.Version,
		"capabilities": caps,
	})
}

func (h *enablementHandler) getCatalogCapability(w http.ResponseWriter, r *http.Request) {
	moduleID := r.PathValue("module")
	capabilityID := r.PathValue("capability")
	version := r.URL.Query().Get("version")
	if version == "" {
		version = "v1"
	}
	definition, err := h.service.GetCatalogCapability(moduleID, capabilityID, version)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{
		"module_id": definition.ModuleID,
		"id":        definition.ID,
		"version":   definition.Version,
	})
}

func (h *enablementHandler) list(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	items, err := h.service.List(r.Context(), r.PathValue("id"), principal.OrganizationID)
	if err != nil {
		writeError(w, err)
		return
	}
	out := make([]enablementResponse, 0, len(items))
	for _, item := range items {
		out = append(out, toEnablementResponse(item))
	}
	writeJSON(w, http.StatusOK, map[string]any{"capabilities": out})
}

func (h *enablementHandler) enable(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	var req enableRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	fields := toFieldReferences(req.DataFields)
	item, err := h.service.Enable(
		r.Context(),
		r.PathValue("id"),
		principal.OrganizationID,
		r.PathValue("module"),
		req.ModuleVersion,
		r.PathValue("capability"),
		req.CapabilityVersion,
		fields,
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toEnablementResponse(item))
}

func (h *enablementHandler) updateFields(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	var req patchFieldsRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	item, err := h.service.UpdateFields(
		r.Context(),
		r.PathValue("id"),
		principal.OrganizationID,
		r.PathValue("module"),
		r.PathValue("capability"),
		toFieldReferences(req.DataFields),
	)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toEnablementResponse(item))
}

func (h *enablementHandler) disable(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	if err := h.service.Disable(r.Context(), r.PathValue("id"), principal.OrganizationID, r.PathValue("module"), r.PathValue("capability")); err != nil {
		writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toFieldReferences(fields []enablementFieldRequest) []configuration.DataFieldReference {
	out := make([]configuration.DataFieldReference, len(fields))
	for i, field := range fields {
		out[i] = configuration.DataFieldReference{
			ModuleID:          field.ModuleID,
			CapabilityID:      field.CapabilityID,
			CapabilityVersion: field.CapabilityVersion,
			ID:                field.ID,
			Version:           field.Version,
		}
	}
	return out
}

func toEnablementResponse(item domainenablement.Enablement) enablementResponse {
	fields := make([]enablementFieldRequest, len(item.DataFields))
	for i, field := range item.DataFields {
		fields[i] = enablementFieldRequest{
			ModuleID:          field.ModuleID,
			CapabilityID:      field.CapabilityID,
			CapabilityVersion: field.CapabilityVersion,
			ID:                field.ID,
			Version:           field.Version,
		}
	}
	return enablementResponse{
		ProjectID:         item.ProjectID,
		ModuleID:          item.ModuleID,
		ModuleVersion:     item.ModuleVersion,
		CapabilityID:      item.CapabilityID,
		CapabilityVersion: item.CapabilityVersion,
		DataFields:        fields,
	}
}
