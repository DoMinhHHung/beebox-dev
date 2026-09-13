package credential_test

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/credential"
)

func TestNew_IssuesActivePublicCredential(t *testing.T) {
	issued, err := credential.New(
		"project-1",
		"credential-1",
		"web client",
		credential.KindPublic,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := issued.Credential

	if got.ProjectID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ProjectID)
	}

	if got.ID != "credential-1" {
		t.Fatalf("expected credential ID %q, got %q", "credential-1", got.ID)
	}

	if got.Name != "web client" {
		t.Fatalf("expected credential name %q, got %q", "web client", got.Name)
	}

	if got.Kind != credential.KindPublic {
		t.Fatalf("expected kind %q, got %q", credential.KindPublic, got.Kind)
	}

	if got.Status != credential.StatusActive {
		t.Fatalf("expected status %q, got %q", credential.StatusActive, got.Status)
	}

	if got.Value != "" {
		t.Fatal("expected public credential to never store the raw value in Value")
	}

	if got.SecretHash == "" {
		t.Fatal("expected public credential to store a secret hash")
	}

	if got.SecretHash == issued.Secret {
		t.Fatal("expected the stored hash to differ from the raw issued secret")
	}

	if !got.Matches(issued.Secret) {
		t.Fatal("expected public credential to match its issued secret")
	}
}

func TestNew_IssuesActiveSecretCredentialWithoutPersistingRawSecret(t *testing.T) {
	issued, err := credential.New(
		"project-1",
		"credential-2",
		"server key",
		credential.KindSecret,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got := issued.Credential

	if got.Kind != credential.KindSecret {
		t.Fatalf("expected kind %q, got %q", credential.KindSecret, got.Kind)
	}

	if got.Value != "" {
		t.Fatal("expected secret credential to never store the raw value in Value")
	}

	if got.SecretHash == "" {
		t.Fatal("expected secret credential to store a secret hash")
	}

	if got.SecretHash == issued.Secret {
		t.Fatal("expected the stored hash to differ from the raw issued secret")
	}
}

func TestString_RedactsRawSecret(t *testing.T) {
	issued, err := credential.New(
		"project-1",
		"credential-2",
		"server key",
		credential.KindSecret,
	)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	rendered := issued.String()

	if strings.Contains(rendered, issued.Secret) {
		t.Fatalf("expected rendered credential to redact the raw secret, got %q", rendered)
	}

	if !strings.Contains(rendered, "REDACTED") {
		t.Fatalf("expected rendered credential to contain REDACTED, got %q", rendered)
	}
}

func TestNew_RejectsMissingMetadata(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		id             string
		credentialName string
		kind           credential.Kind
	}{
		{
			name:           "missing project ID",
			projectID:      "",
			id:             "credential-1",
			credentialName: "server",
			kind:           credential.KindSecret,
		},
		{
			name:           "missing credential ID",
			projectID:      "project-1",
			id:             "",
			credentialName: "server",
			kind:           credential.KindSecret,
		},
		{
			name:           "missing credential name",
			projectID:      "project-1",
			id:             "credential-1",
			credentialName: "",
			kind:           credential.KindSecret,
		},
		{
			name:           "invalid kind",
			projectID:      "project-1",
			id:             "credential-1",
			credentialName: "server",
			kind:           "UNKNOWN",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := credential.New(
				test.projectID,
				test.id,
				test.credentialName,
				test.kind,
			)

			if !errors.Is(err, credential.ErrInvalidCredential) {
				t.Fatalf("expected ErrInvalidCredential, got %v", err)
			}
		})
	}
}

func TestMatches_AcceptsCorrectCandidateWhileActive(t *testing.T) {
	tests := []struct {
		name string
		kind credential.Kind
	}{
		{name: "public credential", kind: credential.KindPublic},
		{name: "secret credential", kind: credential.KindSecret},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			issued, err := credential.New("project-1", "credential-1", "server", test.kind)
			if err != nil {
				t.Fatalf("expected no error, got %v", err)
			}

			if !issued.Credential.Matches(issued.Secret) {
				t.Fatal("expected credential to match its own issued secret")
			}

			if issued.Credential.Matches(issued.Secret + "-wrong") {
				t.Fatal("expected credential to reject a wrong candidate")
			}
		})
	}
}

func TestMatches_RejectsWhenNotActive(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	revoked, err := issued.Credential.Revoke()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if revoked.Matches(issued.Secret) {
		t.Fatal("expected a revoked credential to never match, even with the correct secret")
	}
}

func TestRevoke_AllowsFromActive(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := issued.Credential.Revoke()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Status != credential.StatusRevoked {
		t.Fatalf("expected status %q, got %q", credential.StatusRevoked, got.Status)
	}
}

func TestRevoke_RejectsWhenNotActive(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	revoked, err := issued.Credential.Revoke()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = revoked.Revoke()
	if !errors.Is(err, credential.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestWithExpiry_SetsExpiryWhileActive(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expiresAt := time.Now().Add(24 * time.Hour)

	got, err := issued.Credential.WithExpiry(expiresAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ExpiresAt == nil || !got.ExpiresAt.Equal(expiresAt) {
		t.Fatalf("expected expiry %v, got %v", expiresAt, got.ExpiresAt)
	}
}

func TestWithExpiry_RejectsZeroTime(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = issued.Credential.WithExpiry(time.Time{})
	if !errors.Is(err, credential.ErrInvalidCredential) {
		t.Fatalf("expected ErrInvalidCredential, got %v", err)
	}
}

func TestWithExpiry_RejectsWhenNotActive(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	revoked, err := issued.Credential.Revoke()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = revoked.WithExpiry(time.Now().Add(time.Hour))
	if !errors.Is(err, credential.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestIsExpired(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if issued.Credential.IsExpired(time.Now()) {
		t.Fatal("expected credential without an expiry to never report expired")
	}

	expiresAt := time.Now().Add(time.Hour)

	withExpiry, err := issued.Credential.WithExpiry(expiresAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if withExpiry.IsExpired(expiresAt.Add(-time.Minute)) {
		t.Fatal("expected credential to not be expired before its expiry time")
	}

	if !withExpiry.IsExpired(expiresAt) {
		t.Fatal("expected credential to be expired at its exact expiry time")
	}

	if !withExpiry.IsExpired(expiresAt.Add(time.Minute)) {
		t.Fatal("expected credential to be expired after its expiry time")
	}
}

func TestExpire_TransitionsWhenDue(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expiresAt := time.Now().Add(time.Hour)

	withExpiry, err := issued.Credential.WithExpiry(expiresAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := withExpiry.Expire(expiresAt.Add(time.Minute))
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.Status != credential.StatusExpired {
		t.Fatalf("expected status %q, got %q", credential.StatusExpired, got.Status)
	}
}

func TestExpire_RejectsWhenNotYetDue(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	expiresAt := time.Now().Add(time.Hour)

	withExpiry, err := issued.Credential.WithExpiry(expiresAt)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = withExpiry.Expire(expiresAt.Add(-time.Minute))
	if !errors.Is(err, credential.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}

func TestExpire_RejectsWhenNotActive(t *testing.T) {
	issued, err := credential.New("project-1", "credential-1", "server", credential.KindSecret)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	revoked, err := issued.Credential.Revoke()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = revoked.Expire(time.Now())
	if !errors.Is(err, credential.ErrInvalidTransition) {
		t.Fatalf("expected ErrInvalidTransition, got %v", err)
	}
}
