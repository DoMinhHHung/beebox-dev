package apperror_test

import (
	"errors"
	"testing"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-runtime/apperror"
)

func TestCodeOf(t *testing.T) {
	err := apperror.New(apperror.CodeNotFound, "missing")
	if apperror.CodeOf(err) != apperror.CodeNotFound {
		t.Fatalf("got %s", apperror.CodeOf(err))
	}
	if apperror.CodeOf(errors.New("x")) != apperror.CodeInternal {
		t.Fatal("expected internal")
	}
}
