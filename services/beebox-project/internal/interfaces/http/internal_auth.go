package http

import (
	"crypto/subtle"
	"net/http"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
)

func requireInternalToken(token string, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := strings.TrimSpace(r.Header.Get("Authorization"))
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
			return
		}
		provided := strings.TrimSpace(parts[1])
		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			writeError(w, apperror.New(apperror.CodeUnauthenticated, "unauthenticated"))
			return
		}
		next(w, r)
	}
}
