package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/apperror"
)

const maxJSONBodyBytes = 1 << 20

func decodeJSONBody(w http.ResponseWriter, r *http.Request, dest any) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxJSONBodyBytes)
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(dest); err != nil {
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) {
			return apperror.New(apperror.CodeValidation, "invalid request body")
		}
		return apperror.New(apperror.CodeValidation, "invalid request body")
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return apperror.New(apperror.CodeValidation, "invalid request body")
	}
	return nil
}
