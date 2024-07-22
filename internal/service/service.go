package service

import (
	"context"
	"fmt"
	"log/slog"
)

type Storage interface {
	CreateUser(ctx context.Context, login, password string) error
	Ping(ctx context.Context) error
}

type Service struct {
	storage Storage
	logger  *slog.Logger
}

func NewService(storage Storage, logger *slog.Logger) *Service {
	return &Service{
		storage: storage,
		logger:  logger,
	}
}

func (s *Service) CreateUser(ctx context.Context, login, password string) error {
	err := s.storage.CreateUser(ctx, login, password)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

func (s *Service) Ping(ctx context.Context) error {
	err := s.storage.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to reach storage: %w", err)
	}

	return nil
}
