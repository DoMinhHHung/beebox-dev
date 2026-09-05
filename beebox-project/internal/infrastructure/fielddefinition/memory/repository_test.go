package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/fielddefinition"
)

func mustSchema(t *testing.T, projectID string, version int) fielddefinition.Schema {
	t.Helper()
	f, err := fielddefinition.NewFieldDefinition("email", fielddefinition.FieldKindString, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return fielddefinition.Schema{ProjectID: projectID, Version: version, Fields: []fielddefinition.FieldDefinition{f}}
}

func TestRepository_SaveAndFindLatest(t *testing.T) {
	repo := New()
	ctx := context.Background()
	s := mustSchema(t, "proj-1", 1)

	if err := repo.Save(ctx, s); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindLatest(ctx, "proj-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Version != 1 {
		t.Fatalf("expected version 1, got %d", got.Version)
	}
}

func TestRepository_FindLatest_NotFound(t *testing.T) {
	repo := New()
	_, err := repo.FindLatest(context.Background(), "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_FindVersion_NotFound(t *testing.T) {
	repo := New()
	ctx := context.Background()
	_ = repo.Save(ctx, mustSchema(t, "proj-1", 1))

	_, err := repo.FindVersion(ctx, "proj-1", 2)
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_Save_DuplicateVersion(t *testing.T) {
	repo := New()
	ctx := context.Background()
	s := mustSchema(t, "proj-1", 1)
	_ = repo.Save(ctx, s)

	err := repo.Save(ctx, s)
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_Save_LowerThanLatest(t *testing.T) {
	repo := New()
	ctx := context.Background()
	_ = repo.Save(ctx, mustSchema(t, "proj-1", 2))

	err := repo.Save(ctx, mustSchema(t, "proj-1", 1))
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_Save_ConcurrentSameVersion(t *testing.T) {
	repo := New()
	ctx := context.Background()
	s := mustSchema(t, "proj-1", 1)

	var wg sync.WaitGroup
	errs := make([]error, 10)
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			errs[idx] = repo.Save(ctx, s)
		}(i)
	}
	wg.Wait()

	successCount := 0
	for _, err := range errs {
		if err == nil {
			successCount++
		}
	}
	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful save, got %d", successCount)
	}
}