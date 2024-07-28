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

func (s *Storage) GetUserBalance(ctx context.Context, userID int) (float64, error) {
	var balance float64

	err := s.pool.QueryRow(ctx, "SELECT balance FROM users WHERE id = $1", userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to get user balance: %w", err)
	}

	return balance, nil
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
