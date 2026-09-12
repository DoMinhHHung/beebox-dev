package identity

import (
	"strings"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain"
)

type Identifier string

func NewIdentifier(value string) (Identifier, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", domain.ErrInvalidIdentityIdentifier
	}
	return Identifier(value), nil
}

func (id Identifier) String() string {
	return string(id)
}

func (id Identifier) IsZero() bool {
	return id == ""
}
