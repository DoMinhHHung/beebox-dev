package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/ownersession"
	sessionpostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/ownersession/postgres"
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

func mustTestOwnerID(t *testing.T, pool *pgxpool.Pool) string {
	t.Helper()
	ctx := context.Background()
	ownerID := "test-owner-for-session"
	_, err := pool.Exec(ctx, `
		INSERT INTO owners (id, email, password_hash, created_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (id) DO NOTHING
	`, ownerID, ownerID+"@example.com", "irrelevant-hash")
	if err != nil {
		t.Fatalf("unexpected error seeding owner: %v", err)
	}
	return ownerID
}

func cleanupSession(t *testing.T, pool *pgxpool.Pool, id string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM owner_sessions WHERE id = $1", id)
	})
}

func TestRepository_SaveAndFindByTokenHash(t *testing.T) {
	pool := mustTestPool(t)
	ownerID := mustTestOwnerID(t, pool)
	repo := sessionpostgres.New(pool)
	ctx := context.Background()

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	s, _, err := ownersession.NewSession(ownerID, time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cleanupSession(t, pool, s.ID)

	if err := repo.Save(ctx, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByTokenHash(ctx, s.TokenHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.OwnerID != ownerID {
		t.Fatalf("expected owner id %q, got %q", ownerID, got.OwnerID)
	}
	if got.RevokedAt != nil {
		t.Fatal("expected RevokedAt to be nil")
	}
}

func TestRepository_Save_UpdateRevokedAt(t *testing.T) {
	pool := mustTestPool(t)
	ownerID := mustTestOwnerID(t, pool)
	repo := sessionpostgres.New(pool)
	ctx := context.Background()

	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	s, _, err := ownersession.NewSession(ownerID, time.Hour, createdAt)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	cleanupSession(t, pool, s.ID)

	if err := repo.Save(ctx, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	revoked, err := s.Revoke(createdAt.Add(10 * time.Minute))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(ctx, revoked); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByTokenHash(ctx, s.TokenHash)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.RevokedAt == nil {
		t.Fatal("expected RevokedAt to be set")
	}
}

func TestRepository_FindByTokenHash_NotFound(t *testing.T) {
	pool := mustTestPool(t)
	repo := sessionpostgres.New(pool)

	_, err := repo.FindByTokenHash(context.Background(), "does-not-exist-hash")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}
