package config

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
)

var (
	ErrMissingDatabaseURL = errors.New("DATABASE_URL is required")
	ErrInvalidPort        = errors.New("PORT must be a valid port number")
	ErrInvalidIdentityURL = errors.New("IDENTITY_URL must use https unless host is localhost")
)

type LookupFunc func(key string) (string, bool)

type Config struct {
	DatabaseURL string
	Port        string
	IdentityURL string
}

func Load(lookup LookupFunc) (Config, error) {
	databaseURL, ok := lookup("DATABASE_URL")
	if !ok || databaseURL == "" {
		return Config{}, ErrMissingDatabaseURL
	}

	port, ok := lookup("PORT")
	if !ok || port == "" {
		port = "8080"
	}

	if err := validatePort(port); err != nil {
		return Config{}, err
	}
	identityURL, ok := lookup("IDENTITY_URL")
	if !ok || identityURL == "" {
		identityURL = "http://localhost:8081"
	}
	if err := validateIdentityURL(identityURL); err != nil {
		return Config{}, err
	}

	return Config{
		DatabaseURL: databaseURL,
		Port:        port,
		IdentityURL: identityURL,
	}, nil
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return ErrInvalidPort
	}
	return nil
}

func validateIdentityURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidIdentityURL
	}
	host := strings.ToLower(parsed.Hostname())
	local := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if local {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return ErrInvalidIdentityURL
		}
		return nil
	}
	if parsed.Scheme != "https" {
		return ErrInvalidIdentityURL
	}
	return nil
}
