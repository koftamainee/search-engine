package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/storage"
)

type UserStorage struct {
	db *pgxpool.Pool
}

func NewUserStorage(db *pgxpool.Pool) *UserStorage {
	return &UserStorage{db: db}
}

func (s *UserStorage) Create(ctx context.Context, user *domain.User) error {
	_, err := s.db.Exec(ctx,
		`INSERT INTO users (id, email, password, is_admin, is_banned) 
			VALUES ($1, $2, $3, $4, $5)`,
		user.ID, user.Email, user.Password, user.IsAdmin, user.IsBanned)

	if err != nil {
		if isUniqueViolation(err) {
			return storage.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (s *UserStorage) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	user := &domain.User{}

	err := s.db.QueryRow(ctx,
		`SELECT id, email, password, is_admin, is_banned, created_at, deleted_at
			FROM users
			WHERE id = $1 and deleted_at IS NULL`,
		id,
	).Scan(&user.ID, &user.Email, &user.Password, &user.IsAdmin, &user.IsBanned, &user.CreatedAt, &user.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, storage.ErrDeleted
	}

	return user, err
}

func (s *UserStorage) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	user := &domain.User{}

	err := s.db.QueryRow(ctx,
		`SELECT id, email, password, is_admin, is_banned, created_at, deleted_at
			FROM users
			WHERE email = $1 and deleted_at IS NULL`,
		email,
	).Scan(&user.ID, &user.Email, &user.Password, &user.IsAdmin, &user.IsBanned, &user.CreatedAt, &user.DeletedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, storage.ErrNotFound
		}
		return nil, err
	}
	if user.DeletedAt != nil {
		return nil, storage.ErrDeleted
	}

	return user, err
}

func (s *UserStorage) List(ctx context.Context) ([]*domain.User, error) {
	//TODO implement me
	panic("implement me")
}

func (s *UserStorage) Delete(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s *UserStorage) Ban(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func (s *UserStorage) Unban(ctx context.Context, id uuid.UUID) error {
	//TODO implement me
	panic("implement me")
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "23505")
}
