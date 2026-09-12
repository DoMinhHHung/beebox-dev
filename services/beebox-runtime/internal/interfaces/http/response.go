package http

import (
	"encoding/json"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
)

type errorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type errorEnvelope struct {
	Error errorBody `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err error) {
	code := apperror.CodeOf(err)
	message := "internal error"
	if code != apperror.CodeInternal {
		var appErr *apperror.Error
		if e, ok := err.(*apperror.Error); ok {
			message = e.Message
			_ = appErr
		}
	}
	writeJSON(w, statusForCode(code), errorEnvelope{Error: errorBody{Code: string(code), Message: message}})
}

func statusForCode(code apperror.Code) int {
	switch code {
	case apperror.CodeValidation:
		return http.StatusBadRequest
	case apperror.CodeUnauthenticated:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeConflict:
		return http.StatusConflict
	case apperror.CodeDependencyFailure:
		return http.StatusBadGateway
	default:
		return http.StatusInternalServerError
	}
}
