package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/a-bondar/gophermart/internal/middleware"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/logger"
	"github.com/jackc/pgx/v5"
)

type application struct {
	logger *slog.Logger
	config *config.Config
}

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	cfg := config.NewConfig()
	l := logger.NewLogger()

	app := &application{
		logger: l,
		config: cfg,
	}

	mux := http.NewServeMux()
	conn, err := pgx.Connect(context.Background(), cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("unable to connect to database: %w", err)
	}

	defer func() {
		err = conn.Close(context.Background())
		if err != nil {
			app.logger.ErrorContext(context.Background(), err.Error())
		}
	}()

	err = conn.Ping(context.Background())
	if err != nil {
		return fmt.Errorf("unable to ping database: %w", err)
	}

	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		err = conn.Ping(r.Context())
		if err != nil {
			app.logger.ErrorContext(r.Context(), err.Error())
			http.Error(w, "database is not available", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})

	app.logger.InfoContext(context.Background(), "Starting server...", slog.String("address", app.config.RunAddr))

	loggedMux := middleware.WithLog(app.logger)(mux)
	err = http.ListenAndServe(app.config.RunAddr, loggedMux)
	if err != nil {
		return fmt.Errorf("unable to start server: %w", err)
	}

	return nil
}
