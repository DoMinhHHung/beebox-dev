package idgen

import "testing"

func TestNew_ReturnsWellFormedID(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(id) != 36 {
		t.Fatalf("expected 36-char UUID, got %d chars: %s", len(id), id)
	}
}

func TestNew_ReturnsUniqueValues(t *testing.T) {
	first, err := New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	second, err := New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if first == second {
		t.Fatal("expected two calls to produce different IDs")
	}
}