package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/storage"
)

type SessionStorage struct {
	db *pgxpool.Pool
}

func NewSessionStorage(db *pgxpool.Pool) *SessionStorage {
	return &SessionStorage{db: db}
}

func (s *SessionStorage) Create(ctx context.Context, session *domain.Session) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO sessions (id, user_id, token, created_at, expires_at)
			VALUES ($1, $2, $3, $4)`,
		session.ID, session.UserID, session.Token, session.CreatedAt, session.ExpiresAt,
	)
	return err
}

func (s *SessionStorage) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	session := &domain.Session{}

	err := s.db.QueryRow(ctx,
		`SELECT id, user_id, token, created_at, expires_at
			FROM sessions 
			WHERE token = $1 AND expires_at > NOW()`,
		token,
	).Scan(&session.ID, &session.UserID, &session.Token, &session.CreatedAt, &session.ExpiresAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	return session, nil
}

func (s *SessionStorage) Delete(ctx context.Context, token string) error {
	tag, err := s.db.Exec(ctx,
		`DELETE FROM sessions WHERE token = $1`,
		token,
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return storage.ErrNotFound
	}

	return nil
}
