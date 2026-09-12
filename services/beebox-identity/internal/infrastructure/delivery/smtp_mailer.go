package delivery

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
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
	cfg       SMTPConfig
	tlsConfig *tls.Config
}

func NewSMTPMailer(cfg SMTPConfig) *SMTPMailer {
	return &SMTPMailer{cfg: cfg}
}

func (m *SMTPMailer) SendVerificationEmail(ctx context.Context, message auth.VerificationDeliveryMessage) error {
	subject := "Your verification code"
	body := fmt.Sprintf("Your verification code is %s", message.Code)
	return m.send(ctx, message.Target, subject, body)
}

func (m *SMTPMailer) SendPasswordReset(ctx context.Context, message auth.PasswordResetDeliveryMessage) error {
	subject := "Password reset"
	body := fmt.Sprintf("Your password reset token is %s", message.Token)
	return m.send(ctx, message.Target, subject, body)
}

func (m *SMTPMailer) send(ctx context.Context, to, subject, body string) error {
	if m.cfg.Host == "" || m.cfg.Port == 0 || m.cfg.From == "" {
		return errors.New("smtp is not configured")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	addr := fmt.Sprintf("%s:%d", m.cfg.Host, m.cfg.Port)
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", m.cfg.From),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	client, err := m.dialTLS(ctx, addr)
	if err != nil {
		return err
	}
	defer func() {
		_ = client.Quit()
		_ = client.Close()
	}()

	if err := client.Hello("localhost"); err != nil {
		return err
	}

	if m.cfg.Port != 465 {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			return errors.New("smtp server does not support STARTTLS")
		}
		if err := client.StartTLS(m.clientTLSConfig()); err != nil {
			return err
		}
	}

	if m.cfg.Username != "" {
		auth := smtp.PlainAuth("", m.cfg.Username, m.cfg.Password, m.cfg.Host)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}

	if err := client.Mail(m.cfg.From); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		_ = w.Close()
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return nil
}

func (m *SMTPMailer) clientTLSConfig() *tls.Config {
	if m.tlsConfig != nil {
		return m.tlsConfig.Clone()
	}
	return &tls.Config{
		ServerName: m.cfg.Host,
		MinVersion: tls.VersionTLS12,
	}
}

func (m *SMTPMailer) dialTLS(ctx context.Context, addr string) (*smtp.Client, error) {
	dialer := &net.Dialer{}
	if deadline, ok := ctx.Deadline(); ok {
		dialer.Deadline = deadline
	}

	if m.cfg.Port == 465 {
		conn, err := tls.DialWithDialer(dialer, "tcp", addr, m.clientTLSConfig())
		if err != nil {
			if ctx.Err() != nil {
				return nil, ctx.Err()
			}
			return nil, err
		}
		if err := ctx.Err(); err != nil {
			_ = conn.Close()
			return nil, err
		}
		return smtp.NewClient(conn, m.cfg.Host)
	}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return smtp.NewClient(conn, m.cfg.Host)
}

var _ auth.VerificationMailer = (*SMTPMailer)(nil)
var _ auth.PasswordResetMailer = (*SMTPMailer)(nil)
