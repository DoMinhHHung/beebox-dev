package delivery

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"errors"
	"math/big"
	"net"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
)

func TestSMTPMailerRejectsUnconfigured(t *testing.T) {
	m := NewSMTPMailer(SMTPConfig{})
	err := m.SendVerificationEmail(context.Background(), auth.VerificationDeliveryMessage{
		Target: "a@b.com",
		Code:   "123456",
	})
	if err == nil {
		t.Fatal("expected configuration error")
	}
}

func TestSMTPMailerContextCanceledBeforeDial(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	m := NewSMTPMailer(SMTPConfig{Host: "127.0.0.1", Port: 2525, From: "noreply@example.com"})
	err := m.SendVerificationEmail(ctx, auth.VerificationDeliveryMessage{Target: "a@b.com", Code: "1"})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
	}
}

func TestSMTPMailerRequiresSTARTTLS(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = conn.Write([]byte("220 test ESMTP\r\n"))
		buf := make([]byte, 512)
		n, _ := conn.Read(buf)
		if strings.HasPrefix(string(buf[:n]), "EHLO") || strings.HasPrefix(string(buf[:n]), "HELO") {
			_, _ = conn.Write([]byte("250-test\r\n250 HELP\r\n"))
		}
	}()

	m := NewSMTPMailer(SMTPConfig{
		Host: "127.0.0.1",
		Port: mustPort(ln.Addr().String()),
		From: "noreply@example.com",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	err = m.SendVerificationEmail(ctx, auth.VerificationDeliveryMessage{Target: "a@b.com", Code: "123456"})
	if err == nil {
		t.Fatal("expected STARTTLS requirement failure")
	}
	if !strings.Contains(err.Error(), "STARTTLS") {
		t.Fatalf("expected STARTTLS error, got %v", err)
	}
	wg.Wait()
}

func TestSMTPMailerSTARTTLSAndAuthAfterTLS(t *testing.T) {
	cert, err := selfSignedCert()
	if err != nil {
		t.Fatalf("cert: %v", err)
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer ln.Close()

	var (
		mu          sync.Mutex
		authSeen    bool
		tlsUpgraded bool
		mailFrom    string
		rcptTo      string
		dataBody    string
	)

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		c := conn
		write := func(s string) { _, _ = c.Write([]byte(s)) }
		readLine := func() string {
			buf := make([]byte, 1024)
			n, _ := c.Read(buf)
			return string(buf[:n])
		}
		write("220 test ESMTP\r\n")
		_ = readLine()
		write("250-test\r\n250-STARTTLS\r\n250 AUTH PLAIN\r\n")
		line := readLine()
		if !strings.HasPrefix(line, "STARTTLS") {
			return
		}
		write("220 Ready to start TLS\r\n")
		tlsConn := tls.Server(c, &tls.Config{Certificates: []tls.Certificate{cert}})
		if err := tlsConn.Handshake(); err != nil {
			return
		}
		mu.Lock()
		tlsUpgraded = true
		mu.Unlock()
		c = tlsConn
		_ = readLine()
		write("250-test\r\n250 AUTH PLAIN LOGIN\r\n")
		line = readLine()
		if strings.HasPrefix(line, "AUTH") {
			mu.Lock()
			authSeen = true
			mu.Unlock()
			write("235 OK\r\n")
			line = readLine()
		}
		if strings.HasPrefix(line, "MAIL FROM:") {
			mu.Lock()
			mailFrom = line
			mu.Unlock()
			write("250 OK\r\n")
			line = readLine()
		}
		if strings.HasPrefix(line, "RCPT TO:") {
			mu.Lock()
			rcptTo = line
			mu.Unlock()
			write("250 OK\r\n")
			line = readLine()
		}
		if strings.HasPrefix(line, "DATA") {
			write("354 End data with <CR><LF>.<CR><LF>\r\n")
			var body strings.Builder
			for {
				chunk := readLine()
				body.WriteString(chunk)
				if strings.Contains(chunk, "\r\n.\r\n") || strings.HasSuffix(chunk, ".\r\n") {
					break
				}
			}
			mu.Lock()
			dataBody = body.String()
			mu.Unlock()
			write("250 OK\r\n")
		}
	}()

	m := NewSMTPMailer(SMTPConfig{
		Host:     "127.0.0.1",
		Port:     mustPort(ln.Addr().String()),
		Username: "user",
		Password: "pass",
		From:     "noreply@example.com",
	})
	m.tlsConfig = &tls.Config{
		InsecureSkipVerify: true,
		MinVersion:         tls.VersionTLS12,
		ServerName:         "127.0.0.1",
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	err = m.SendPasswordReset(ctx, auth.PasswordResetDeliveryMessage{
		Target: "reset@example.com",
		Token:  "tok-1",
	})
	<-done
	if err != nil {
		t.Fatalf("send: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if !tlsUpgraded {
		t.Fatal("expected TLS upgrade via STARTTLS")
	}
	if !authSeen {
		t.Fatal("expected AUTH after TLS")
	}
	if !strings.Contains(mailFrom, "noreply@example.com") {
		t.Fatalf("unexpected mail from %q", mailFrom)
	}
	if !strings.Contains(rcptTo, "reset@example.com") {
		t.Fatalf("unexpected rcpt %q", rcptTo)
	}
	if !strings.Contains(dataBody, "tok-1") {
		t.Fatalf("expected token in body, got %q", dataBody)
	}
}

func TestSMTPMailerDialRespectsDeadline(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	addr := ln.Addr().String()
	_ = ln.Close()

	m := NewSMTPMailer(SMTPConfig{
		Host: "127.0.0.1",
		Port: mustPort(addr),
		From: "noreply@example.com",
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	err = m.SendVerificationEmail(ctx, auth.VerificationDeliveryMessage{Target: "a@b.com", Code: "1"})
	if err == nil {
		t.Fatal("expected dial failure under short deadline")
	}
}

func mustPort(addr string) int {
	_, port, err := net.SplitHostPort(addr)
	if err != nil {
		panic(err)
	}
	var p int
	for _, c := range port {
		p = p*10 + int(c-'0')
	}
	return p
}

func selfSignedCert() (tls.Certificate, error) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return tls.Certificate{}, err
	}
	template := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject:      pkix.Name{CommonName: "127.0.0.1"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, template, template, &key.PublicKey, key)
	if err != nil {
		return tls.Certificate{}, err
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der})
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	return tls.X509KeyPair(certPEM, keyPEM)
}
