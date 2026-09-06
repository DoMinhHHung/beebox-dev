package credential

import (
	"reflect"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func TestNewCredential_Success(t *testing.T) {
	c, secret, err := NewCredential("proj-1", EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Status != StatusActive {
		t.Fatalf("expected StatusActive, got %s", c.Status)
	}
	if c.SecretHash == secret {
		t.Fatal("SecretHash must not equal plaintext secret")
	}
	wantPKPrefix := "pk_test_"
	if len(c.PublicKey) <= len(wantPKPrefix) || c.PublicKey[:len(wantPKPrefix)] != wantPKPrefix {
		t.Fatalf("expected PublicKey to have prefix %q, got %q", wantPKPrefix, c.PublicKey)
	}
	wantSKPrefix := "sk_test_"
	if len(secret) <= len(wantSKPrefix) || secret[:len(wantSKPrefix)] != wantSKPrefix {
		t.Fatalf("expected secret to have prefix %q, got %q", wantSKPrefix, secret)
	}
}

func TestNewCredential_InvalidProjectID(t *testing.T) {
	_, _, err := NewCredential("", EnvironmentTest)
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestNewCredential_InvalidEnvironment(t *testing.T) {
	_, _, err := NewCredential("proj-1", Environment("staging"))
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %v", apperror.CodeOf(err))
	}
}

func TestCredential_VerifySecret_Success(t *testing.T) {
	c, secret, err := NewCredential("proj-1", EnvironmentLive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := c.VerifySecret(secret); err != nil {
		t.Fatalf("expected valid secret, got error: %v", err)
	}
}

func TestCredential_VerifySecret_WrongSecret(t *testing.T) {
	c, _, err := NewCredential("proj-1", EnvironmentLive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = c.VerifySecret("sk_live_wrongvalue")
	if apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected CodeCredentialInvalid, got %v", apperror.CodeOf(err))
	}
}

func TestCredential_VerifySecret_RevokedTakesPrecedence(t *testing.T) {
	c, secret, err := NewCredential("proj-1", EnvironmentLive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	revoked, err := c.Revoke()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	err = revoked.VerifySecret(secret)
	if apperror.CodeOf(err) != apperror.CodeCredentialRevoked {
		t.Fatalf("expected CodeCredentialRevoked even with correct plaintext, got %v", apperror.CodeOf(err))
	}
}

func TestCredential_Rotate_ChangesSecretKeepsIdentity(t *testing.T) {
	c, oldSecret, err := NewCredential("proj-1", EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	rotated, newSecret, err := c.Rotate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rotated.ID != c.ID || rotated.ProjectID != c.ProjectID || rotated.PublicKey != c.PublicKey || rotated.Environment != c.Environment {
		t.Fatal("Rotate must preserve ID, ProjectID, PublicKey, Environment")
	}
	if newSecret == oldSecret {
		t.Fatal("Rotate must produce a new plaintext secret")
	}
	if err := rotated.VerifySecret(newSecret); err != nil {
		t.Fatalf("expected new secret to verify, got error: %v", err)
	}
	if err := rotated.VerifySecret(oldSecret); apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected old secret to be invalid after rotate, got %v", apperror.CodeOf(err))
	}
}

func TestCredential_Rotate_RevokedFails(t *testing.T) {
	c, _, err := NewCredential("proj-1", EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	revoked, err := c.Revoke()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	_, _, err = revoked.Rotate()
	if apperror.CodeOf(err) != apperror.CodeCredentialRevoked {
		t.Fatalf("expected CodeCredentialRevoked, got %v", apperror.CodeOf(err))
	}
}

func TestCredential_Revoke_Idempotent_SecondCallFails(t *testing.T) {
	c, _, err := NewCredential("proj-1", EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	revoked, err := c.Revoke()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revoked.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}
	_, err = revoked.Revoke()
	if apperror.CodeOf(err) != apperror.CodeCredentialRevoked {
		t.Fatalf("expected CodeCredentialRevoked on second revoke, got %v", apperror.CodeOf(err))
	}
}

func TestCredential_Metadata_DoesNotExposeSecretHash(t *testing.T) {
	mdType := reflect.TypeOf(Metadata{})
	for i := 0; i < mdType.NumField(); i++ {
		if mdType.Field(i).Name == "SecretHash" {
			t.Fatal("Metadata must not expose a SecretHash field")
		}
	}

	c, _, err := NewCredential("proj-1", EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	md := c.Metadata()
	if md.ID != c.ID || md.ProjectID != c.ProjectID || md.PublicKey != c.PublicKey || md.Status != c.Status {
		t.Fatal("Metadata must mirror the non-secret fields of Credential")
	}
}
