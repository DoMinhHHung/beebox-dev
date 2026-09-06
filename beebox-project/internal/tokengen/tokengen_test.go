package tokengen

import "testing"

func TestNew_ReturnsNonEmptyToken(t *testing.T) {
	token, err := New(32)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestNew_ReturnsUniqueValues(t *testing.T) {
	first, err := New(32)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	second, err := New(32)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if first == second {
		t.Fatal("expected two calls to produce different tokens")
	}
}

func TestHash_IsDeterministic(t *testing.T) {
	if Hash("same-input") != Hash("same-input") {
		t.Fatal("expected hashing the same input twice to produce the same result")
	}
}

func TestHash_DifferentInputsDifferentHash(t *testing.T) {
	if Hash("input-a") == Hash("input-b") {
		t.Fatal("expected different inputs to produce different hashes")
	}
}
