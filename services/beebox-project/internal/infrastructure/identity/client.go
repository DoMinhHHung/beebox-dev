package identity

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 5 * time.Second}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient}
}

var _ auth.Authenticator = (*Client)(nil)

type sessionResponse struct {
	UserID    string `json:"user_id"`
	SessionID string `json:"session_id"`
}

func (c *Client) Authenticate(ctx context.Context, token string) (auth.Principal, error) {
	if token == "" {
		return auth.Principal{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/auth/session", nil)
	if err != nil {
		return auth.Principal{}, apperror.Wrap(apperror.CodeDependencyFailure, "identity authentication failed", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.httpClient.Do(req)
	if err != nil {
		return auth.Principal{}, apperror.Wrap(apperror.CodeDependencyFailure, "identity authentication failed", err)
	}
	defer res.Body.Close()

	if res.StatusCode == http.StatusUnauthorized {
		return auth.Principal{}, apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	if res.StatusCode != http.StatusOK {
		return auth.Principal{}, apperror.New(apperror.CodeDependencyFailure, "identity authentication failed")
	}

	var session sessionResponse
	if err := json.NewDecoder(res.Body).Decode(&session); err != nil || session.UserID == "" {
		return auth.Principal{}, apperror.New(apperror.CodeDependencyFailure, "identity authentication failed")
	}

	return auth.Principal{UserID: session.UserID, OrganizationID: session.UserID}, nil
}
