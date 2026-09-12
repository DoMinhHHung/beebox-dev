package config

import (
	"errors"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var (
	ErrInvalidPort          = errors.New("PORT must be a valid port number")
	ErrMissingProjectURL    = errors.New("PROJECT_BASE_URL is required")
	ErrInvalidProjectURL    = errors.New("PROJECT_BASE_URL must use https unless host is localhost")
	ErrMissingInternalToken = errors.New("BEEBOX_INTERNAL_TOKEN is required")
)

type LookupFunc func(key string) (string, bool)

type Config struct {
	Port           string
	ProjectBaseURL string
	InternalToken  string
	HTTPTimeout    time.Duration
}

func Load(lookup LookupFunc) (Config, error) {
	port, ok := lookup("PORT")
	if !ok || port == "" {
		port = "8084"
	}
	if err := validatePort(port); err != nil {
		return Config{}, err
	}

	projectBaseURL, ok := lookup("PROJECT_BASE_URL")
	if !ok || strings.TrimSpace(projectBaseURL) == "" {
		return Config{}, ErrMissingProjectURL
	}
	projectBaseURL = strings.TrimRight(strings.TrimSpace(projectBaseURL), "/")
	if err := validateServiceURL(projectBaseURL); err != nil {
		return Config{}, err
	}

	internalToken, ok := lookup("BEEBOX_INTERNAL_TOKEN")
	if !ok || strings.TrimSpace(internalToken) == "" {
		return Config{}, ErrMissingInternalToken
	}

	timeout := 3 * time.Second
	if raw, ok := lookup("HTTP_TIMEOUT_MS"); ok && strings.TrimSpace(raw) != "" {
		ms, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil || ms < 1 {
			return Config{}, errors.New("HTTP_TIMEOUT_MS must be a positive integer")
		}
		timeout = time.Duration(ms) * time.Millisecond
	}

	return Config{
		Port:           port,
		ProjectBaseURL: projectBaseURL,
		InternalToken:  strings.TrimSpace(internalToken),
		HTTPTimeout:    timeout,
	}, nil
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return ErrInvalidPort
	}
	return nil
}

func validateServiceURL(raw string) error {
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return ErrInvalidProjectURL
	}
	host := strings.ToLower(parsed.Hostname())
	local := host == "localhost" || host == "127.0.0.1" || host == "::1"
	if local {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return ErrInvalidProjectURL
		}
		return nil
	}
	if parsed.Scheme != "https" {
		return ErrInvalidProjectURL
	}
	return nil
}
