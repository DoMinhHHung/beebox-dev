package delivery

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

type TwilioSMSConfig struct {
	AccountSID string
	AuthToken  string
	From       string
	APIBaseURL string
	HTTPClient *http.Client
}

type TwilioSMSSender struct {
	cfg TwilioSMSConfig
}

func NewTwilioSMSSender(cfg TwilioSMSConfig) *TwilioSMSSender {
	if cfg.APIBaseURL == "" {
		cfg.APIBaseURL = "https://api.twilio.com"
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: 10 * time.Second}
	}
	return &TwilioSMSSender{cfg: cfg}
}

func (s *TwilioSMSSender) SendVerificationSMS(ctx context.Context, message auth.VerificationDeliveryMessage) error {
	if message.Target == "" || message.Code == "" {
		return fmt.Errorf("invalid sms payload")
	}
	if s.cfg.AccountSID == "" || s.cfg.AuthToken == "" || s.cfg.From == "" {
		return fmt.Errorf("sms configuration incomplete")
	}

	endpoint := fmt.Sprintf("%s/2010-04-01/Accounts/%s/Messages.json",
		strings.TrimRight(s.cfg.APIBaseURL, "/"),
		url.PathEscape(s.cfg.AccountSID),
	)

	form := url.Values{}
	form.Set("To", message.Target)
	form.Set("From", s.cfg.From)
	form.Set("Body", "Your BeeBox verification code is "+message.Code)

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}
	req.SetBasicAuth(s.cfg.AccountSID, s.cfg.AuthToken)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.cfg.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("sms provider returned status %d", resp.StatusCode)
	}
	return nil
}

var _ auth.VerificationSMSSender = (*TwilioSMSSender)(nil)
