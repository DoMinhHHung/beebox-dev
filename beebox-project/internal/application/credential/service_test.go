package credential_test

import (
	"context"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	credentialApp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/memory"
)

func TestService_Create(t *testing.T) {
	repo := memory.New()
	svc := credentialApp.NewService(repo)
	ctx := context.Background()

	md, secret, err := svc.Create(ctx, "proj-1", credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if md.Status != credential.StatusActive {
		t.Fatalf("expected StatusActive, got %s", md.Status)
	}
	if secret == "" {
		t.Fatal("expected non-empty plaintext secret on create")
	}
}

func TestService_Rotate_InvalidatesOldSecret(t *testing.T) {
	repo := memory.New()
	svc := credentialApp.NewService(repo)
	ctx := context.Background()

	md, oldSecret, err := svc.Create(ctx, "proj-1", credential.EnvironmentLive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rotatedMd, newSecret, err := svc.Rotate(ctx, md.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rotatedMd.PublicKey != md.PublicKey {
		t.Fatal("expected PublicKey unchanged after rotate")
	}
	if newSecret == oldSecret {
		t.Fatal("expected a new plaintext secret after rotate")
	}

	stored, err := repo.FindByID(ctx, md.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := stored.VerifySecret(oldSecret); apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected old secret invalid, got %v", apperror.CodeOf(err))
	}
	if err := stored.VerifySecret(newSecret); err != nil {
		t.Fatalf("expected new secret valid, got error: %v", err)
	}
}

func TestService_Revoke_ThenRotateFails(t *testing.T) {
	repo := memory.New()
	svc := credentialApp.NewService(repo)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "proj-1", credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	revokedMd, err := svc.Revoke(ctx, md.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokedMd.Status != credential.StatusRevoked {
		t.Fatalf("expected StatusRevoked, got %s", revokedMd.Status)
	}

	_, _, err = svc.Rotate(ctx, md.ID)
	if apperror.CodeOf(err) != apperror.CodeCredentialRevoked {
		t.Fatalf("expected CodeCredentialRevoked, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetMetadata_NotFound(t *testing.T) {
	repo := memory.New()
	svc := credentialApp.NewService(repo)

	_, err := svc.GetMetadata(context.Background(), "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetMetadata_ReturnsMetadataType(t *testing.T) {
	repo := memory.New()
	svc := credentialApp.NewService(repo)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "proj-1", credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.GetMetadata(ctx, md.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicKey != md.PublicKey {
		t.Fatalf("expected PublicKey %q, got %q", md.PublicKey, got.PublicKey)
	}
}