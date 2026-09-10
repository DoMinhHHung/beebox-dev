//go:build integration

package postgres_test

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	domainproject "github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/internal/infrastructure/postgres"
)

func newTestPool(t *testing.T) *postgres.ProjectRepository {
	t.Helper()

	connString := os.Getenv("BEEBOX_PROJECT_TEST_DATABASE_URL")
	if connString == "" {
		t.Skip("BEEBOX_PROJECT_TEST_DATABASE_URL not set, skipping integration test")
	}

	pool, err := postgres.NewPool(context.Background(), connString)
	if err != nil {
		t.Fatalf("failed to connect: %v", err)
	}
	t.Cleanup(pool.Close)

	return postgres.NewProjectRepository(pool)
}

func TestProjectRepository_CreateGetUpdate(t *testing.T) {
	repo := newTestPool(t)
	ctx := context.Background()

	p := domainproject.Project{
		ID:             fmt.Sprintf("integration-project-%d", time.Now().UnixNano()),
		OrganizationID: "org-1",
		Status:         domainproject.StatusDraft,
	}

	if err := repo.Create(ctx, p); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	got, err := repo.Get(ctx, p.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != domainproject.StatusDraft {
		t.Fatalf("expected DRAFT, got %q", got.Status)
	}
}
