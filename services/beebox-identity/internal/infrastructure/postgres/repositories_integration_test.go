//go:build integration

package postgres_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/user"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/infrastructure/postgres"
)

func testPool(t *testing.T) *postgres.UserRepository {
	t.Helper()
	conn := os.Getenv("BEEBOX_IDENTITY_TEST_DATABASE_URL")
	if conn == "" {
		t.Skip("BEEBOX_IDENTITY_TEST_DATABASE_URL not set")
	}
	pool, err := postgres.NewPool(context.Background(), conn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return postgres.NewUserRepository(pool)
}

func openPool(t *testing.T) interface {
	Close()
} {
	t.Helper()
	conn := os.Getenv("BEEBOX_IDENTITY_TEST_DATABASE_URL")
	if conn == "" {
		t.Skip("BEEBOX_IDENTITY_TEST_DATABASE_URL not set")
	}
	pool, err := postgres.NewPool(context.Background(), conn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func TestUserAndCredentialRepositories(t *testing.T) {
	conn := os.Getenv("BEEBOX_IDENTITY_TEST_DATABASE_URL")
	if conn == "" {
		t.Skip("BEEBOX_IDENTITY_TEST_DATABASE_URL not set")
	}
	pool, err := postgres.NewPool(context.Background(), conn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	users := postgres.NewUserRepository(pool)
	credentials := postgres.NewCredentialRepository(pool)
	tx := postgres.NewTransactor(pool)
	ctx := context.Background()
	now := time.Now().UTC()
	id, _ := identity.NewIdentifier("integration-user-" + now.Format("150405.000000000"))
	u, _ := user.New(id, now)
	c, _ := credential.NewPassword(id, "hash-value", now)

	if err := tx.WithinTransaction(ctx, func(ctx context.Context) error {
		if err := users.Create(ctx, u); err != nil {
			return err
		}
		return credentials.Create(ctx, c)
	}); err != nil {
		t.Fatalf("create: %v", err)
	}

	foundUser, err := users.FindByIdentifier(ctx, id)
	if err != nil {
		t.Fatalf("find user: %v", err)
	}
	if foundUser.ID() != id {
		t.Fatal("unexpected user")
	}

	foundCred, err := credentials.FindByUserID(ctx, id)
	if err != nil {
		t.Fatalf("find cred: %v", err)
	}
	if foundCred.PasswordHash() != "hash-value" {
		t.Fatal("unexpected hash")
	}

	updated, err := foundCred.ChangePassword("new-hash")
	if err != nil {
		t.Fatalf("change: %v", err)
	}
	if err := credentials.Update(ctx, updated); err != nil {
		t.Fatalf("update: %v", err)
	}

	_, err = users.FindByIdentifier(ctx, identity.Identifier("missing-user"))
	if err != auth.ErrNotFound {
		t.Fatalf("expected not found, got %v", err)
	}
}

func TestSessionRepository(t *testing.T) {
	conn := os.Getenv("BEEBOX_IDENTITY_TEST_DATABASE_URL")
	if conn == "" {
		t.Skip("BEEBOX_IDENTITY_TEST_DATABASE_URL not set")
	}
	pool, err := postgres.NewPool(context.Background(), conn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	users := postgres.NewUserRepository(pool)
	sessions := postgres.NewSessionRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()
	id, _ := identity.NewIdentifier("session-user-" + now.Format("150405.000000000"))
	u, _ := user.New(id, now)
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("user: %v", err)
	}

	token := "opaque-token-for-session-integration-test-0001"
	storageID := hashToken(token)
	s, _ := session.New(storageID, id, now, now.Add(time.Hour))
	if err := sessions.Create(ctx, s); err != nil {
		t.Fatalf("create session: %v", err)
	}
	found, err := sessions.FindByID(ctx, storageID)
	if err != nil {
		t.Fatalf("find: %v", err)
	}
	if !found.IsActive(now) {
		t.Fatal("expected active")
	}
	revoked, err := found.Revoke(now.Add(time.Minute))
	if err != nil {
		t.Fatalf("revoke domain: %v", err)
	}
	if err := sessions.Revoke(ctx, revoked); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	found, err = sessions.FindByID(ctx, storageID)
	if err != nil {
		t.Fatalf("find revoked: %v", err)
	}
	if found.IsActive(now.Add(2 * time.Minute)) {
		t.Fatal("expected inactive")
	}
}

func TestVerificationAndPasswordResetRepositories(t *testing.T) {
	conn := os.Getenv("BEEBOX_IDENTITY_TEST_DATABASE_URL")
	if conn == "" {
		t.Skip("BEEBOX_IDENTITY_TEST_DATABASE_URL not set")
	}
	pool, err := postgres.NewPool(context.Background(), conn)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}
	t.Cleanup(pool.Close)

	users := postgres.NewUserRepository(pool)
	verifications := postgres.NewVerificationRepository(pool)
	resets := postgres.NewPasswordResetRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()
	id, _ := identity.NewIdentifier("verify-user-" + now.Format("150405.000000000"))
	u, _ := user.New(id, now)
	if err := users.Create(ctx, u); err != nil {
		t.Fatalf("user: %v", err)
	}

	v, _ := verification.New("ver-1", id, verification.TypeEmail, "a@b.com", "code-hash", now, now.Add(15*time.Minute))
	if err := verifications.Create(ctx, v); err != nil {
		t.Fatalf("create verification: %v", err)
	}
	pending, err := verifications.FindPending(ctx, id, verification.TypeEmail, "a@b.com")
	if err != nil {
		t.Fatalf("find pending: %v", err)
	}
	used, err := pending.Consume(now.Add(time.Minute))
	if err != nil {
		t.Fatalf("consume: %v", err)
	}
	if err := verifications.MarkUsed(ctx, used); err != nil {
		t.Fatalf("mark used: %v", err)
	}
	_, err = verifications.FindPending(ctx, id, verification.TypeEmail, "a@b.com")
	if err != auth.ErrNotFound {
		t.Fatalf("expected not found after use, got %v", err)
	}

	reset, _ := passwordreset.New("reset-1", id, "token-hash", now, now.Add(time.Hour))
	if err := resets.Create(ctx, reset); err != nil {
		t.Fatalf("create reset: %v", err)
	}
	found, err := resets.FindByID(ctx, "reset-1")
	if err != nil {
		t.Fatalf("find by id: %v", err)
	}
	pendingReset, err := resets.FindPending(ctx, id)
	if err != nil {
		t.Fatalf("find pending reset: %v", err)
	}
	if pendingReset.ID() != found.ID() {
		t.Fatal("unexpected pending reset")
	}
	consumed, err := found.Consume(now.Add(time.Minute))
	if err != nil {
		t.Fatalf("consume reset: %v", err)
	}
	if err := resets.MarkUsed(ctx, consumed); err != nil {
		t.Fatalf("mark used reset: %v", err)
	}
	_, err = resets.FindPending(ctx, id)
	if err != auth.ErrNotFound {
		t.Fatalf("expected not found pending reset, got %v", err)
	}
}
