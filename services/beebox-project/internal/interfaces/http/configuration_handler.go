package http

import (
	"net/http"
	"strconv"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/configuration"
	domainconfiguration "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/configuration"
)

type configurationHandler struct {
	service *applicationconfiguration.Service
}

type configurationRequest struct {
	ModuleID          string             `json:"module_id"`
	ModuleVersion     string             `json:"module_version"`
	CapabilityID      string             `json:"capability_id"`
	CapabilityVersion string             `json:"capability_version"`
	DataFields        []dataFieldRequest `json:"data_fields"`
}

type dataFieldRequest struct {
	ModuleID          string `json:"module_id"`
	CapabilityID      string `json:"capability_id"`
	CapabilityVersion string `json:"capability_version"`
	ID                string `json:"id"`
	Version           string `json:"version"`
}

type configurationResponse struct {
	ProjectID         string             `json:"project_id"`
	Version           int                `json:"version"`
	ModuleID          string             `json:"module_id"`
	ModuleVersion     string             `json:"module_version"`
	CapabilityID      string             `json:"capability_id"`
	CapabilityVersion string             `json:"capability_version"`
	DataFields        []dataFieldRequest `json:"data_fields"`
	Status            string             `json:"status"`
}

type rolloutResponse struct {
	ProjectID      string `json:"project_id"`
	DesiredVersion int    `json:"desired_version"`
	AppliedVersion int    `json:"applied_version"`
}

type transitionConfigurationRequest struct {
	Status string `json:"status"`
}

func (h *configurationHandler) createOrUpdate(w http.ResponseWriter, r *http.Request) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	var req configurationRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	config, err := toConfiguration(r.PathValue("id"), req)
	if err != nil {
		writeError(w, err)
		return
	}
	version, err := h.service.CreateOrUpdate(r.Context(), r.PathValue("id"), principal.OrganizationID, config)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toConfigurationResponse(version))
}

func (h *configurationHandler) current(w http.ResponseWriter, r *http.Request) {
	version, err := h.authorizedVersion(r)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfigurationResponse(version))
}

func (h *configurationHandler) version(w http.ResponseWriter, r *http.Request) {
	versionNumber, err := parseConfigurationVersion(r.PathValue("version"))
	if err != nil {
		writeError(w, err)
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	version, err := h.service.Version(r.Context(), r.PathValue("id"), principal.OrganizationID, versionNumber)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfigurationResponse(version))
}

func (h *configurationHandler) transition(w http.ResponseWriter, r *http.Request) {
	versionNumber, err := parseConfigurationVersion(r.PathValue("version"))
	if err != nil {
		writeError(w, err)
		return
	}
	var req transitionConfigurationRequest
	if err := decodeJSONBody(w, r, &req); err != nil {
		writeError(w, err)
		return
	}
	status := domainconfiguration.LifecycleStatus(req.Status)
	if _, ok := knownConfigurationStatuses[status]; !ok {
		writeError(w, apperror.New(apperror.CodeValidation, "unknown configuration status"))
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	version, err := h.service.Transition(r.Context(), r.PathValue("id"), principal.OrganizationID, versionNumber, status)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfigurationResponse(version))
}

func (h *configurationHandler) rollout(w http.ResponseWriter, r *http.Request) {
	versionNumber, err := parseConfigurationVersion(r.PathValue("version"))
	if err != nil {
		writeError(w, err)
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	state, err := h.service.Rollout(r.Context(), r.PathValue("id"), principal.OrganizationID, versionNumber)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, rolloutResponse{ProjectID: state.ProjectID, DesiredVersion: state.DesiredVersion, AppliedVersion: state.AppliedVersion})
}

func (h *configurationHandler) apply(w http.ResponseWriter, r *http.Request) {
	versionNumber, err := parseConfigurationVersion(r.PathValue("version"))
	if err != nil {
		writeError(w, err)
		return
	}
	principal, ok := principalFromContext(r.Context())
	if !ok {
		writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
		return
	}
	version, _, err := h.service.Apply(r.Context(), r.PathValue("id"), principal.OrganizationID, versionNumber)
	if err != nil {
		writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toConfigurationResponse(version))
}

func (h *configurationHandler) authorizedVersion(r *http.Request) (domainconfiguration.Version, error) {
	principal, ok := principalFromContext(r.Context())
	if !ok {
		return domainconfiguration.Version{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	return h.service.Current(r.Context(), r.PathValue("id"), principal.OrganizationID)
}

func toConfiguration(projectID string, req configurationRequest) (domainconfiguration.Configuration, error) {
	config, err := domainconfiguration.New(projectID, req.ModuleID, req.ModuleVersion, req.CapabilityID, req.CapabilityVersion)
	if err != nil {
		return domainconfiguration.Configuration{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	fields := make([]domainconfiguration.DataFieldReference, len(req.DataFields))
	for i, field := range req.DataFields {
		fields[i] = domainconfiguration.DataFieldReference{ModuleID: field.ModuleID, CapabilityID: field.CapabilityID, CapabilityVersion: field.CapabilityVersion, ID: field.ID, Version: field.Version}
	}
	config, err = config.WithDataFields(fields)
	if err != nil {
		return domainconfiguration.Configuration{}, apperror.New(apperror.CodeValidation, err.Error())
	}
	return config, nil
}

func toConfigurationResponse(version domainconfiguration.Version) configurationResponse {
	fields := make([]dataFieldRequest, len(version.Configuration.DataFields))
	for i, field := range version.Configuration.DataFields {
		fields[i] = dataFieldRequest{ModuleID: field.ModuleID, CapabilityID: field.CapabilityID, CapabilityVersion: field.CapabilityVersion, ID: field.ID, Version: field.Version}
	}
	return configurationResponse{ProjectID: version.ProjectID, Version: version.Number, ModuleID: version.Configuration.ModuleID, ModuleVersion: version.Configuration.ModuleVersion, CapabilityID: version.Configuration.CapabilityID, CapabilityVersion: version.Configuration.CapabilityVersion, DataFields: fields, Status: string(version.Status)}
}

var knownConfigurationStatuses = map[domainconfiguration.LifecycleStatus]struct{}{
	domainconfiguration.StatusDraft: {}, domainconfiguration.StatusValidated: {}, domainconfiguration.StatusPublished: {}, domainconfiguration.StatusApplied: {},
}

func parseConfigurationVersion(raw string) (int, error) {
	version, err := strconv.Atoi(raw)
	if err != nil || version < 1 {
		return 0, apperror.New(apperror.CodeValidation, "configuration version must be positive")
	}
	return version, nil
}
