package delivery

import (
	"context"
	"fmt"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

type LoggingSMSSender struct{}

func NewLoggingSMSSender() *LoggingSMSSender {
	return &LoggingSMSSender{}
}

func (s *LoggingSMSSender) SendVerificationSMS(_ context.Context, message auth.VerificationDeliveryMessage) error {
	if message.Target == "" || message.Code == "" {
		return fmt.Errorf("invalid sms payload")
	}
	return nil
}

var _ auth.VerificationSMSSender = (*LoggingSMSSender)(nil)
