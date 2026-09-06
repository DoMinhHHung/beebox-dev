package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	projectpostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mustTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping postgres integration test")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := infrapostgres.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func mustProject(t *testing.T, ownerID, name string) project.Project {
	t.Helper()
	id, err := idgen.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, err := project.NewProject(id, ownerID, name, project.TierFree, time.Now().UTC().Truncate(time.Microsecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return p
}

func cleanupProject(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM projects WHERE id = $1", id)
	})
}

func TestRepository_SaveAndFindByID(t *testing.T) {
	pool := mustTestPool(t)
	repo := projectpostgres.New(pool)
	ctx := context.Background()

	p := mustProject(t, "owner-1", "Test Project")
	cleanupProject(t, pool, p.ID)

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != p.Name || got.OwnerID != p.OwnerID || got.Tier != p.Tier {
		t.Fatalf("expected %+v, got %+v", p, got)
	}
}

func TestRepository_Save_UpsertUpdatesFieldsKeepsCreatedAt(t *testing.T) {
	pool := mustTestPool(t)
	repo := projectpostgres.New(pool)
	ctx := context.Background()

	p := mustProject(t, "owner-1", "Original Name")
	cleanupProject(t, pool, p.ID)

	if err := repo.Save(ctx, p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	updated := p
	updated.Name = "Updated Name"
	updated.OwnerID = "owner-2"
	if err := repo.Save(ctx, updated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(ctx, p.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name != "Updated Name" || got.OwnerID != "owner-2" {
		t.Fatalf("expected updated fields, got %+v", got)
	}
	if !got.CreatedAt.Equal(p.CreatedAt) {
		t.Fatalf("expected CreatedAt unchanged, got %v want %v", got.CreatedAt, p.CreatedAt)
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	pool := mustTestPool(t)
	repo := projectpostgres.New(pool)

	_, err := repo.FindByID(context.Background(), "does-not-exist")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}
