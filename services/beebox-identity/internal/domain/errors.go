package domain

import "errors"

var (
	ErrInvalidIdentityIdentifier = errors.New("invalid identity identifier")
	ErrInvalidUser               = errors.New("invalid user")
	ErrInvalidCredential         = errors.New("invalid credential")
	ErrInvalidSession            = errors.New("invalid session")
	ErrSessionRevoked            = errors.New("session revoked")
)
