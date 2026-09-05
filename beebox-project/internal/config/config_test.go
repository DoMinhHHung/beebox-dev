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

func TestConfig_Validate_EmptyDatabaseURLIsAllowed(t *testing.T) {
	cfg := Config{HTTPPort: "8080", ShutdownTimeout: 10 * time.Second, LogLevel: "info", DatabaseURL: ""}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected empty DatabaseURL to be valid, got error: %v", err)
	}
}

func TestConfig_Validate_InvalidDatabaseURL(t *testing.T) {
	cfg := Config{HTTPPort: "8080", ShutdownTimeout: 10 * time.Second, LogLevel: "info", DatabaseURL: "not-a-valid-url"}
	err := cfg.Validate()
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestConfig_Validate_ValidDatabaseURL(t *testing.T) {
	cfg := Config{
		HTTPPort:        "8080",
		ShutdownTimeout: 10 * time.Second,
		LogLevel:        "info",
		DatabaseURL:     "postgres://user:pass@localhost:5432/dbname",
	}
	if err := cfg.Validate(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}