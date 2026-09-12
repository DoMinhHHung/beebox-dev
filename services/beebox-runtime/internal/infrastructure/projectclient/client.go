package projectclient

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/projectresolve"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/domain"
)

type Client struct {
	baseURL    string
	token      string
	httpClient *http.Client
}

func New(baseURL, token string, timeout time.Duration, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		token:      token,
		httpClient: httpClient,
	}
}

var (
	_ projectresolve.CredentialVerifier         = (*Client)(nil)
	_ projectresolve.AppliedConfigurationLoader = (*Client)(nil)
)

type verifyRequest struct {
	Credential string `json:"credential"`
}

type verifyResponse struct {
	ProjectID    string `json:"project_id"`
	CredentialID string `json:"credential_id"`
	Kind         string `json:"kind"`
}

type appliedResponse struct {
	ProjectID         string `json:"project_id"`
	ProjectStatus     string `json:"project_status"`
	AppliedVersion    int    `json:"applied_version"`
	ModuleID          string `json:"module_id"`
	ModuleVersion     string `json:"module_version"`
	CapabilityID      string `json:"capability_id"`
	CapabilityVersion string `json:"capability_version"`
	DataFields        []struct {
		ModuleID          string `json:"module_id"`
		CapabilityID      string `json:"capability_id"`
		CapabilityVersion string `json:"capability_version"`
		ID                string `json:"id"`
		Version           string `json:"version"`
	} `json:"data_fields"`
}

func (c *Client) VerifyPublic(ctx context.Context, projectID, rawCredential string) (domain.ProjectContext, error) {
	body, err := json.Marshal(verifyRequest{Credential: rawCredential})
	if err != nil {
		return domain.ProjectContext{}, apperror.Wrap(apperror.CodeInternal, "failed to encode credential verification request", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/internal/v1/projects/"+projectID+"/credentials/verify", bytes.NewReader(body))
	if err != nil {
		return domain.ProjectContext{}, apperror.Wrap(apperror.CodeInternal, "failed to create credential verification request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/json")

	res, err := c.httpClient.Do(req)
	if err != nil {
		return domain.ProjectContext{}, apperror.Wrap(apperror.CodeDependencyFailure, "project service unavailable", err)
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return domain.ProjectContext{}, apperror.Wrap(apperror.CodeDependencyFailure, "project service unavailable", err)
	}

	switch res.StatusCode {
	case http.StatusOK:
		var out verifyResponse
		if err := json.Unmarshal(payload, &out); err != nil || out.ProjectID == "" || out.CredentialID == "" {
			return domain.ProjectContext{}, apperror.New(apperror.CodeDependencyFailure, "project service returned invalid verification response")
		}
		if out.ProjectID != projectID {
			return domain.ProjectContext{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
		}
		if out.Kind != "" && out.Kind != "PUBLIC" {
			return domain.ProjectContext{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
		}
		return domain.ProjectContext{ProjectID: out.ProjectID, CredentialID: out.CredentialID}, nil
	case http.StatusUnauthorized:
		return domain.ProjectContext{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	case http.StatusForbidden:
		return domain.ProjectContext{}, apperror.New(apperror.CodeForbidden, "forbidden")
	case http.StatusNotFound:
		return domain.ProjectContext{}, apperror.New(apperror.CodeNotFound, "project not found")
	default:
		return domain.ProjectContext{}, apperror.New(apperror.CodeDependencyFailure, "project service unavailable")
	}
}

func (c *Client) GetAppliedConfiguration(ctx context.Context, projectID string) (domain.AppliedConfiguration, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/internal/v1/projects/"+projectID+"/applied-configuration", nil)
	if err != nil {
		return domain.AppliedConfiguration{}, apperror.Wrap(apperror.CodeInternal, "failed to create applied configuration request", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return domain.AppliedConfiguration{}, apperror.Wrap(apperror.CodeDependencyFailure, "project service unavailable", err)
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return domain.AppliedConfiguration{}, apperror.Wrap(apperror.CodeDependencyFailure, "project service unavailable", err)
	}

	switch res.StatusCode {
	case http.StatusOK:
		var out appliedResponse
		if err := json.Unmarshal(payload, &out); err != nil || out.ProjectID == "" || out.AppliedVersion < 1 {
			return domain.AppliedConfiguration{}, apperror.New(apperror.CodeDependencyFailure, "project service returned invalid applied configuration")
		}
		fields := make([]domain.DataField, 0, len(out.DataFields))
		for _, f := range out.DataFields {
			fields = append(fields, domain.DataField{
				ModuleID:          f.ModuleID,
				CapabilityID:      f.CapabilityID,
				CapabilityVersion: f.CapabilityVersion,
				ID:                f.ID,
				Version:           f.Version,
			})
		}
		return domain.AppliedConfiguration{
			ProjectID:         out.ProjectID,
			ProjectStatus:     out.ProjectStatus,
			AppliedVersion:    out.AppliedVersion,
			ModuleID:          out.ModuleID,
			ModuleVersion:     out.ModuleVersion,
			CapabilityID:      out.CapabilityID,
			CapabilityVersion: out.CapabilityVersion,
			DataFields:        fields,
		}, nil
	case http.StatusUnauthorized:
		return domain.AppliedConfiguration{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	case http.StatusForbidden:
		return domain.AppliedConfiguration{}, apperror.New(apperror.CodeForbidden, "forbidden")
	case http.StatusNotFound:
		return domain.AppliedConfiguration{}, apperror.New(apperror.CodeNotFound, "configuration not applied")
	default:
		return domain.AppliedConfiguration{}, apperror.New(apperror.CodeDependencyFailure, "project service unavailable")
	}
}
