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
	GetUserBalance(ctx context.Context, userID int) (*models.Balance, error)
	CreateOrder(ctx context.Context, userID int, orderNumber string,
		status models.OrderStatus) (*models.Order, bool, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
	GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
	UserWithdrawBonuses(ctx context.Context, userID int, orderNumber string, sum float64) error
	UpdateOrder(ctx context.Context, orderNumber string, status models.OrderStatus, accrual float64) error
	GetPendingOrders(ctx context.Context) ([]models.Order, error)
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
