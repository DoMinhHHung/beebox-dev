package http

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-project/apperror"
)

func TestWriteError_UsesAppErrorMessageAndCode(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, apperror.New(apperror.CodeNotFound, "project not found"))

	if rec.Code != 404 {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var got errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if got.Error.Code != "NOT_FOUND" {
		t.Fatalf("expected code NOT_FOUND, got %q", got.Error.Code)
	}
	if got.Error.Message != "project not found" {
		t.Fatalf("expected message %q, got %q", "project not found", got.Error.Message)
	}
}

func TestWriteError_HidesRawMessageForNonAppError(t *testing.T) {
	rec := httptest.NewRecorder()

	writeError(rec, errPlain("dial tcp 10.0.0.1:5432: connection refused"))

	if rec.Code != 500 {
		t.Fatalf("expected status 500, got %d", rec.Code)
	}

	var got errorEnvelope
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}

	if got.Error.Message != "internal error" {
		t.Fatalf("expected generic internal error message, got %q", got.Error.Message)
	}
	if got.Error.Code != "INTERNAL" {
		t.Fatalf("expected code INTERNAL, got %q", got.Error.Code)
	}
}

type errPlain string

func (e errPlain) Error() string { return string(e) }
