package config

import (
	"errors"
	"strconv"
)

var ErrInvalidPort = errors.New("PORT must be a valid port number")

type LookupFunc func(key string) (string, bool)

type Config struct {
	Port string
}

func Load(lookup LookupFunc) (Config, error) {
	port, ok := lookup("PORT")
	if !ok || port == "" {
		port = "8081"
	}

	if err := validatePort(port); err != nil {
		return Config{}, err
	}

	return Config{
		Port: port,
	}, nil
}

func validatePort(port string) error {
	n, err := strconv.Atoi(port)
	if err != nil || n < 1 || n > 65535 {
		return ErrInvalidPort
	}
	return nil
}