package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/models"

	"golang.org/x/crypto/bcrypt"
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
}

type Service struct {
	storage    Storage
	logger     *slog.Logger
	cfg        *config.Config
	httpClient *http.Client
	ticker     *time.Ticker
	sleepUntil int64
}

func NewService(storage Storage, logger *slog.Logger, cfg *config.Config) *Service {
	return &Service{
		storage:    storage,
		logger:     logger,
		cfg:        cfg,
		httpClient: &http.Client{Timeout: HTTPClientTimeout},
	}
}

const CheckOrderAccrualStatusInterval = 10 * time.Second
const HTTPClientTimeout = 5 * time.Second

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

func (s *Service) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	balance, err := s.storage.GetUserBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
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

func (s *Service) UserWithdrawBonuses(ctx context.Context, userID int, orderNumber string, sum float64) error {
	err := validateOrderNumber(orderNumber)
	if err != nil {
		return fmt.Errorf("%w: %s", models.ErrInvalidOrderNumber, orderNumber)
	}

	err = s.storage.UserWithdrawBonuses(ctx, userID, orderNumber, sum)
	if err != nil {
		return fmt.Errorf("failed to withdraw bonuses: %w", err)
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

func (s *Service) processOrder(ctx context.Context, orderNumber string) error {
	requestURL, err := url.JoinPath(s.cfg.AccrualSystemAddress, "api/orders", orderNumber)
	if err != nil {
		return fmt.Errorf("failed to join URL: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, http.NoBody)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	defer func() {
		err = resp.Body.Close()
		if err != nil {
			s.logger.ErrorContext(ctx, err.Error())
		}
	}()

	switch resp.StatusCode {
	case http.StatusOK:
		var orderUpdate models.AccrualServiceResponse
		if err = json.NewDecoder(resp.Body).Decode(&orderUpdate); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		err = s.storage.UpdateOrder(ctx, orderNumber, orderUpdate.Status, orderUpdate.Accrual)
		if err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}
	case http.StatusTooManyRequests:
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			duration, err := strconv.Atoi(retryAfter)
			if err == nil {
				atomic.StoreInt64(&s.sleepUntil, time.Now().Add(time.Duration(duration)*time.Second).Unix())
				s.logger.WarnContext(ctx, "Rate limited, pausing updates", slog.Int("retryAfterSeconds", duration))
			}
		}
	case http.StatusNoContent:
		s.logger.WarnContext(ctx, "Order not found", slog.String("orderNumber", orderNumber))
	case http.StatusInternalServerError:
		s.logger.WarnContext(ctx, "Internal server error from external service")
	default:
		s.logger.WarnContext(ctx, "Unexpected status from external service", slog.Int("status", resp.StatusCode))
	}

	return nil
}

func (s *Service) updateOrderStatuses(ctx context.Context) {
	orders, err := s.storage.GetPendingOrders(ctx)
	if err != nil {
		s.logger.ErrorContext(ctx, err.Error())
		return
	}

	for _, order := range orders {
		err = s.processOrder(ctx, order.OrderNumber)
		if err != nil {
			s.logger.ErrorContext(ctx, err.Error(), slog.String("orderNumber", order.OrderNumber))
		}
	}
}

func (s *Service) StartOrderAccrualStatusJob(ctx context.Context) {
	s.ticker = time.NewTicker(CheckOrderAccrualStatusInterval)

	go func() {
		for {
			select {
			case <-s.ticker.C:
				if time.Now().Before(time.Unix(atomic.LoadInt64(&s.sleepUntil), 0)) {
					s.logger.InfoContext(ctx,
						"Sleeping due to rate limit", slog.Time("until", time.Unix(atomic.LoadInt64(&s.sleepUntil), 0)))
					continue
				}

				s.updateOrderStatuses(ctx)
			case <-ctx.Done():
				s.logger.InfoContext(ctx, "Check Accrual Order Status job stopped")
				return
			}
		}
	}()
}

func (s *Service) StopOrderAccrualStatusJob() {
	s.ticker.Stop()
}
