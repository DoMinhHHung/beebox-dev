package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/owner"
)

func TestOwnerRepository_SaveAndFind(t *testing.T) {
	repo := NewOwnerRepository()
	o, err := owner.NewOwner("owner-1", "minh@example.com", "password123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error building owner: %v", err)
	}

	if err := repo.Save(context.Background(), o); err != nil {
		t.Fatalf("expected no error saving, got %v", err)
	}

	byID, err := repo.FindByID(context.Background(), "owner-1")
	if err != nil {
		t.Fatalf("expected to find owner by id, got %v", err)
	}
	if byID.Email != "minh@example.com" {
		t.Fatalf("expected email %q, got %q", "minh@example.com", byID.Email)
	}

	byEmail, err := repo.FindByEmail(context.Background(), "minh@example.com")
	if err != nil {
		t.Fatalf("expected to find owner by email, got %v", err)
	}
	if byEmail.ID != "owner-1" {
		t.Fatalf("expected id %q, got %q", "owner-1", byEmail.ID)
	}
}

func TestOwnerRepository_Save_DuplicateEmail(t *testing.T) {
	repo := NewOwnerRepository()
	first, err := owner.NewOwner("owner-1", "minh@example.com", "password123", time.Now())
	if err != nil {
		t.Fatalf("unexpected error building owner: %v", err)
	}
	if err := repo.Save(context.Background(), first); err != nil {
		t.Fatalf("unexpected error saving first owner: %v", err)
	}

	second, err := owner.NewOwner("owner-2", "minh@example.com", "differentpass", time.Now())
	if err != nil {
		t.Fatalf("unexpected error building owner: %v", err)
	}

	err = repo.Save(context.Background(), second)
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
	if apperror.CodeOf(err) != apperror.CodeConflict {
		t.Fatalf("expected CodeConflict, got %s", apperror.CodeOf(err))
	}
}

func TestOwnerRepository_FindByID_NotFound(t *testing.T) {
	repo := NewOwnerRepository()

	_, err := repo.FindByID(context.Background(), "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", apperror.CodeOf(err))
	}
}

func TestOwnerRepository_FindByEmail_NotFound(t *testing.T) {
	repo := NewOwnerRepository()

	_, err := repo.FindByEmail(context.Background(), "missing@example.com")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %s", apperror.CodeOf(err))
	}
}
