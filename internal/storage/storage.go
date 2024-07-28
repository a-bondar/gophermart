package storage

import (
	"context"
	"fmt"

	"github.com/a-bondar/gophermart/internal/models"

	"github.com/a-bondar/gophermart/internal/storage/postgres"
)

type Storage interface {
	CreateUser(ctx context.Context, login string, hashedPassword []byte) error
	SelectUser(ctx context.Context, login string) (*models.User, error)
	Ping(ctx context.Context) error
	Close()
}

func NewStorage(ctx context.Context, dsn string) (Storage, error) {
	storage, err := postgres.NewStorage(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize storage: %w", err)
	}

	return storage, nil
}
