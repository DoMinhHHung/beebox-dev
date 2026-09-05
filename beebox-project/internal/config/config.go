package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	HTTPPort        string
	ShutdownTimeout time.Duration
	LogLevel        string
	DatabaseURL     string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		LogLevel:    getEnv("LOG_LEVEL", "info"),
		DatabaseURL: getEnv("DATABASE_URL", ""),
	}

	timeoutSeconds, err := strconv.Atoi(getEnv("SHUTDOWN_TIMEOUT_SECONDS", "10"))
	if err != nil {
		return Config{}, apperror.Wrap(apperror.CodeInvalidInput, "SHUTDOWN_TIMEOUT_SECONDS must be an integer", err)
	}
	cfg.ShutdownTimeout = time.Duration(timeoutSeconds) * time.Second

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	port, err := strconv.Atoi(c.HTTPPort)
	if err != nil || port < 1 || port > 65535 {
		return apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("HTTP_PORT must be between 1 and 65535, got %q", c.HTTPPort))
	}

	if c.ShutdownTimeout <= 0 {
		return apperror.New(apperror.CodeInvalidInput, "SHUTDOWN_TIMEOUT_SECONDS must be greater than zero")
	}

	switch c.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		return apperror.New(apperror.CodeInvalidInput, fmt.Sprintf("LOG_LEVEL must be one of debug/info/warn/error, got %q", c.LogLevel))
	}

	if c.DatabaseURL != "" {
		if _, err := pgxpool.ParseConfig(c.DatabaseURL); err != nil {
			return apperror.New(apperror.CodeInvalidInput, "DATABASE_URL is not a valid postgres connection string")
		}
	}

	return nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}