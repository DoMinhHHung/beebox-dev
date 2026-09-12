package identityclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/application/authcap"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func New(baseURL string, timeout time.Duration, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: timeout}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

var _ authcap.SessionReader = (*Client)(nil)

type sessionResponse struct {
	UserID         string `json:"user_id"`
	SessionID      string `json:"session_id"`
	OrganizationID string `json:"organization_id"`
}

func (c *Client) GetSession(ctx context.Context, bearerToken string) (authcap.Session, error) {
	if bearerToken == "" {
		return authcap.Session{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/session", nil)
	if err != nil {
		return authcap.Session{}, apperror.Wrap(apperror.CodeInternal, "failed to create identity session request", err)
	}
	req.Header.Set("Authorization", "Bearer "+bearerToken)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return authcap.Session{}, apperror.Wrap(apperror.CodeDependencyFailure, "identity service unavailable", err)
	}
	defer res.Body.Close()

	payload, err := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if err != nil {
		return authcap.Session{}, apperror.Wrap(apperror.CodeDependencyFailure, "identity service unavailable", err)
	}

	switch res.StatusCode {
	case http.StatusOK:
		var out sessionResponse
		if err := json.Unmarshal(payload, &out); err != nil || out.UserID == "" || out.SessionID == "" {
			return authcap.Session{}, apperror.New(apperror.CodeDependencyFailure, "identity service returned invalid session")
		}
		return authcap.Session{
			UserID:         out.UserID,
			SessionID:      out.SessionID,
			OrganizationID: out.OrganizationID,
		}, nil
	case http.StatusUnauthorized:
		return authcap.Session{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	default:
		return authcap.Session{}, apperror.New(apperror.CodeDependencyFailure, "identity service unavailable")
	}
}
