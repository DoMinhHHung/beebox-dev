package credential_test

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
	applicationcredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/application/project"
	domaincredential "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/credential"
	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/memory"
)

func newCredentialService(t *testing.T) (*applicationcredential.Service, *project.Service, *memory.CredentialRepository) {
	t.Helper()
	projectsRepo := memory.NewProjectRepository()
	projects := project.NewService(projectsRepo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatalf("create project: %v", err)
	}
	if _, err := projects.Transition(context.Background(), "project-1", domainproject.StatusActive); err != nil {
		t.Fatalf("activate: %v", err)
	}
	repo := memory.NewCredentialRepository()
	return applicationcredential.NewService(repo, projects), projects, repo
}

func TestIssuePublic_ReturnsRawOnceAndPersistsHashOnly(t *testing.T) {
	svc, _, repo := newCredentialService(t)
	ctx := context.Background()

	issued, err := svc.IssuePublic(ctx, "project-1", "organization-1", "frontend")
	if err != nil {
		t.Fatalf("issue: %v", err)
	}
	if issued.Secret == "" {
		t.Fatal("expected raw credential")
	}
	if issued.Credential.Value != "" {
		t.Fatal("persisted credential must not keep Value")
	}
	if issued.Credential.SecretHash == "" || issued.Credential.SecretHash == issued.Secret {
		t.Fatal("expected non-reversible secret hash")
	}

	stored, err := repo.FindByID(ctx, "project-1", issued.Credential.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.Value != "" {
		t.Fatal("repository must not store raw value")
	}
	if stored.SecretHash != issued.Credential.SecretHash {
		t.Fatal("hash mismatch")
	}
}

func TestIssuePublic_RejectsInactiveAndWrongOrg(t *testing.T) {
	projectsRepo := memory.NewProjectRepository()
	projects := project.NewService(projectsRepo)
	if _, err := projects.Create(context.Background(), "project-1", "organization-1"); err != nil {
		t.Fatal(err)
	}
	svc := applicationcredential.NewService(memory.NewCredentialRepository(), projects)

	_, err := svc.IssuePublic(context.Background(), "project-1", "organization-1", "frontend")
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("draft project expected forbidden, got %v", err)
	}

	if _, err := projects.Transition(context.Background(), "project-1", domainproject.StatusActive); err != nil {
		t.Fatal(err)
	}
	_, err = svc.IssuePublic(context.Background(), "project-1", "organization-2", "frontend")
	if apperror.CodeOf(err) != apperror.CodeForbidden {
		t.Fatalf("wrong org expected forbidden, got %v", err)
	}
}

func TestVerifyPublic_SuccessAndFailures(t *testing.T) {
	svc, projects, repo := newCredentialService(t)
	ctx := context.Background()
	issued, err := svc.IssuePublic(ctx, "project-1", "organization-1", "frontend")
	if err != nil {
		t.Fatal(err)
	}

	verified, err := svc.VerifyPublic(ctx, "project-1", issued.Secret)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if verified.ProjectID != "project-1" || verified.CredentialID != issued.Credential.ID {
		t.Fatalf("unexpected verify result %#v", verified)
	}

	_, err = svc.VerifyPublic(ctx, "project-1", "wrong-credential-value-00000000000000000000000000000000")
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("wrong credential expected unauthenticated, got %v", err)
	}

	if _, err := projects.Create(ctx, "project-2", "organization-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := projects.Transition(ctx, "project-2", domainproject.StatusActive); err != nil {
		t.Fatal(err)
	}
	_, err = svc.VerifyPublic(ctx, "project-2", issued.Secret)
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("wrong project expected unauthenticated, got %v", err)
	}

	revoked, err := svc.Revoke(ctx, "project-1", "organization-1", issued.Credential.ID)
	if err != nil {
		t.Fatal(err)
	}
	if revoked.Status != domaincredential.StatusRevoked {
		t.Fatalf("status %s", revoked.Status)
	}
	_, err = svc.VerifyPublic(ctx, "project-1", issued.Secret)
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("revoked expected unauthenticated, got %v", err)
	}

	// expired
	issued2, err := svc.IssuePublic(ctx, "project-1", "organization-1", "frontend-2")
	if err != nil {
		t.Fatal(err)
	}
	stored, err := repo.FindByID(ctx, "project-1", issued2.Credential.ID)
	if err != nil {
		t.Fatal(err)
	}
	past := time.Now().Add(-time.Hour)
	stored.ExpiresAt = &past
	if err := repo.Update(ctx, stored); err != nil {
		t.Fatal(err)
	}
	_, err = svc.VerifyPublic(ctx, "project-1", issued2.Secret)
	if apperror.CodeOf(err) != apperror.CodeUnauthenticated {
		t.Fatalf("expired expected unauthenticated, got %v", err)
	}
}

func TestIssuePublic_AllowsMultipleActiveCredentials(t *testing.T) {
	svc, _, _ := newCredentialService(t)
	ctx := context.Background()
	first, err := svc.IssuePublic(ctx, "project-1", "organization-1", "a")
	if err != nil {
		t.Fatal(err)
	}
	second, err := svc.IssuePublic(ctx, "project-1", "organization-1", "b")
	if err != nil {
		t.Fatal(err)
	}
	if first.Credential.ID == second.Credential.ID {
		t.Fatal("expected distinct credential ids")
	}
	if _, err := svc.VerifyPublic(ctx, "project-1", first.Secret); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.VerifyPublic(ctx, "project-1", second.Secret); err != nil {
		t.Fatal(err)
	}
}
