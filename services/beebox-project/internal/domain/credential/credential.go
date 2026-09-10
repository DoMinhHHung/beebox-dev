package credential

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

type Kind string

const (
	KindPublic Kind = "PUBLIC"
	KindSecret Kind = "SECRET"
)

type Status string

const (
	StatusActive  Status = "ACTIVE"
	StatusRevoked Status = "REVOKED"
	StatusExpired Status = "EXPIRED"
)

var (
	ErrInvalidCredential = errors.New("invalid credential")
	ErrInvalidTransition = errors.New("invalid credential transition")
)

type Credential struct {
	ProjectID  string
	ID         string
	Name       string
	Kind       Kind
	Value      string
	SecretHash string
	Status     Status
	ExpiresAt  *time.Time
}

type IssuedCredential struct {
	Credential Credential
	Secret     string
}

var _ fmt.Stringer = IssuedCredential{}

func New(
	projectID string,
	id string,
	name string,
	kind Kind,
) (IssuedCredential, error) {
	if projectID == "" || id == "" || name == "" {
		return IssuedCredential{}, ErrInvalidCredential
	}

	if kind != KindPublic && kind != KindSecret {
		return IssuedCredential{}, ErrInvalidCredential
	}

	raw, err := generateSecret()
	if err != nil {
		return IssuedCredential{}, ErrInvalidCredential
	}

	credentialValue := Credential{
		ProjectID: projectID,
		ID:        id,
		Name:      name,
		Kind:      kind,
		Status:    StatusActive,
	}

	switch kind {
	case KindPublic:
		credentialValue.Value = raw
	case KindSecret:
		credentialValue.SecretHash = hashSecret(raw)
	}

	return IssuedCredential{
		Credential: credentialValue,
		Secret:     raw,
	}, nil
}

func (i IssuedCredential) String() string {
	return fmt.Sprintf(
		"IssuedCredential{ProjectID:%s ID:%s Name:%s Kind:%s Status:%s Secret:REDACTED}",
		i.Credential.ProjectID,
		i.Credential.ID,
		i.Credential.Name,
		i.Credential.Kind,
		i.Credential.Status,
	)
}

func (c Credential) WithExpiry(expiresAt time.Time) (Credential, error) {
	if c.Status != StatusActive {
		return Credential{}, ErrInvalidTransition
	}

	if expiresAt.IsZero() {
		return Credential{}, ErrInvalidCredential
	}

	c.ExpiresAt = &expiresAt
	return c, nil
}

func (c Credential) Revoke() (Credential, error) {
	if c.Status != StatusActive {
		return Credential{}, ErrInvalidTransition
	}

	c.Status = StatusRevoked
	return c, nil
}

func (c Credential) IsExpired(now time.Time) bool {
	if c.ExpiresAt == nil {
		return false
	}

	return !now.Before(*c.ExpiresAt)
}

func (c Credential) Expire(now time.Time) (Credential, error) {
	if c.Status != StatusActive {
		return Credential{}, ErrInvalidTransition
	}

	if !c.IsExpired(now) {
		return Credential{}, ErrInvalidTransition
	}

	c.Status = StatusExpired
	return c, nil
}

func (c Credential) Matches(candidate string) bool {
	if c.Status != StatusActive || c.IsExpired(time.Now()) {
		return false
	}

	if candidate == "" {
		return false
	}

	switch c.Kind {
	case KindPublic:
		return subtle.ConstantTimeCompare(
			[]byte(c.Value),
			[]byte(candidate),
		) == 1
	case KindSecret:
		return subtle.ConstantTimeCompare(
			[]byte(hashSecret(candidate)),
			[]byte(c.SecretHash),
		) == 1
	default:
		return false
	}
}

func generateSecret() (string, error) {
	buf := make([]byte, 32)

	if _, err := rand.Read(buf); err != nil {
		return "", err
	}

	return hex.EncodeToString(buf), nil
}

func hashSecret(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
