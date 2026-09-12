package auth

import "context"

type Principal struct {
	UserID         string
	OrganizationID string
}

type Authenticator interface {
	Authenticate(ctx context.Context, token string) (Principal, error)
}
