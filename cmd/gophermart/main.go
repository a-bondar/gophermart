package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/handlers"
	"github.com/a-bondar/gophermart/internal/logger"
	"github.com/a-bondar/gophermart/internal/router"
	"github.com/a-bondar/gophermart/internal/service"
	"github.com/a-bondar/gophermart/internal/storage"
)

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	cfg := config.NewConfig()
	l := logger.NewLogger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	s, err := storage.NewStorage(ctx, cfg.DatabaseURI)
	if err != nil {
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer s.Close()

	svc := service.NewService(s, l, cfg)
	h := handlers.NewHandler(svc, l, cfg)
	r := router.Router(h, l, cfg)

	svc.StartOrderAccrualStatusJob(ctx)
	defer svc.StopOrderAccrualStatusJob()

	l.InfoContext(ctx, "Running server", slog.String("address", cfg.RunAddr))

	err = http.ListenAndServe(cfg.RunAddr, r)
	if err != nil {
		if !errors.Is(err, http.ErrServerClosed) {
			l.ErrorContext(ctx, err.Error())

			return fmt.Errorf("HTTP server has encountered an error: %w", err)
		}
	}

	return nil
}
