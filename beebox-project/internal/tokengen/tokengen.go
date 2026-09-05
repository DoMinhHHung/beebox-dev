package tokengen

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func New(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to generate random token", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func Hash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}