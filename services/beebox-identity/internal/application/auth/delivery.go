package auth

import "context"

type VerificationDeliveryMessage struct {
	UserID string
	Type   string
	Target string
	Code   string
}

type VerificationMailer interface {
	SendVerificationEmail(ctx context.Context, message VerificationDeliveryMessage) error
}

type VerificationSMSSender interface {
	SendVerificationSMS(ctx context.Context, message VerificationDeliveryMessage) error
}

type PasswordResetDeliveryMessage struct {
	UserID    string
	Target    string
	ResetID   string
	Token     string
	ExpiresAt string
}

type PasswordResetMailer interface {
	SendPasswordReset(ctx context.Context, message PasswordResetDeliveryMessage) error
}
