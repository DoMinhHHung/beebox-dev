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

func TestNew_ReturnsUUIDv7(t *testing.T) {
	id, err := New()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	// Version nibble is the first character of the 3rd group (index 14).
	if id[14] != '7' {
		t.Fatalf("expected UUIDv7 (version nibble 7), got %q in %s", id[14], id)
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
