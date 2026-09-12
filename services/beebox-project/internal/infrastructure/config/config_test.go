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

func withToken(values map[string]string) map[string]string {
	if _, ok := values["BEEBOX_INTERNAL_TOKEN"]; !ok {
		values["BEEBOX_INTERNAL_TOKEN"] = "test-internal-token"
	}
	return values
}

func TestLoad_ReturnsConfigWhenValid(t *testing.T) {
	got, err := config.Load(lookupFrom(withToken(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"PORT":         "9090",
	})))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.DatabaseURL != "postgres://user:pass@localhost:5432/beebox" {
		t.Fatalf("unexpected DatabaseURL: %q", got.DatabaseURL)
	}
	if got.Port != "9090" {
		t.Fatalf("expected port 9090, got %q", got.Port)
	}
	if got.InternalToken != "test-internal-token" {
		t.Fatalf("unexpected InternalToken: %q", got.InternalToken)
	}
}

func TestLoad_DefaultsPortWhenMissing(t *testing.T) {
	got, err := config.Load(lookupFrom(withToken(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
	})))
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
	_, err := config.Load(lookupFrom(withToken(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"PORT":         "not-a-port",
	})))
	if err != config.ErrInvalidPort {
		t.Fatalf("expected ErrInvalidPort, got %v", err)
	}
}

func TestLoad_RejectsOutOfRangePort(t *testing.T) {
	_, err := config.Load(lookupFrom(withToken(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"PORT":         "70000",
	})))
	if err != config.ErrInvalidPort {
		t.Fatalf("expected ErrInvalidPort, got %v", err)
	}
}

func TestLoad_RejectsNonLocalHTTPIdentityURL(t *testing.T) {
	_, err := config.Load(lookupFrom(withToken(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"IDENTITY_URL": "http://identity.example.com",
	})))
	if err != config.ErrInvalidIdentityURL {
		t.Fatalf("expected ErrInvalidIdentityURL, got %v", err)
	}
}

func TestLoad_AllowsLocalHTTPIdentityURL(t *testing.T) {
	got, err := config.Load(lookupFrom(withToken(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
		"IDENTITY_URL": "http://127.0.0.1:8081",
	})))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.IdentityURL != "http://127.0.0.1:8081" {
		t.Fatalf("unexpected IdentityURL %q", got.IdentityURL)
	}
}

func TestLoad_RejectsMissingInternalToken(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL": "postgres://user:pass@localhost:5432/beebox",
	}))
	if err != config.ErrMissingInternalToken {
		t.Fatalf("expected ErrMissingInternalToken, got %v", err)
	}
}

func TestLoad_RejectsEmptyInternalToken(t *testing.T) {
	_, err := config.Load(lookupFrom(map[string]string{
		"DATABASE_URL":          "postgres://user:pass@localhost:5432/beebox",
		"BEEBOX_INTERNAL_TOKEN": "   ",
	}))
	if err != config.ErrMissingInternalToken {
		t.Fatalf("expected ErrMissingInternalToken, got %v", err)
	}
}
