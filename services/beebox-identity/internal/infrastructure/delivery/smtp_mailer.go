package delivery

import (
	"context"
	"fmt"
	"net"
	"net/smtp"
	"strconv"
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

type SMTPConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	From     string
}

type SMTPMailer struct {
	cfg SMTPConfig
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendVerificationEmail(_ context.Context, message auth.VerificationDeliveryMessage) error {
	subject := "BeeBox verification code"
	body := fmt.Sprintf("Your verification code is %s", message.Code)
	return m.send(message.Target, subject, body)
}

func (m *SMTPMailer) SendPasswordReset(_ context.Context, message auth.PasswordResetDeliveryMessage) error {
	subject := "BeeBox password reset"
	body := fmt.Sprintf("Your password reset id is %s and token is %s. It expires at %s.", message.ResetID, message.Token, message.ExpiresAt)
	return m.send(message.Target, subject, body)
}

func (m *SMTPMailer) send(to, subject, body string) error {
	if m.cfg.Host == "" || m.cfg.Port == 0 || m.cfg.From == "" || to == "" {
		return fmt.Errorf("smtp configuration incomplete")
	}
	addr := net.JoinHostPort(m.cfg.Host, strconv.Itoa(m.cfg.Port))
	msg := strings.Join([]string{
		"From: " + m.cfg.From,
		"To: " + to,
		"Subject: " + subject,
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	var auth smtp.Auth
	if m.cfg.Username != "" {
		auth = smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
	}
	return smtp.SendMail(addr, auth, m.cfg.From, []string{to}, []byte(msg))
}

var _ auth.VerificationMailer = (*SMTPMailer)(nil)
var _ auth.PasswordResetMailer = (*SMTPMailer)(nil)
