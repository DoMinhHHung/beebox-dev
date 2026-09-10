package http

import (
	"net/http"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
)

func TestStatusForCode_MapsEveryKnownCode(t *testing.T) {
	tests := []struct {
		code   apperror.Code
		status int
	}{
		{apperror.CodeValidation, http.StatusBadRequest},
		{apperror.CodeUnauthenticated, http.StatusUnauthorized},
		{apperror.CodeForbidden, http.StatusForbidden},
		{apperror.CodeNotFound, http.StatusNotFound},
		{apperror.CodeConflict, http.StatusConflict},
		{apperror.CodeDependencyFailure, http.StatusBadGateway},
		{apperror.CodeInternal, http.StatusInternalServerError},
	}

	for _, test := range tests {
		t.Run(string(test.code), func(t *testing.T) {
			if got := statusForCode(test.code); got != test.status {
				t.Fatalf("expected status %d, got %d", test.status, got)
			}
		})
	}
}

func TestStatusForCode_DefaultsToInternalForUnknownCode(t *testing.T) {
	if got := statusForCode(apperror.Code("SOMETHING_NEW")); got != http.StatusInternalServerError {
		t.Fatalf("expected 500 for unknown code, got %d", got)
	}
}
