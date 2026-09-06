package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/ownersession"
)

func TestSessionRepository_SaveAndFindByTokenHash(t *testing.T) {
	repo := NewSessionRepository()
	s, _, err := ownersession.NewSession("owner-1", time.Hour, time.Now())
	if err != nil {
		t.Fatalf("unexpected error building session: %v", err)
	}

	if err := repo.Save(context.Background(), s); err != nil {
		t.Fatalf("expected no error saving, got %v", err)
	}

	found, err := repo.FindByTokenHash(context.Background(), s.TokenHash)
	if err != nil {
		t.Fatalf("expected to find session, got %v", err)
	}
	if found.OwnerID != "owner-1" {
		t.Fatalf("expected owner id %q, got %q", "owner-1", found.OwnerID)
	}
}

func TestSessionRepository_FindByTokenHash_NotFound(t *testing.T) {
	repo := NewSessionRepository()

	_, err := repo.FindByTokenHash(context.Background(), "missing-hash")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", apperror.CodeOf(err))
	}
}
