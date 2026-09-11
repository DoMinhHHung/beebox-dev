package domain

import "errors"

var (
	ErrInvalidIdentityIdentifier = errors.New("invalid identity identifier")
	ErrInvalidUser               = errors.New("invalid user")
	ErrInvalidCredential         = errors.New("invalid credential")
	ErrInvalidSession            = errors.New("invalid session")
	ErrSessionRevoked            = errors.New("session revoked")
	ErrInvalidVerification       = errors.New("invalid verification")
	ErrVerificationUsed          = errors.New("verification used")
	ErrVerificationExpired       = errors.New("verification expired")
	ErrInvalidPasswordReset      = errors.New("invalid password reset")
	ErrPasswordResetUsed         = errors.New("password reset used")
	ErrPasswordResetExpired      = errors.New("password reset expired")
)
