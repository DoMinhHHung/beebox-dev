package http

import (
	"encoding/json"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
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
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, err error) {
	code := apperror.CodeOf(err)

	message := "internal error"
	if code != apperror.CodeInternal {
		var appErr *apperror.Error
		if e, ok := err.(*apperror.Error); ok {
			appErr = e
			message = appErr.Message
		}
	}

	writeJSON(w, statusForCode(code), errorEnvelope{
		Error: errorBody{
			Code:    string(code),
			Message: message,
		},
	})
}
