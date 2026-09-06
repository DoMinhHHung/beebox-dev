package apperror

import (
	"errors"
	"fmt"
)

type Code string

const (
	CodeInvalidInput              Code = "INVALID_INPUT"
	CodeUnauthenticated           Code = "UNAUTHENTICATED"
	CodeForbidden                 Code = "FORBIDDEN"
	CodeTenantAccessDenied        Code = "TENANT_ACCESS_DENIED"
	CodeNotFound                  Code = "NOT_FOUND"
	CodeConflict                  Code = "CONFLICT"
	CodeCredentialInvalid         Code = "CREDENTIAL_INVALID"
	CodeCredentialRevoked         Code = "CREDENTIAL_REVOKED"
	CodeCredentialTypeUnsupported Code = "CREDENTIAL_TYPE_UNSUPPORTED"
	CodeInternal                  Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	Err     error
}

// New tạo lỗi ứng dụng với mã lỗi và thông báo được cung cấp.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Wrap tạo lỗi ứng dụng với mã, thông báo và lỗi nguyên nhân được cung cấp.
func Wrap(code Code, message string, err error) *Error {
	return &Error{Code: code, Message: message, Err: err}
}

func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Err
}

// CodeOf extracts the application error code from an error chain, returning CodeInternal when no application error is found.
func CodeOf(err error) Code {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}
