package http

import (
	"context"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type contextKey int

const authenticationContextKey contextKey = 1

type AuthenticationContext struct {
	UserID    identity.Identifier
	SessionID string
}

func withAuthenticationContext(ctx context.Context, value AuthenticationContext) context.Context {
	return context.WithValue(ctx, authenticationContextKey, value)
}

func AuthenticationFromContext(ctx context.Context) (AuthenticationContext, bool) {
	value, ok := ctx.Value(authenticationContextKey).(AuthenticationContext)
	return value, ok
}
