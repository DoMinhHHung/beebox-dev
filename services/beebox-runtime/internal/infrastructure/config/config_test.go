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

func TestLoad_RequiresProjectURLAndToken(t *testing.T) {
	_, err := config.Load(lookup(map[string]string{}))
	if err != config.ErrMissingProjectURL {
		t.Fatalf("got %v", err)
	}
	_, err = config.Load(lookup(map[string]string{
		"PROJECT_BASE_URL": "http://127.0.0.1:8082",
	}))
	if err != config.ErrMissingInternalToken {
		t.Fatalf("got %v", err)
	}
}

func TestLoad_OK(t *testing.T) {
	got, err := config.Load(lookup(map[string]string{
		"PROJECT_BASE_URL":      "http://127.0.0.1:8082",
		"BEEBOX_INTERNAL_TOKEN": "tok",
		"PORT":                  "9090",
	}))
	if err != nil {
		t.Fatal(err)
	}
	if got.Port != "9090" || got.InternalToken != "tok" {
		t.Fatalf("%#v", got)
	}
}
