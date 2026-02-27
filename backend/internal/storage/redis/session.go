package redis

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/storage"
	"github.com/redis/go-redis/v9"
)

type SessionStorage struct {
	rdb *redis.Client
}

func (s *SessionStorage) Create(ctx context.Context, session *domain.Session, expiresIn time.Duration) error {
	if expiresIn <= 0 {
		return errors.New("expiration duration must be positive")
	}

	return s.rdb.Set(ctx, sessionKey(session.Token), session.UserID.String(), expiresIn).Err()
}

func (s *SessionStorage) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	val, err := s.rdb.Get(ctx, sessionKey(token)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}

	userID, err := uuid.Parse(val)
	if err != nil {
		return nil, err
	}
	return &domain.Session{
		UserID: userID,
		Token:  token,
	}, nil
}

func (s *SessionStorage) Delete(ctx context.Context, token string) error {
	count, err := s.rdb.Del(ctx, sessionKey(token)).Result()
	if err != nil {
		return err
	}
	if count == 0 {
		return storage.ErrNotFound
	}
	return nil
}

func NewSessionStorage(rdb *redis.Client) *SessionStorage {
	return &SessionStorage{rdb}
}

func sessionKey(token string) string {
	return "session:" + token
}
