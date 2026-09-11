package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/application/auth"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/identity"
	"github.com/DoMinhHHung/beebox-dev/services/beebox-identity/internal/domain/session"
)

type SessionRepository struct {
	pool *pgxpool.Pool
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{pool: pool}
}

func (r *SessionRepository) Create(ctx context.Context, value session.Session) error {
	q := querierFrom(ctx, r.pool)
	_, err := q.Exec(ctx,
		`INSERT INTO sessions (id, user_id, created_at, expires_at, revoked_at) VALUES ($1, $2, $3, $4, $5)`,
		value.ID(), value.UserID().String(), value.CreatedAt().UTC(), value.ExpiresAt().UTC(), sessionRevokedAt(value),
	)
	return err
}

func (r *SessionRepository) FindByID(ctx context.Context, id string) (session.Session, error) {
	q := querierFrom(ctx, r.pool)
	var (
		sid       string
		uid       string
		createdAt time.Time
		expiresAt time.Time
		revokedAt *time.Time
	)
	err := q.QueryRow(ctx,
		`SELECT id, user_id, created_at, expires_at, revoked_at FROM sessions WHERE id = $1`,
		id,
	).Scan(&sid, &uid, &createdAt, &expiresAt, &revokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return session.Session{}, auth.ErrNotFound
	}
	if err != nil {
		return session.Session{}, err
	}
	ident, err := identity.NewIdentifier(uid)
	if err != nil {
		return session.Session{}, err
	}
	s, err := session.New(sid, ident, createdAt.UTC(), expiresAt.UTC())
	if err != nil {
		return session.Session{}, err
	}
	if revokedAt != nil {
		s, err = s.Revoke(revokedAt.UTC())
		if err != nil {
			return session.Session{}, err
		}
	}
	return s, nil
}

func (r *SessionRepository) Revoke(ctx context.Context, value session.Session) error {
	q := querierFrom(ctx, r.pool)
	revokedAt, ok := value.RevokedAt()
	if !ok {
		return errors.New("session is not revoked")
	}
	tag, err := q.Exec(ctx,
		`UPDATE sessions SET revoked_at = $2 WHERE id = $1`,
		value.ID(), revokedAt.UTC(),
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return auth.ErrNotFound
	}
	return nil
}

func sessionRevokedAt(value session.Session) *time.Time {
	if t, ok := value.RevokedAt(); ok {
		utc := t.UTC()
		return &utc
	}
	return nil
}

var _ auth.SessionRepository = (*SessionRepository)(nil)
