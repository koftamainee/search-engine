package storage

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/koftamainee/search-engine/backend/internal/domain"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrDeleted       = errors.New("already deleted")
)

type UserStorage interface {
	Create(ctx context.Context, user *domain.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	List(ctx context.Context) ([]*domain.User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	Ban(ctx context.Context, id uuid.UUID) error
	Unban(ctx context.Context, id uuid.UUID) error
}

type BookmarkStorage interface {
	Create(ctx context.Context, bookmark *domain.Bookmark) error
	GetByUserID(ctx context.Context, userID uuid.UUID) (*[]domain.Bookmark, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type HistoryStorage interface {
	Create(ctx context.Context, query *domain.SearchHistory) error
	GetByUserID(ctx context.Context, userID uuid.UUID, num int, offset int) (*[]domain.SearchHistory, error)
	Delete(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
}

type SessionStorage interface {
	Create(ctx context.Context, session *domain.Session, expiresIn time.Duration) error
	GetByToken(ctx context.Context, token string) (*domain.Session, error)
	Delete(ctx context.Context, token string) error
}
