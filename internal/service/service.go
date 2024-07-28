package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int64
}

type Storage interface {
	CreateUser(ctx context.Context, login string, hashedPassword []byte) error
	SelectUser(ctx context.Context, login string) (*models.User, error)
	Ping(ctx context.Context) error
}

type Service struct {
	storage Storage
	logger  *slog.Logger
	cfg     *config.Config
}

func NewService(storage Storage, logger *slog.Logger, cfg *config.Config) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
		cfg:     cfg,
	}
}

func (s *Service) CreateUser(ctx context.Context, login, password string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.storage.CreateUser(ctx, login, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *Service) AuthenticateUser(ctx context.Context, login, password string) (string, error) {
	user, err := s.storage.SelectUser(ctx, login)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			return "", models.ErrUserInvalidCredentials
		}

		return "", fmt.Errorf("failed to select user: %w", err)
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.HashedPassword), []byte(password))
	if err != nil {
		return "", models.ErrUserInvalidCredentials
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.JWTExp)),
		},
		UserID: user.ID,
	})

	tokenString, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	return tokenString, nil
}

func (s *Service) Ping(ctx context.Context) error {
	err := s.storage.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to reach storage: %w", err)
	}

	return nil
}
