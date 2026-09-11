package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/credential"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
)

type CredentialRepository struct {
	pool *pgxpool.Pool
}

func NewCredentialRepository(pool *pgxpool.Pool) *CredentialRepository {
	return &CredentialRepository{pool: pool}
}

func (r *CredentialRepository) Create(ctx context.Context, value credential.Credential) error {
	q := querierFrom(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO credentials (user_id, password_hash, created_at, revoked_at) VALUES ($1, $2, $3, $4)`,
		value.UserID().String(), value.PasswordHash(), value.CreatedAt().UTC(), revokedAtPtr(value),
	)
	return err
}

func (r *CredentialRepository) FindByUserID(ctx context.Context, userID identity.Identifier) (credential.Credential, error) {
	q := querierFrom(ctx, r.pool)
	var (
		uid          string
		passwordHash string
		createdAt    time.Time
		revokedAt    *time.Time
	)
	err := q.QueryRow(ctx,
		`SELECT user_id, password_hash, created_at, revoked_at FROM credentials WHERE user_id = $1`,
		userID.String(),
	).Scan(&uid, &passwordHash, &createdAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return credential.Credential{}, auth.ErrNotFound
	}
	if err != nil {
		return credential.Credential{}, err
	}
	ident, err := identity.NewIdentifier(uid)
	if err != nil {
		return credential.Credential{}, err
	}
	cred, err := credential.NewPassword(ident, passwordHash, createdAt.UTC())
	if err != nil {
		return credential.Credential{}, err
	}
	if revokedAt != nil {
		cred, err = cred.Revoke(revokedAt.UTC())
		if err != nil {
			return credential.Credential{}, err
		}
	}
	return cred, nil
}

func (r *CredentialRepository) Update(ctx context.Context, value credential.Credential) error {
	q := querierFrom(ctx, r.pool)
	tag, err := q.Exec(ctx,
		`UPDATE credentials SET password_hash = $2, revoked_at = $3 WHERE user_id = $1`,
		value.UserID().String(), value.PasswordHash(), revokedAtPtr(value),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return auth.ErrNotFound
	}
	return nil
}

func revokedAtPtr(value credential.Credential) *time.Time {
	if t, ok := value.RevokedAt(); ok {
		utc := t.UTC()
		return &utc
	}
	return nil
}

var _ auth.CredentialRepository = (*CredentialRepository)(nil)
