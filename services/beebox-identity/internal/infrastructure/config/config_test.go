package config

import (
	"errors"
	"testing"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name    string
		lookup  LookupFunc
		want    Config
		wantErr error
	}{
		{
			name: "default port",
			lookup: func(key string) (string, bool) {
				if key == "DATABASE_URL" {
					return "postgres://identity:identity@localhost:5432/identity?sslmode=disable", true
				}
				return "", false
			},
			want: Config{
				Port:        "8081",
				DatabaseURL: "postgres://identity:identity@localhost:5432/identity?sslmode=disable",
			},
		},
		{
			name: "missing database url",
			lookup: func(string) (string, bool) {
				return "", false
			},
			wantErr: ErrMissingDatabaseURL,
		},
		{
			name: "invalid port",
			lookup: func(key string) (string, bool) {
				if key == "PORT" {
					return "65536", true
				}
				if key == "DATABASE_URL" {
					return "postgres://identity:identity@localhost:5432/identity?sslmode=disable", true
				}
				return "", false
			},
			wantErr: ErrInvalidPort,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Load(test.lookup)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}
			if test.wantErr == nil && got.Port != test.want.Port {
				t.Fatalf("expected port %s, got %s", test.want.Port, got.Port)
			}
		})
	}
}
