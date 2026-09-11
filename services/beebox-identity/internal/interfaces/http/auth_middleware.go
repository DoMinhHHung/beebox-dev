package http

import (
	"net/http"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

func requireAuthentication(authenticator *auth.AuthenticateSessionService, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, err := extractBearerToken(r.Header.Get("Authorization"))
		if err != nil {
			writeError(w, err)
			return
		}

		result, err := authenticator.AuthenticateSession(r.Context(), auth.AuthenticateSessionInput{Token: token})
		if err != nil {
			writeError(w, err)
			return
		}

		ctx := withAuthenticationContext(r.Context(), AuthenticationContext{
			UserID:    result.UserID,
			SessionID: result.SessionID,
		})
		next(w, r.WithContext(ctx))
	}
}

func extractBearerToken(header string) (string, error) {
	header = strings.TrimSpace(header)
	if header == "" {
		return "", apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return "", apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	token := strings.TrimSpace(parts[1])
	if token == "" {
		return "", apperror.New(apperror.CodeUnauthenticated, "unauthenticated")
	}
	return token, nil
}
