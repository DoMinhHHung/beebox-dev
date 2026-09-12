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
				switch key {
				case "DATABASE_URL":
					return "postgres://identity:identity@localhost:5432/identity?sslmode=disable", true
				case "VERIFICATION_CODE_SECRET":
					return "test-verification-secret", true
				default:
					return "", false
				}
			},
			want: Config{
				Port:                   "8081",
				DatabaseURL:            "postgres://identity:identity@localhost:5432/identity?sslmode=disable",
				VerificationCodeSecret: "test-verification-secret",
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
			name: "whitespace only database url",
			lookup: func(key string) (string, bool) {
				if key == "DATABASE_URL" {
					return "   \t  ", true
				}
				return "", false
			},
			wantErr: ErrMissingDatabaseURL,
		},
		{
			name: "trims surrounding whitespace",
			lookup: func(key string) (string, bool) {
				switch key {
				case "DATABASE_URL":
					return "  postgres://identity:identity@localhost:5432/identity?sslmode=disable  ", true
				case "VERIFICATION_CODE_SECRET":
					return "  secret-value  ", true
				default:
					return "", false
				}
			},
			want: Config{
				Port:                   "8081",
				DatabaseURL:            "postgres://identity:identity@localhost:5432/identity?sslmode=disable",
				VerificationCodeSecret: "secret-value",
			},
		},
		{
			name: "invalid port",
			lookup: func(key string) (string, bool) {
				switch key {
				case "PORT":
					return "65536", true
				case "DATABASE_URL":
					return "postgres://identity:identity@localhost:5432/identity?sslmode=disable", true
				case "VERIFICATION_CODE_SECRET":
					return "secret", true
				default:
					return "", false
				}
			},
			wantErr: ErrInvalidPort,
		},
		{
			name: "missing verification code secret",
			lookup: func(key string) (string, bool) {
				if key == "DATABASE_URL" {
					return "postgres://identity:identity@localhost:5432/identity?sslmode=disable", true
				}
				return "", false
			},
			wantErr: ErrMissingVerificationCodeSecret,
		},
		{
			name: "whitespace only verification code secret",
			lookup: func(key string) (string, bool) {
				switch key {
				case "DATABASE_URL":
					return "postgres://identity:identity@localhost:5432/identity?sslmode=disable", true
				case "VERIFICATION_CODE_SECRET":
					return "  \t ", true
				default:
					return "", false
				}
			},
			wantErr: ErrMissingVerificationCodeSecret,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := Load(test.lookup)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("expected error %v, got %v", test.wantErr, err)
			}
			if test.wantErr == nil {
				if got.Port != test.want.Port {
					t.Fatalf("expected port %s, got %s", test.want.Port, got.Port)
				}
				if got.DatabaseURL != test.want.DatabaseURL {
					t.Fatalf("expected database url %q, got %q", test.want.DatabaseURL, got.DatabaseURL)
				}
				if got.VerificationCodeSecret != test.want.VerificationCodeSecret {
					t.Fatalf("expected verification secret %q, got %q", test.want.VerificationCodeSecret, got.VerificationCodeSecret)
				}
			}
		})
	}
}
