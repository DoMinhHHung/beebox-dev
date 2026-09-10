package config_test

import (
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/config"
)

func lookupFrom(values map[string]string) config.LookupFunc {
	return func(key string) (string, bool) {
		v, ok := values[key]
		return v, ok
	}
}

func TestLoad_ReturnsConfigWhenValid(t *testing.T) {
	got, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"PORT":         "9090",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.DatabaseURL != "postgres://user:pass@localhost:5432/beebox" {
		t.Fatalf("unexpected DatabaseURL: %q", got.DatabaseURL)
	}
	if got.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", got.Port)
	}
}

func TestLoad_DefaultsPortWhenMissing(t *testing.T) {
	got, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
	}))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Port != "8080" {
		t.Fatalf("expected default port 8080, got %q", got.Port)
	}
}

func TestLoad_RejectsMissingDatabaseURL(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"PORT": "9090",
	}))
	if err != config.ErrMissingDatabaseURL {
		t.Fatalf("expected ErrMissingDatabaseURL, got %v", err)
	}
}

func TestLoad_RejectsEmptyDatabaseURL(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL": "",
	}))
	if err != config.ErrMissingDatabaseURL {
		t.Fatalf("expected ErrMissingDatabaseURL, got %v", err)
	}
}

func TestLoad_RejectsNonNumericPort(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"PORT":         "not-a-port",
	}))
	if err != config.ErrInvalidPort {
		t.Fatalf("expected ErrInvalidPort, got %v", err)
	}
}

func TestLoad_RejectsOutOfRangePort(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"PORT":         "70000",
	}))
	if err != config.ErrInvalidPort {
		t.Fatalf("expected ErrInvalidPort, got %v", err)
	}
}
