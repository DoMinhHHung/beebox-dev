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
			lookup: func(string) (string, bool) {
				return "", false
			},
			want: Config{Port: "8081"},
		},
		{
			name: "configured port",
			lookup: func(key string) (string, bool) {
				if key == "PORT" {
					return "9090", true
				}
				return "", false
			},
			want: Config{Port: "9090"},
		},
		{
			name: "invalid port",
			lookup: func(key string) (string, bool) {
				if key == "PORT" {
					return "65536", true
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
			if got != test.want {
				t.Fatalf("expected config %#v, got %#v", test.want, got)
			}
		})
	}
}
