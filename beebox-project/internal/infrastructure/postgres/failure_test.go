package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
)

func TestNewPool_UnreachableHost_ReturnsErrorNotHang(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	start := time.Now()
	_, err := postgres.NewPool(ctx, "postgresql://user:pass@10.255.255.1:5432/db?sslmode=disable")
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("expected an error connecting to an unreachable host")
	}
	if elapsed > 5*time.Second {
		t.Fatalf("expected NewPool to respect context timeout (~2s), took %v", elapsed)
	}
}

func TestPool_AfterClose_QueryReturnsErrorNotPanic(t *testing.T) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping postgres failure integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	pool, err := postgres.NewPool(ctx, databaseURL)
	if err != nil {
		t.Fatalf("failed to connect to test database: %v", err)
	}
	pool.Close()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("expected no panic after querying a closed pool, got: %v", r)
		}
	}()

	_, err = pool.Exec(context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected an error when querying a closed pool")
	}
}