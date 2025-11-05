package service

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserBanned         = errors.New("user is banned")
	ErrAlreadyExists      = errors.New("email already taken")
)

const sessionDuration = 7 * 24 * time.Hour

type AuthService struct {
	users    storage.UserStorage
	sessions storage.SessionStorage
}

func NewAuthService(users storage.UserStorage, sessions storage.SessionStorage) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
	}
}

func (s *AuthService) Register(ctx context.Context, email string, password string) (*domain.User, error) {

	hashedPass, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:       uuid.New(),
		Email:    email,
		Password: string(hashedPass),
	}

	if err := s.users.Create(ctx, user); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, ErrAlreadyExists
		}
		return nil, err
	}

	return user, nil
}
