package service

import (
	"context"
	"fmt"
	"log/slog"
)

type Storage interface {
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

func (s *Service) Ping(ctx context.Context) error {
	err := s.storage.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to reach storage: %w", err)
	}

	return nil
}
