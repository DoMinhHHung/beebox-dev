package project

import (
	"testing"
	"time"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
)

func TestNewProject_Valid(t *testing.T) {
	p, err := NewProject("id-1", "owner-1", "My Project", TierFree, time.Now())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p.ID != "id-1" || p.OwnerID != "owner-1" || p.Name != "My Project" || p.Tier != TierFree {
		t.Fatalf("unexpected project fields: %+v", p)
	}
}

func TestNewProject_EmptyName(t *testing.T) {
	_, err := NewProject("id-1", "owner-1", "", TierFree, time.Now())
	assertInvalidInput(t, err)
}

func TestNewProject_EmptyOwnerID(t *testing.T) {
	_, err := NewProject("id-1", "", "My Project", TierFree, time.Now())
	assertInvalidInput(t, err)
}

func TestNewProject_NameTooLong(t *testing.T) {
	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	_, err := NewProject("id-1", "owner-1", string(longName), TierFree, time.Now())
	assertInvalidInput(t, err)
}

func TestNewProject_InvalidTier(t *testing.T) {
	_, err := NewProject("id-1", "owner-1", "My Project", Tier("unknown"), time.Now())
	assertInvalidInput(t, err)
}

func assertInvalidInput(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if apperror.CodeOf(err) != apperror.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %s", apperror.CodeOf(err))
	}
}
