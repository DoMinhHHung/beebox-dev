package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is required")
		os.Exit(1)
	}

	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	if migrationsDir == "" {
		migrationsDir = "migrations"
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read migrations dir: %v\n", err)
		os.Exit(1)
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasSuffix(name, ".sql") {
			files = append(files, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(files)
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "no migration files found")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect: %v\n", err)
		os.Exit(1)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, `
CREATE TABLE IF NOT EXISTS schema_migrations (
    filename TEXT PRIMARY KEY,
    applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)`); err != nil {
		fmt.Fprintf(os.Stderr, "ensure schema_migrations: %v\n", err)
		os.Exit(1)
	}

	for _, file := range files {
		base := filepath.Base(file)
		var existing string
		err := pool.QueryRow(ctx, `SELECT filename FROM schema_migrations WHERE filename = $1`, base).Scan(&existing)
		if err == nil {
			fmt.Printf("skip %s (already applied)\n", base)
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read %s: %v\n", file, err)
			os.Exit(1)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "begin: %v\n", err)
			os.Exit(1)
		}

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback(ctx)
			fmt.Fprintf(os.Stderr, "apply %s: %v\n", base, err)
			os.Exit(1)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO schema_migrations (filename) VALUES ($1)`, base); err != nil {
			_ = tx.Rollback(ctx)
			fmt.Fprintf(os.Stderr, "record %s: %v\n", base, err)
			os.Exit(1)
		}
		if err := tx.Commit(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "commit %s: %v\n", base, err)
			os.Exit(1)
		}
		fmt.Printf("applied %s\n", base)
	}
}
