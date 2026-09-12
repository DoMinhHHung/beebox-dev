package http

import (
	"context"
	"net/http"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/auth"
)

type authenticationContextKey int

const principalContextKey authenticationContextKey = 1

func principalFromContext(ctx context.Context) (auth.Principal, bool) {
	principal, ok := ctx.Value(principalContextKey).(auth.Principal)
	return principal, ok
}

func withPrincipal(ctx context.Context, principal auth.Principal) context.Context {
	return context.WithValue(ctx, principalContextKey, principal)
}

func requireAuthentication(authenticator auth.Authenticator, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || strings.TrimSpace(parts[1]) == "" {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
			return
		}

		principal, err := authenticator.Authenticate(r.Context(), strings.TrimSpace(parts[1]))
		if err != nil {
			writeError(w, err)
			return
		}
		if principal.UserID == "" || principal.OrganizationID == "" {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
			return
		}
		next(w, r.WithContext(withPrincipal(r.Context(), principal)))
	}
}
