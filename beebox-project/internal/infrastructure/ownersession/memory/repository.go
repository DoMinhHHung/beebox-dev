package memory

import (
	"context"
	"sync"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/ownersession"
)

type SessionRepository struct {
	mu          sync.RWMutex
	byTokenHash map[string]ownersession.Session
}

var _ ownersession.Repository = (*SessionRepository)(nil)

// NewSessionRepository creates an empty in-memory session repository.
func NewSessionRepository() *SessionRepository {
	return &SessionRepository{byTokenHash: make(map[string]ownersession.Session)}
}

func (r *SessionRepository) Save(ctx context.Context, s ownersession.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.byTokenHash[s.TokenHash] = s
	return nil
}

func (r *SessionRepository) FindByTokenHash(ctx context.Context, hash string) (ownersession.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	s, ok := r.byTokenHash[hash]
	if !ok {
		return ownersession.Session{}, apperror.New(apperror.CodeNotFound, "session not found")
	}
	return s, nil
}
