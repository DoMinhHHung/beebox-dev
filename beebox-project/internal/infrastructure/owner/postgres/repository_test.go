package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/owner"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	ownerpostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/owner/postgres"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
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

func mustOwner(t *testing.T, email string) owner.Owner {
	t.Helper()
	id, err := idgen.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	o, err := owner.NewOwner(id, email, "password123", time.Now().UTC().Truncate(time.Microsecond))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return o
}

func cleanupOwner(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM owners WHERE id = $1", id)
	})
}

func TestRepository_SaveAndFindByID(t *testing.T) {
	pool := mustTestPool(t)
	repo := ownerpostgres.New(pool)
	ctx := context.Background()

	o := mustOwner(t, "minh-8d-1@example.com")
	cleanupOwner(t, pool, o.ID)

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByID(ctx, o.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Email != o.Email {
		t.Fatalf("expected email %q, got %q", o.Email, got.Email)
	}
}

func TestRepository_FindByEmail(t *testing.T) {
	pool := mustTestPool(t)
	repo := ownerpostgres.New(pool)
	ctx := context.Background()

	o := mustOwner(t, "minh-8d-2@example.com")
	cleanupOwner(t, pool, o.ID)

	if err := repo.Save(ctx, o); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByEmail(ctx, o.Email)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != o.ID {
		t.Fatalf("expected id %q, got %q", o.ID, got.ID)
	}
}

func TestRepository_Save_DuplicateEmailConflict(t *testing.T) {
	pool := mustTestPool(t)
	repo := ownerpostgres.New(pool)
	ctx := context.Background()

	first := mustOwner(t, "minh-8d-3@example.com")
	cleanupOwner(t, pool, first.ID)
	if err := repo.Save(ctx, first); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	second := mustOwner(t, "minh-8d-3@example.com")
	cleanupOwner(t, pool, second.ID)

	err := repo.Save(ctx, second)
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	pool := mustTestPool(t)
	repo := ownerpostgres.New(pool)

	_, err := repo.FindByID(context.Background(), "does-not-exist")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}
