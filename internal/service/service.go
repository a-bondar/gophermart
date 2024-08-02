package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/models"

	"golang.org/x/crypto/bcrypt"
)

type Storage interface {
	CreateUser(ctx context.Context, login string, hashedPassword []byte) error
	SelectUser(ctx context.Context, login string) (*models.User, error)
	GetUserBalance(ctx context.Context, userID int) (float64, error)
	CreateOrder(ctx context.Context, userID int, orderNumber string,
		status models.OrderStatus) (*models.Order, bool, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.Order, error)
	GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error)
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

func validateOrderNumber(orderNumber string) error {
	number, err := strconv.Atoi(orderNumber)
	if err != nil {
		return fmt.Errorf("failed to convert order number to int: %w", err)
	}

	const doubleDigitThreshold = 9
	const modValue = 10

	var sum int
	double := false

	for number > 0 {
		digit := number % modValue
		number /= modValue

		if double {
			digit *= 2
			if digit > doubleDigitThreshold {
				digit -= doubleDigitThreshold
			}
		}

		sum += digit
		double = !double
	}

	if sum%modValue != 0 {
		return errors.New("invalid order number")
	}

	return nil
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

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, models.Claims{
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

func (s *Service) GetUserBalance(ctx context.Context, userID int) (float64, error) {
	balance, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

func (s *Service) CreateOrder(ctx context.Context, userID int, orderNumber string) (*models.Order, bool, error) {
	err := validateOrderNumber(orderNumber)
	if err != nil {
		return nil, false, fmt.Errorf("%w: %s", models.ErrInvalidOrderNumber, orderNumber)
	}

	order, isNew, err := s.storage.CreateOrder(ctx, userID, orderNumber, models.OrderStatusNew)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create order: %w", err)
	}

	return order, isNew, nil
}

func (s *Service) GetUserOrders(ctx context.Context, userID int) ([]models.UserOrderResult, error) {
	orders, err := s.storage.GetUserOrders(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}

	if len(orders) == 0 {
		return nil, models.ErrUserHasNoOrders
	}

	result := make([]models.UserOrderResult, len(orders))
	for i, order := range orders {
		result[i] = models.UserOrderResult{
			OrderNumber: order.OrderNumber,
			Status:      order.Status,
			Accrual:     order.Accrual,
			UploadedAt:  order.UploadedAt,
		}
	}

	return result, nil
}

func (s *Service) GetUserWithdrawals(ctx context.Context, userID int) ([]models.UserWithdrawalResult, error) {
	withdrawals, err := s.storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user withdrawals: %w", err)
	}

	if len(withdrawals) == 0 {
		return nil, models.ErrUserHasNoWithdrawals
	}

	result := make([]models.UserWithdrawalResult, len(withdrawals))
	for i, withdrawal := range withdrawals {
		result[i] = models.UserWithdrawalResult{
			Order:       withdrawal.OrderNumber,
			Sum:         withdrawal.Sum,
			ProcessedAt: withdrawal.ProcessedAt,
		}
	}

	return result, nil
}

func (s *Service) Ping(ctx context.Context) error {
	err := s.storage.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to reach storage: %w", err)
	}

	return nil
}
