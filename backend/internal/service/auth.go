package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/koftamainee/search-engine/backend/internal/domain"
	"github.com/koftamainee/search-engine/backend/internal/storage"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials    = errors.New("invalid credentials")
	ErrUserBanned            = errors.New("user is banned")
	ErrAlreadyExists         = errors.New("email already taken")
	ErrFailedToGenerateToken = errors.New("failed to generate token")
	ErrNotFound              = errors.New("not found")
)

const SessionDuration = 7 * 24 * time.Hour

type AuthService struct {
	users    storage.UserStorage
	sessions storage.SessionStorage
	IsProd   bool
}

func NewAuthService(users storage.UserStorage, sessions storage.SessionStorage, isProd bool) *AuthService {
	return &AuthService{
		users:    users,
		sessions: sessions,
		IsProd:   isProd,
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

func (s *AuthService) Login(ctx context.Context, email string, password string) (*domain.Session, error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.IsBanned {
		return nil, ErrUserBanned
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	token, err := generateToken()
	if err != nil {
		return nil, ErrFailedToGenerateToken
	}
	session := &domain.Session{
		UserID: user.ID,
		Token:  token,
	}

	err = s.sessions.Create(ctx, session, SessionDuration)
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (s *AuthService) Logout(ctx context.Context, token string) error {
	err := s.sessions.Delete(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *AuthService) ValidateToken(ctx context.Context, token string) (*domain.User, error) {
	session, err := s.sessions.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	user, err := s.users.GetByID(ctx, session.UserID)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	if user.IsBanned {
		return nil, ErrUserBanned
	}
	if user.DeletedAt != nil {
		return nil, ErrInvalidCredentials
	}

	return user, nil
}

func generateToken() (string, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
