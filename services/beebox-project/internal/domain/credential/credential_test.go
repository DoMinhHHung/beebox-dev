package credential_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/credential"
)

func TestNew_CreatesCredentialMetadata(t *testing.T) {
	got, err := credential.New("project-1", "credential-1", "server")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if got.ProjectID != "project-1" {
		t.Fatalf("expected project ID %q, got %q", "project-1", got.ProjectID)
	}

	if got.ID != "credential-1" {
		t.Fatalf("expected credential ID %q, got %q", "credential-1", got.ID)
	}

	if got.Name != "server" {
		t.Fatalf("expected credential name %q, got %q", "server", got.Name)
	}
}

func TestNew_RejectsMissingMetadata(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		id             string
		credentialName string
	}{
		{
			name:           "missing project ID",
			projectID:      "",
			id:             "credential-1",
			credentialName: "server",
		},
		{
			name:           "missing credential ID",
			projectID:      "project-1",
			id:             "",
			credentialName: "server",
		},
		{
			name:           "missing credential name",
			projectID:      "project-1",
			id:             "credential-1",
			credentialName: "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := credential.New(
				test.projectID,
				test.id,
				test.credentialName,
			)

			if !errors.Is(err, credential.ErrInvalidCredential) {
				t.Fatalf("expected ErrInvalidCredential, got %v", err)
			}
		})
	}
}
