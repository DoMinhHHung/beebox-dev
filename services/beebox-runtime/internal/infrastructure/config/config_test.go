package config_test

import (
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/internal/infrastructure/config"
)

func lookup(m map[string]string) config.LookupFunc {
	return func(k string) (string, bool) {
		v, ok := m[k]
		return v, ok
	}
}

func TestLoad_OK(t *testing.T) {
	got, err := config.Load(lookup(map[string]string{
		"PROJECT_BASE_URL":      "http://127.0.0.1:8082",
		"IDENTITY_BASE_URL":     "http://127.0.0.1:8081",
		"BEEBOX_INTERNAL_TOKEN": "tok",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.IdentityBaseURL != "http://127.0.0.1:8081" {
		t.Fatalf("%#v", got)
	}
}

func TestLoad_RequiresIdentity(t *testing.T) {
	_, err := config.Load(lookup(map[string]string{
		"PROJECT_BASE_URL":      "http://127.0.0.1:8082",
		"BEEBOX_INTERNAL_TOKEN": "tok",
	}))
	if err != config.ErrMissingIdentityURL {
		t.Fatalf("got %v", err)
	}
}
