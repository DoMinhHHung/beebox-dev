package memory

import (
	"context"
	"sync"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
)

func mustCredential(t *testing.T) (credential.Credential, string) {
	t.Helper()
	c, secret, err := credential.NewCredential("proj-1", credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return c, secret
}

func TestRepository_SaveAndFindByID(t *testing.T) {
	repo := New()
	ctx := context.Background()
	c, _ := mustCredential(t)

	if err := repo.Save(ctx, c); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got, err := repo.FindByID(ctx, c.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublicKey != c.PublicKey {
		t.Fatalf("expected PublicKey %q, got %q", c.PublicKey, got.PublicKey)
	}
}

func TestRepository_FindByPublicKey(t *testing.T) {
	repo := New()
	ctx := context.Background()
	c, _ := mustCredential(t)
	_ = repo.Save(ctx, c)

	got, err := repo.FindByPublicKey(ctx, c.PublicKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ID != c.ID {
		t.Fatalf("expected ID %q, got %q", c.ID, got.ID)
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	repo := New()
	_, err := repo.FindByID(context.Background(), "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_FindByPublicKey_NotFound(t *testing.T) {
	repo := New()
	_, err := repo.FindByPublicKey(context.Background(), "pk_test_missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %v", apperror.CodeOf(err))
	}
}

func TestRepository_FindByPublicKey_ReflectsRotatedSecretHash(t *testing.T) {
	repo := New()
	ctx := context.Background()
	c, oldSecret := mustCredential(t)
	_ = repo.Save(ctx, c)

	rotated, newSecret, err := c.Rotate()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := repo.Save(ctx, rotated); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := repo.FindByPublicKey(ctx, c.PublicKey)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := got.VerifySecret(oldSecret); apperror.CodeOf(err) != apperror.CodeCredentialInvalid {
		t.Fatalf("expected old secret invalid after rotate, got %v", apperror.CodeOf(err))
	}
	if err := got.VerifySecret(newSecret); err != nil {
		t.Fatalf("expected new secret valid, got error: %v", err)
	}
}

func TestRepository_Save_ConcurrentDifferentCredentials(t *testing.T) {
	repo := New()
	ctx := context.Background()

	var wg sync.WaitGroup
	errs := make([]error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			c, _, err := credential.NewCredential("proj-1", credential.EnvironmentTest)
			if err != nil {
				errs[idx] = err
				return
			}
			errs[idx] = repo.Save(ctx, c)
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d: unexpected error: %v", i, err)
		}
	}
}