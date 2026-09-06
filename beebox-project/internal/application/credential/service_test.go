package credential_test

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	credentialApp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/memory"
	projectmemory "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/memory"
)

func mustCredentialTestProject(t *testing.T, repo *projectmemory.ProjectRepository, ownerID string) project.Project {
	t.Helper()
	p, err := project.NewProject("proj-1", ownerID, "Test Project", project.TierFree, time.Now())
	if err != nil {
		t.Fatalf("unexpected error building project: %v", err)
	}
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("unexpected error saving project: %v", err)
	}
	return p
}

func TestService_Create(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, secret, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentTest)
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

func TestService_Create_WrongOwnerDenied(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	_, _, err := svc.Create(ctx, "owner-2", p.ID, credential.EnvironmentTest)
	if apperror.CodeOf(err) != apperror.CodeTenantAccessDenied {
		t.Fatalf("expected CodeTenantAccessDenied, got %v", apperror.CodeOf(err))
	}
}

func TestService_Rotate_InvalidatesOldSecret(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, oldSecret, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentLive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	rotatedMd, newSecret, err := svc.Rotate(ctx, "owner-1", md.ID)
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

func TestService_Rotate_WrongOwnerDenied(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, _, err = svc.Rotate(ctx, "owner-2", md.ID)
	if apperror.CodeOf(err) != apperror.CodeTenantAccessDenied {
		t.Fatalf("expected CodeTenantAccessDenied, got %v", apperror.CodeOf(err))
	}
}

func TestService_Revoke_ThenRotateFails(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	revokedMd, err := svc.Revoke(ctx, "owner-1", md.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokedMd.Status != credential.StatusRevoked {
		t.Fatalf("expected StatusRevoked, got %s", revokedMd.Status)
	}

	_, _, err = svc.Rotate(ctx, "owner-1", md.ID)
	if apperror.CodeOf(err) != apperror.CodeCredentialRevoked {
		t.Fatalf("expected CodeCredentialRevoked, got %v", apperror.CodeOf(err))
	}
}

func TestService_Revoke_WrongOwnerDenied(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.Revoke(ctx, "owner-2", md.ID)
	if apperror.CodeOf(err) != apperror.CodeTenantAccessDenied {
		t.Fatalf("expected CodeTenantAccessDenied, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetMetadata_NotFound(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	svc := credentialApp.NewService(repo, projects)

	_, err := svc.GetMetadata(context.Background(), "owner-1", "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestService_GetMetadata_ReturnsMetadataType(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := svc.GetMetadata(ctx, "owner-1", md.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicKey != md.PublicKey {
		t.Fatalf("expected PublicKey %q, got %q", md.PublicKey, got.PublicKey)
	}
}

func TestService_GetMetadata_WrongOwnerDenied(t *testing.T) {
	repo := memory.New()
	projects := projectmemory.NewProjectRepository()
	p := mustCredentialTestProject(t, projects, "owner-1")
	svc := credentialApp.NewService(repo, projects)
	ctx := context.Background()

	md, _, err := svc.Create(ctx, "owner-1", p.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	_, err = svc.GetMetadata(ctx, "owner-2", md.ID)
	if apperror.CodeOf(err) != apperror.CodeTenantAccessDenied {
		t.Fatalf("expected CodeTenantAccessDenied, got %v", apperror.CodeOf(err))
	}
}
