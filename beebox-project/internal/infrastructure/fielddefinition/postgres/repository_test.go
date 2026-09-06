package postgres_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	fielddefinitionpostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/fielddefinition/postgres"
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

func cleanupSchemas(t *testing.T, pool *pgxpool.Pool, projectID string) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(ctx, "DELETE FROM field_definition_schemas WHERE project_id = $1", projectID)
	})
}

func mustField(t *testing.T, name string, kind fielddefinition.FieldKind, required bool) fielddefinition.FieldDefinition {
	t.Helper()
	f, err := fielddefinition.NewFieldDefinition(name, kind, required)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return f
}

func TestRepository_SaveAndFindLatest(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	cleanupSchemas(t, pool, proj.ID)
	repo := fielddefinitionpostgres.New(pool)
	ctx := context.Background()

	schema, err := fielddefinition.NewSchema(proj.ID, []fielddefinition.FieldDefinition{
		mustField(t, "email", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := repo.Save(ctx, schema); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindLatest(ctx, proj.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Version != schema.Version || len(got.Fields) != 1 {
		t.Fatalf("expected %+v, got %+v", schema, got)
	}
}

func TestRepository_Save_DuplicateVersionConflict(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	cleanupSchemas(t, pool, proj.ID)
	repo := fielddefinitionpostgres.New(pool)
	ctx := context.Background()

	schema, err := fielddefinition.NewSchema(proj.ID, []fielddefinition.FieldDefinition{
		mustField(t, "email", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(ctx, schema); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	err = repo.Save(ctx, schema)
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_FindLatest_ReturnsHighestVersion(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	cleanupSchemas(t, pool, proj.ID)
	repo := fielddefinitionpostgres.New(pool)
	ctx := context.Background()

	v1, err := fielddefinition.NewSchema(proj.ID, []fielddefinition.FieldDefinition{
		mustField(t, "email", fielddefinition.FieldKindString, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(ctx, v1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	v2, err := v1.NextVersion([]fielddefinition.FieldDefinition{
		mustField(t, "email", fielddefinition.FieldKindString, true),
		mustField(t, "isVerify", fielddefinition.FieldKindBoolean, true),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(ctx, v2); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindLatest(ctx, proj.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Version != v2.Version || len(got.Fields) != 2 {
		t.Fatalf("expected latest version %d with 2 fields, got version %d with %d fields", v2.Version, got.Version, len(got.Fields))
	}
}

func TestRepository_FindVersion_NotFound(t *testing.T) {
	pool := mustTestPool(t)
	proj := mustTestProject(t, pool)
	cleanupSchemas(t, pool, proj.ID)
	repo := fielddefinitionpostgres.New(pool)

	_, err := repo.FindVersion(context.Background(), proj.ID, 99)
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}
