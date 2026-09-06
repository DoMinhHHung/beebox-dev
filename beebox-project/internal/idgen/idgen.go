package idgen

import (
	"fmt"

	"github.com/google/uuid"
)

// New returns a UUIDv7 string (time-ordered).
// Generated in-process because managed Postgres (Supabase) does not expose uuidv7() yet.
func New() (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("idgen: uuidv7: %w", err)
	}
	return id.String(), nil
}
