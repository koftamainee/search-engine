package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/storage"
)

type HistoryStorage struct {
	db *pgxpool.Pool
}

func NewHistoryStorage(db *pgxpool.Pool) *HistoryStorage {
	return &HistoryStorage{db: db}
}

func (h *HistoryStorage) Create(ctx context.Context, history *domain.SearchHistory) error {
	_, err := h.db.Exec(ctx,
		`INSERT INTO search_history (id, user_id, query, created_at)
			VALUES ($1, $2, $3, $4)`,
		history.ID, history.UserID, history.Query, history.CreatedAt)

	if err != nil {
		return err
	}

	return nil
}

func (h *HistoryStorage) GetByUserID(ctx context.Context, userID uuid.UUID, num int, offset int) (*[]domain.SearchHistory, error) {
	rows, err := h.db.Query(ctx,
		`SELECT id, user_id, query, created_at
			FROM search_history
			WHERE user_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`,
		userID, num, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	history := make([]domain.SearchHistory, 0)

	for rows.Next() {
		var h domain.SearchHistory
		err := rows.Scan(&h.ID, &h.UserID, &h.Query, &h.CreatedAt)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return nil, storage.ErrNotFound
			}
			return nil, err
		}
		history = append(history, h)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &history, nil
}

func (h *HistoryStorage) Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	ct, err := h.db.Exec(ctx,
		`DELETE FROM search_history
			WHERE id = $1 AND user_id = $2`,
		id, userID)

	if err != nil {
		return err
	}

	if ct.RowsAffected() == 0 {
		return storage.ErrNotFound
	}

	return nil
}
