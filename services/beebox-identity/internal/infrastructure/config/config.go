package config

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrInvalidPort                   = errors.New("PORT must be a valid port number")
	ErrMissingDatabaseURL            = errors.New("DATABASE_URL is required")
	ErrMissingVerificationCodeSecret = errors.New("VERIFICATION_CODE_SECRET is required")
)

type LookupFunc func(key string) (string, bool)

type Config struct {
	Port                   string
	DatabaseURL            string
	VerificationCodeSecret string
	SMTPHost               string
	SMTPPort               int
	SMTPUser               string
	SMTPPass               string
	SMTPFrom               string
	TwilioAccountSID       string
	TwilioAuthToken        string
	TwilioFromNumber       string
}

func Load(lookup LookupFunc) (Config, error) {
	port, ok := lookup("PORT")
	if !ok || port == "" {
		port = "8081"
	}
	if err := validatePort(port); err != nil {
		return Config{}, err
	}

	databaseURL, ok := lookup("DATABASE_URL")
	if !ok {
		return Config{}, ErrMissingDatabaseURL
	}
	databaseURL = strings.TrimSpace(databaseURL)
	if databaseURL == "" {
		return Config{}, ErrMissingDatabaseURL
	}

	verificationSecret, ok := lookup("VERIFICATION_CODE_SECRET")
	if !ok {
		return Config{}, ErrMissingVerificationCodeSecret
	}
	verificationSecret = strings.TrimSpace(verificationSecret)
	if verificationSecret == "" {
		return Config{}, ErrMissingVerificationCodeSecret
	}

	smtpHost, _ := lookup("SMTP_HOST")
	smtpPortStr, _ := lookup("SMTP_PORT")
	smtpPort := 0
	if smtpPortStr != "" {
		n, err := strconv.Atoi(smtpPortStr)
		if err != nil || n < 1 || n > 65535 {
			return Config{}, errors.New("SMTP_PORT must be a valid port number")
		}
		smtpPort = n
	}
	smtpUser, _ := lookup("SMTP_USERNAME")
	smtpPass, _ := lookup("SMTP_PASSWORD")
	smtpFrom, _ := lookup("SMTP_FROM")
	twilioAccountSID, _ := lookup("TWILIO_ACCOUNT_SID")
	twilioAuthToken, _ := lookup("TWILIO_AUTH_TOKEN")
	twilioFrom, _ := lookup("TWILIO_FROM_NUMBER")

	return Config{
		Port:                   port,
		DatabaseURL:            databaseURL,
		VerificationCodeSecret: verificationSecret,
		SMTPHost:               smtpHost,
		SMTPPort:               smtpPort,
		SMTPUser:               smtpUser,
		SMTPPass:               smtpPass,
		SMTPFrom:               smtpFrom,
		TwilioAccountSID:       twilioAccountSID,
		TwilioAuthToken:        twilioAuthToken,
		TwilioFromNumber:       twilioFrom,
	}, nil
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return ErrInvalidPort
	}
	return nil
}
