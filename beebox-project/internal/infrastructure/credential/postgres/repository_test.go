package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	credentialpostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/postgres"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	projectpostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/postgres"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
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

func mustTestProject(t *testing.T, pool *pgxpool.Pool) project.Project {
	t.Helper()
	id, err := idgen.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	p, err := project.NewProject(id, "owner-1", "Test Project", project.TierFree, time.Now().UTC().Truncate(time.Microsecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	repo := projectpostgres.New(pool)
	if err := repo.Save(context.Background(), p); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM projects WHERE id = $1", p.ID)
	})
	return p
}

func cleanupCredential(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM credentials WHERE id = $1", id)
	})
}

func TestRepository_SaveAndFindByID(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	repo := credentialpostgres.New(pool)
	ctx := context.Background()

	c, _, err := credential.NewCredential(proj.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cleanupCredential(t, pool, c.ID)

	if err := repo.Save(ctx, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicKey != c.PublicKey || got.Status != c.Status {
		t.Fatalf("expected %+v, got %+v", c, got)
	}
}

func TestRepository_FindByPublicKey(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	repo := credentialpostgres.New(pool)
	ctx := context.Background()

	c, _, err := credential.NewCredential(proj.ID, credential.EnvironmentLive)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cleanupCredential(t, pool, c.ID)
	if err := repo.Save(ctx, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByPublicKey(ctx, c.PublicKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != c.ID {
		t.Fatalf("expected ID %q, got %q", c.ID, got.ID)
	}
}

func TestRepository_Save_UpsertReflectsRevocation(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	repo := credentialpostgres.New(pool)
	ctx := context.Background()

	c, _, err := credential.NewCredential(proj.ID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cleanupCredential(t, pool, c.ID)
	if err := repo.Save(ctx, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	revoked, err := c.Revoke()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(ctx, revoked); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Status != credential.StatusRevoked || got.PublicKey != c.PublicKey {
		t.Fatalf("expected revoked status with unchanged public key, got %+v", got)
	}
	if got.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	pool := mustTestPool(t)
	repo := credentialpostgres.New(pool)

	_, err := repo.FindByID(context.Background(), "does-not-exist")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}