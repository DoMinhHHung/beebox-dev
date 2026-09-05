package memory

import (
	"context"
	"sync"

	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/apperror"
	"github.com/DoMinhHHung/beebox-dev/beebox-project/internal/domain/owner"
)

type OwnerRepository struct {
	mu      sync.RWMutex
	byID    map[string]owner.Owner
	byEmail map[string]string
}

var _ owner.Repository = (*OwnerRepository)(nil)

func NewOwnerRepository() *OwnerRepository {
	return &OwnerRepository{
		byID:    make(map[string]owner.Owner),
		byEmail: make(map[string]string),
	}
}

func (r *OwnerRepository) Save(ctx context.Context, o owner.Owner) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if existingID, ok := r.byEmail[o.Email]; ok && existingID != o.ID {
		return apperror.New(apperror.CodeConflict, "email already registered")
	}

	r.byID[o.ID] = o
	r.byEmail[o.Email] = o.ID
	return nil
}

func (r *OwnerRepository) FindByID(ctx context.Context, id string) (owner.Owner, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	o, ok := r.byID[id]
	if !ok {
		return owner.Owner{}, apperror.New(apperror.CodeNotFound, "owner not found")
	}
	return o, nil
}

func (r *OwnerRepository) FindByEmail(ctx context.Context, email string) (owner.Owner, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	id, ok := r.byEmail[email]
	if !ok {
		return owner.Owner{}, apperror.New(apperror.CodeNotFound, "owner not found")
	}
	return r.byID[id], nil
}