package config

import (
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.HTTPPort != "8080" {
		t.Fatalf("expected default HTTPPort 8080, got %s", cfg.HTTPPort)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("expected default LogLevel info, got %s", cfg.LogLevel)
	}
	if cfg.ShutdownTimeout != 10*time.Second {
		t.Fatalf("expected default ShutdownTimeout 10s, got %v", cfg.ShutdownTimeout)
	}
	if cfg.OwnerSessionTTL != 24*time.Hour {
		t.Fatalf("expected default OwnerSessionTTL 24h, got %v", cfg.OwnerSessionTTL)
	}
}

func TestLoad_ValidOverride(t *testing.T) {
	t.Setenv("HTTP_PORT", "9090")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.HTTPPort != "9090" {
		t.Fatalf("expected HTTPPort 9090, got %s", cfg.HTTPPort)
	}
}

func TestLoad_InvalidPort(t *testing.T) {
	t.Setenv("HTTP_PORT", "abc")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid HTTP_PORT")
	}
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %s", apperror.CodeOf(err))
	}
}

func TestLoad_InvalidLogLevel(t *testing.T) {
	t.Setenv("LOG_LEVEL", "verbose")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid LOG_LEVEL")
	}
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %s", apperror.CodeOf(err))
	}
}

func TestLoad_InvalidShutdownTimeout(t *testing.T) {
	t.Setenv("SHUTDOWN_TIMEOUT_SECONDS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for zero shutdown timeout")
	}
}

func TestLoad_DefaultOwnerSessionTTL(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if cfg.OwnerSessionTTL != 24*time.Hour {
		t.Fatalf("expected 24h, got %v", cfg.OwnerSessionTTL)
	}
}

func TestLoad_InvalidOwnerSessionTTL(t *testing.T) {
	t.Setenv("OWNER_SESSION_TTL_HOURS", "0")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for zero session ttl")
	}
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %s", apperror.CodeOf(err))
	}
}

func TestLoad_InvalidDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "not-a-valid-url")

	_, err := Load()
	if err == nil {
		t.Fatal("expected error for invalid DATABASE_URL")
	}
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %s", apperror.CodeOf(err))
	}
}
