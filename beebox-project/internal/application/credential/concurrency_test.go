package credential_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	credentialApp "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/application/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/project"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/idgen"
	credentialpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/credential/postgres"
	infrapostgres "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/postgres"
	projectpg "github.com/DoMinhHHung/beebox-dev/beebox-project/internal/infrastructure/project/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

func mustConcurrencyTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL not set, skipping concurrency integration test")
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

func TestService_Rotate_ConcurrentCalls_ExposesLostUpdateWithoutLocking(t *testing.T) {
	pool := mustConcurrencyTestPool(t)
	ctx := context.Background()

	projectRepo := projectpg.New(pool)
	credentialRepo := credentialpg.New(pool)
	svc := credentialApp.NewService(credentialRepo, projectRepo)

	projectID, err := idgen.New()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ownerID := "concurrency-test-owner-" + projectID

	p, err := project.NewProject(projectID, ownerID, "Concurrency Test Project", project.TierFree, time.Now())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := projectRepo.Save(ctx, p); err != nil {
		t.Fatalf("unexpected error saving project: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(cleanupCtx, "DELETE FROM projects WHERE id = $1", projectID)
	})

	md, _, err := svc.Create(ctx, ownerID, projectID, credential.EnvironmentTest)
	if err != nil {
		t.Fatalf("unexpected error creating credential: %v", err)
	}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, _ = pool.Exec(cleanupCtx, "DELETE FROM credentials WHERE id = $1", md.ID)
	})

	const goroutines = 10
	var wg sync.WaitGroup
	secrets := make([]string, goroutines)
	errs := make([]error, goroutines)

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			_, secret, err := svc.Rotate(context.Background(), ownerID, md.ID)
			secrets[idx] = secret
			errs[idx] = err
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("goroutine %d unexpectedly failed with %v (expected all concurrent Rotate calls to succeed, since there is no optimistic locking to reject any of them)", i, err)
		}
	}

	finalCredential, err := credentialRepo.FindByID(ctx, md.ID)
	if err != nil {
		t.Fatalf("unexpected error fetching final credential: %v", err)
	}

	matchCount := 0
	for _, secret := range secrets {
		if secret == "" {
			continue
		}
		if verr := finalCredential.VerifySecret(secret); verr == nil {
			matchCount++
		} else if apperror.CodeOf(verr) != apperror.CodeCredentialInvalid {
			t.Fatalf("unexpected error verifying secret: %v", verr)
		}
	}

	if matchCount == 0 {
		t.Fatal("expected at least one of the concurrently-rotated secrets to match the final stored credential, got zero — Rotate appears broken, not just unlocked")
	}

	t.Logf("known limitation (no optimistic locking on Rotate): %d of %d concurrently-issued secrets are now unusable; only %d still match the final stored hash", goroutines-matchCount, goroutines, matchCount)
}