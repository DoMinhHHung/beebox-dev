package httpapi

import (
	"errors"
	"net/http"

	"github.com/DoMinhHHung/beebox-dev/beebox-identity/internal/apperror"
	"github.com/gin-gonic/gin"
)

var statusByCode = map[apperror.Code]int{
	apperror.CodeInvalidInput:              http.StatusBadRequest,
	apperror.CodeUnauthenticated:           http.StatusUnauthorized,
	apperror.CodeForbidden:                 http.StatusForbidden,
	apperror.CodeTenantAccessDenied:        http.StatusForbidden,
	apperror.CodeNotFound:                  http.StatusNotFound,
	apperror.CodeConflict:                  http.StatusConflict,
	apperror.CodeCredentialInvalid:         http.StatusUnauthorized,
	apperror.CodeCredentialRevoked:         http.StatusUnauthorized,
	apperror.CodeCredentialTypeUnsupported: http.StatusBadRequest,
	apperror.CodeInternal:                  http.StatusInternalServerError,
}

type errorBody struct {
	Error errorDetail `json:"error"`
}

type errorDetail struct {
	Code    apperror.Code `json:"code"`
	Message string        `json:"message"`
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last().Err

		code := apperror.CodeInternal
		message := "internal error"

		var appErr *apperror.Error
		if errors.As(err, &appErr) {
			code = appErr.Code
			if code != apperror.CodeInternal {
				message = appErr.Message
			}
		}

		status, ok := statusByCode[code]
		if !ok {
			status = http.StatusInternalServerError
			code = apperror.CodeInternal
		}

		c.JSON(status, errorBody{Error: errorDetail{Code: code, Message: message}})
	}
}
