package credential

import (
	"crypto/rand"
	"encoding/base64"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"golang.org/x/crypto/bcrypt"
)

func randomToken(nBytes int) (string, error) {
	buf := make([]byte, nBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", apperror.Wrap(apperror.CodeInternal, "failed to generate random token", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

func newPublicKey(env Environment) (string, error) {
	token, err := randomToken(24)
	if err != nil {
		return "", err
	}
	return "pk_" + string(env) + "_" + token, nil
}

func newSecret(env Environment) (plaintext string, hash string, err error) {
	token, err := randomToken(32)
	if err != nil {
		return "", "", err
	}
	plaintext = "sk_" + string(env) + "_" + token
	hashBytes, hashErr := bcrypt.GenerateFromPassword([]byte(plaintext), bcrypt.DefaultCost)
	if hashErr != nil {
		return "", "", apperror.Wrap(apperror.CodeInternal, "failed to hash secret key", hashErr)
	}
	return plaintext, string(hashBytes), nil
}