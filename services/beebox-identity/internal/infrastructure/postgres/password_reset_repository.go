package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/passwordreset"
)

type PasswordResetRepository struct {
	pool *pgxpool.Pool
}

func NewPasswordResetRepository(pool *pgxpool.Pool) *PasswordResetRepository {
	return &PasswordResetRepository{pool: pool}
}

func (r *PasswordResetRepository) Create(ctx context.Context, value passwordreset.PasswordReset) error {
	q := querierFrom(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO password_resets (id, user_id, token_hash, created_at, expires_at, used_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		value.ID(), value.UserID().String(), value.TokenHash(),
		value.CreatedAt().UTC(), value.ExpiresAt().UTC(), passwordResetUsedAt(value),
	)
	return err
}

func (r *PasswordResetRepository) FindByID(ctx context.Context, id string) (passwordreset.PasswordReset, error) {
	q := querierFrom(ctx, r.pool)
	var (
		rid       string
		uid       string
		tokenHash string
		createdAt time.Time
		expiresAt time.Time
		usedAt    *time.Time
	)
	err := q.QueryRow(ctx,
		`SELECT id, user_id, token_hash, created_at, expires_at, used_at FROM password_resets WHERE id = $1`,
		id,
	).Scan(&rid, &uid, &tokenHash, &createdAt, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return passwordreset.PasswordReset{}, auth.ErrNotFound
	}
	if err != nil {
		return passwordreset.PasswordReset{}, err
	}
	return mapPasswordReset(rid, uid, tokenHash, createdAt, expiresAt, usedAt)
}

func (r *PasswordResetRepository) FindPending(ctx context.Context, userID identity.Identifier) (passwordreset.PasswordReset, error) {
	q := querierFrom(ctx, r.pool)
	var (
		rid       string
		uid       string
		tokenHash string
		createdAt time.Time
		expiresAt time.Time
		usedAt    *time.Time
	)
	err := q.QueryRow(ctx,
		`SELECT id, user_id, token_hash, created_at, expires_at, used_at
		 FROM password_resets
		 WHERE user_id = $1 AND used_at IS NULL AND expires_at > NOW()
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID.String(),
	).Scan(&rid, &uid, &tokenHash, &createdAt, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return passwordreset.PasswordReset{}, auth.ErrNotFound
	}
	if err != nil {
		return passwordreset.PasswordReset{}, err
	}
	return mapPasswordReset(rid, uid, tokenHash, createdAt, expiresAt, usedAt)
}

func (r *PasswordResetRepository) MarkUsed(ctx context.Context, value passwordreset.PasswordReset) error {
	q := querierFrom(ctx, r.pool)
	usedAt, ok := value.UsedAt()
	if !ok {
		return errors.New("password reset is not used")
	}
	tag, err := q.Exec(ctx,
		`UPDATE password_resets SET used_at = $2 WHERE id = $1 AND used_at IS NULL`,
		value.ID(), usedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return auth.ErrNotFound
	}
	return nil
}

func passwordResetUsedAt(value passwordreset.PasswordReset) *time.Time {
	if t, ok := value.UsedAt(); ok {
		utc := t.UTC()
		return &utc
	}
	return nil
}

func mapPasswordReset(id, uid, tokenHash string, createdAt, expiresAt time.Time, usedAt *time.Time) (passwordreset.PasswordReset, error) {
	ident, err := identity.NewIdentifier(uid)
	if err != nil {
		return passwordreset.PasswordReset{}, err
	}
	reset, err := passwordreset.New(id, ident, tokenHash, createdAt.UTC(), expiresAt.UTC())
	if err != nil {
		return passwordreset.PasswordReset{}, err
	}
	if usedAt != nil {
		reset, err = reset.Consume(usedAt.UTC())
		if err != nil {
			return passwordreset.PasswordReset{}, err
		}
	}
	return reset, nil
}

var _ auth.PasswordResetRepository = (*PasswordResetRepository)(nil)
