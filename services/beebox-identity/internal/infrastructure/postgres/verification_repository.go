package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/verification"
)

type VerificationRepository struct {
	pool *pgxpool.Pool
}

func NewVerificationRepository(pool *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{pool: pool}
}

func (r *VerificationRepository) Create(ctx context.Context, value verification.Verification) error {
	q := querierFrom(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO verifications (id, user_id, type, target, code_hash, created_at, expires_at, used_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		value.ID(), value.UserID().String(), string(value.Type()), value.Target(), value.CodeHash(),
		value.CreatedAt().UTC(), value.ExpiresAt().UTC(), verificationUsedAt(value),
	)
	return err
}

func (r *VerificationRepository) FindPending(ctx context.Context, userID identity.Identifier, vtype verification.Type, target string) (verification.Verification, error) {
	q := querierFrom(ctx, r.pool)
	var (
		id        string
		uid       string
		typ       string
		tgt       string
		codeHash  string
		createdAt time.Time
		expiresAt time.Time
		usedAt    *time.Time
	)
	err := q.QueryRow(ctx,
		`SELECT id, user_id, type, target, code_hash, created_at, expires_at, used_at
		 FROM verifications
		 WHERE user_id = $1 AND type = $2 AND target = $3 AND used_at IS NULL AND expires_at > NOW()
		 ORDER BY created_at DESC
		 LIMIT 1`,
		userID.String(), string(vtype), target,
	).Scan(&id, &uid, &typ, &tgt, &codeHash, &createdAt, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return verification.Verification{}, auth.ErrNotFound
	}
	if err != nil {
		return verification.Verification{}, err
	}
	return mapVerification(id, uid, typ, tgt, codeHash, createdAt, expiresAt, usedAt)
}

func (r *VerificationRepository) MarkUsed(ctx context.Context, value verification.Verification) error {
	q := querierFrom(ctx, r.pool)
	usedAt, ok := value.UsedAt()
	if !ok {
		return errors.New("verification is not used")
	}
	tag, err := q.Exec(ctx,
		`UPDATE verifications SET used_at = $2 WHERE id = $1`,
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

func verificationUsedAt(value verification.Verification) *time.Time {
	if t, ok := value.UsedAt(); ok {
		utc := t.UTC()
		return &utc
	}
	return nil
}

func mapVerification(id, uid, typ, tgt, codeHash string, createdAt, expiresAt time.Time, usedAt *time.Time) (verification.Verification, error) {
	ident, err := identity.NewIdentifier(uid)
	if err != nil {
		return verification.Verification{}, err
	}
	v, err := verification.New(id, ident, verification.Type(typ), tgt, codeHash, createdAt.UTC(), expiresAt.UTC())
	if err != nil {
		return verification.Verification{}, err
	}
	if usedAt != nil {
		v, err = v.Consume(usedAt.UTC())
		if err != nil {
			return verification.Verification{}, err
		}
	}
	return v, nil
}

var _ auth.VerificationRepository = (*VerificationRepository)(nil)
