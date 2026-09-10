package apperror

import "errors"

type Code string

const (
	CodeValidation        Code = "VALIDATION"
	CodeUnauthenticated   Code = "UNAUTHENTICATED"
	CodeForbidden         Code = "FORBIDDEN"
	CodeNotFound          Code = "NOT_FOUND"
	CodeConflict          Code = "CONFLICT"
	CodeDependencyFailure Code = "DEPENDENCY_FAILURE"
	CodeInternal          Code = "INTERNAL"
)

type Error struct {
	Code    Code
	Message string
	Cause   error
}

var _ error = (*Error)(nil)

func (e *Error) Error() string {
	return e.Message
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func New(code Code, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func Wrap(code Code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

func IsCode(err error, code Code) bool {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code == code
	}
	return false
}

func CodeOf(err error) Code {
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr.Code
	}
	return CodeInternal
}