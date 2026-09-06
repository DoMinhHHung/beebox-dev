package credential

import (
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	"golang.org/x/crypto/bcrypt"
)

type Environment string

const (
	EnvironmentTest Environment = "test"
	EnvironmentLive Environment = "live"
)

func (e Environment) Valid() bool {
	switch e {
	case EnvironmentTest, EnvironmentLive:
		return true
	default:
		return false
	}
}

type Status string

const (
	StatusActive  Status = "ACTIVE"
	StatusRevoked Status = "REVOKED"
)

type Credential struct {
	ID          string
	ProjectID   string
	Environment Environment
	PublicKey   string
	SecretHash  string
	Status      Status
	CreatedAt   time.Time
	RevokedAt   *time.Time
}

type Metadata struct {
	ID          string
	ProjectID   string
	Environment Environment
	PublicKey   string
	Status      Status
	CreatedAt   time.Time
	RevokedAt   *time.Time
}

func (c Credential) Metadata() Metadata {
	return Metadata{
		ID:          c.ID,
		ProjectID:   c.ProjectID,
		Environment: c.Environment,
		PublicKey:   c.PublicKey,
		Status:      c.Status,
		CreatedAt:   c.CreatedAt,
		RevokedAt:   c.RevokedAt,
	}
}

// NewCredential tạo credential đang hoạt động cho dự án và môi trường được chỉ định.
// Trả về credential, secret dạng văn bản để sử dụng ban đầu và lỗi nếu dữ liệu đầu vào
// không hợp lệ hoặc không thể tạo credential.
func NewCredential(projectID string, env Environment) (Credential, string, error) {
	if projectID == "" {
		return Credential{}, "", apperror.New(apperror.CodeInvalidInput, "projectID must not be empty")
	}
	if !env.Valid() {
		return Credential{}, "", apperror.New(apperror.CodeInvalidInput, "unknown environment: "+string(env))
	}

	id, err := idgen.New()
	if err != nil {
		return Credential{}, "", apperror.Wrap(apperror.CodeInternal, "failed to generate credential id", err)
	}

	publicKey, err := newPublicKey(env)
	if err != nil {
		return Credential{}, "", err
	}

	plaintext, hash, err := newSecret(env)
	if err != nil {
		return Credential{}, "", err
	}

	c := Credential{
		ID:          id,
		ProjectID:   projectID,
		Environment: env,
		PublicKey:   publicKey,
		SecretHash:  hash,
		Status:      StatusActive,
		CreatedAt:   time.Now().UTC(),
	}
	return c, plaintext, nil
}

func (c Credential) VerifySecret(plaintext string) error {
	if c.Status == StatusRevoked {
		return apperror.New(apperror.CodeCredentialRevoked, "credential has been revoked")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(c.SecretHash), []byte(plaintext)); err != nil {
		return apperror.New(apperror.CodeCredentialInvalid, "secret key does not match")
	}
	return nil
}

func (c Credential) Rotate() (Credential, string, error) {
	if c.Status == StatusRevoked {
		return Credential{}, "", apperror.New(apperror.CodeCredentialRevoked, "cannot rotate a revoked credential")
	}

	plaintext, hash, err := newSecret(c.Environment)
	if err != nil {
		return Credential{}, "", err
	}

	rotated := c
	rotated.SecretHash = hash
	return rotated, plaintext, nil
}

func (c Credential) Revoke() (Credential, error) {
	if c.Status == StatusRevoked {
		return Credential{}, apperror.New(apperror.CodeCredentialRevoked, "credential already revoked")
	}
	now := time.Now().UTC()
	revoked := c
	revoked.Status = StatusRevoked
	revoked.RevokedAt = &now
	return revoked, nil
}
