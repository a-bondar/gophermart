package postgres

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"time"

	"github.com/a-bondar/gophermart/internal/models"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

//go:embed migrations/*.sql
var migrationsDir embed.FS

func runMigrations(dsn string) error {
	d, err := iofs.New(migrationsDir, "migrations")
	if err != nil {
		return fmt.Errorf("failed to return an iofs driver: %w", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", d, dsn)
	if err != nil {
		return fmt.Errorf("failed to get a new migrate instance: %w", err)
	}
	if err = m.Up(); err != nil {
		if !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("failed to apply migrations to the DB: %w", err)
		}
	}
	return nil
}

func NewStorage(ctx context.Context, dsn string) (*Storage, error) {
	if err := runMigrations(dsn); err != nil {
		return nil, fmt.Errorf("failed to run DB migrations: %w", err)
	}

	poolCfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to initalize a connection pool: %w", err)
	}

	return &Storage{pool: pool}, nil
}

func (s *Storage) CreateUser(ctx context.Context, login string, hashedPassword []byte) error {
	_, err := s.pool.Exec(ctx, "INSERT INTO users (login, hashed_password) VALUES ($1, $2)", login, hashedPassword)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
			return fmt.Errorf("login already exists: %w", models.ErrUserDuplicateLogin)
		}

		return fmt.Errorf("failed to insert user: %w", err)
	}

	return nil
}

func (s *Storage) SelectUser(ctx context.Context, login string) (*models.User, error) {
	var id int
	var hashedPassword string
	var createdAt time.Time

	err := s.pool.
		QueryRow(ctx, "SELECT id, hashed_password, created_at FROM users WHERE login = $1", login).
		Scan(&id, &hashedPassword, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}

		return nil, fmt.Errorf("failed to select user: %w", err)
	}

	return &models.User{
		ID:             id,
		Login:          login,
		HashedPassword: hashedPassword,
		CreatedAt:      createdAt.Format(time.RFC3339),
	}, nil
}

func (s *Storage) GetUserBalance(ctx context.Context, userID int) (*models.Balance, error) {
	balance := &models.Balance{}
	query := `
		SELECT u.balance,
		   	(SELECT COALESCE(SUM(w.sum), 0) 
		   	 FROM withdrawals w 
		   	 WHERE w.user_id = u.id) AS withdrawn
		FROM users u
		WHERE u.id = $1
	`
	err := s.pool.QueryRow(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
}

func (s *Storage) CreateOrder(
	ctx context.Context, userID int, orderNumber string, status models.OrderStatus) (*models.Order, bool, error) {
	query := `
		WITH ins AS (
			INSERT INTO orders (user_id, order_number, status)
			VALUES ($1, $2, $3)
			ON CONFLICT (order_number) DO NOTHING
			RETURNING id, user_id, order_number, accrual, status, uploaded_at, true AS is_new
		)
		SELECT id, user_id, order_number, accrual, status, uploaded_at, is_new
		FROM ins
		UNION ALL
		SELECT id, user_id, order_number, accrual, status, uploaded_at, false AS is_new
		FROM orders
		WHERE order_number = $2
		LIMIT 1;
	`

	var order models.Order
	var isNew bool
	err := s.pool.QueryRow(ctx, query, userID, orderNumber, status).Scan(
		&order.ID, &order.UserID, &order.OrderNumber, &order.Accrual, &order.Status, &order.UploadedAt, &isNew,
	)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create order: %w", err)
	}

	return &order, isNew, nil
}

func (s *Storage) GetUserOrders(ctx context.Context, userID int) ([]models.Order, error) {
	query := "SELECT * FROM orders WHERE user_id = $1 ORDER BY uploaded_at"
	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query orders: %w", err)
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Order])
	if err != nil {
		return nil, fmt.Errorf("unable to collect rows: %w", err)
	}

	return orders, nil
}

func (s *Storage) GetUserWithdrawals(ctx context.Context, userID int) ([]models.Withdrawal, error) {
	query := "SELECT * FROM withdrawals WHERE user_id = $1 ORDER BY processed_at"
	rows, err := s.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("unable to query withdrawals: %w", err)
	}

	withdrawals, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Withdrawal])
	if err != nil {
		return nil, fmt.Errorf("unable to collect rows: %w", err)
	}

	return withdrawals, nil
}

func (s *Storage) UserWithdrawBonuses(ctx context.Context, userID int, orderNumber string, sum float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
		}
	}()

	updateQuery := `
		UPDATE users
		SET balance = balance - $1
		WHERE id = $2
	`
	_, err = tx.Exec(ctx, updateQuery, sum, userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == pgerrcode.CheckViolation {
				return models.ErrUserInsufficientFunds
			}
		}

		return fmt.Errorf("failed to update user balance: %w", err)
	}

	insertQuery := `
		INSERT INTO withdrawals (user_id, sum, order_number)
		VALUES ($1, $2, $3)
	`
	_, err = tx.Exec(ctx, insertQuery, userID, sum, orderNumber)
	if err != nil {
		return fmt.Errorf("failed to insert withdrawal: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *Storage) UpdateOrder(ctx context.Context,
	orderNumber string, status models.OrderStatus, accrual float64) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			err = tx.Rollback(ctx)
		} else {
			err = tx.Commit(ctx)
		}
	}()

	updateOrderQuery := "UPDATE orders SET status = $1, accrual = $2 WHERE order_number = $3 RETURNING user_id"
	var userID int
	err = tx.QueryRow(ctx, updateOrderQuery, status, accrual, orderNumber).Scan(&userID)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	updateBalanceQuery := "UPDATE users SET balance = balance + $1 WHERE id = $2"
	_, err = tx.Exec(ctx, updateBalanceQuery, accrual, userID)
	if err != nil {
		return fmt.Errorf("failed to update user balance: %w", err)
	}

	return nil
}

func (s *Storage) GetPendingOrders(ctx context.Context) ([]models.Order, error) {
	query := "SELECT * FROM orders WHERE status IN ($1, $2)"
	rows, err := s.pool.Query(ctx, query, models.OrderStatusNew, models.OrderStatusProcessing)
	if err != nil {
		return nil, fmt.Errorf("unable to query orders: %w", err)
	}

	orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Order])
	if err != nil {
		return nil, fmt.Errorf("unable to collect rows: %w", err)
	}

	return orders, nil
}

func (s *Storage) Ping(ctx context.Context) error {
	err := s.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("failed to ping DB: %w", err)
	}

	return nil
}

func (s *Storage) Close() {
	s.pool.Close()
}
