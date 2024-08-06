package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/handlers"
	"github.com/a-bondar/gophermart/internal/logger"
	"github.com/a-bondar/gophermart/internal/router"
	"github.com/a-bondar/gophermart/internal/service"
	"github.com/a-bondar/gophermart/internal/storage"
)

const serverShutdownTimeout = 5 * time.Second

func main() {
	if err := Run(); err != nil {
		log.Fatal(err)
	}
}

func Run() error {
	cfg := config.NewConfig()
	l := logger.NewLogger()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s, err := storage.NewStorage(ctx, cfg.DatabaseURI)
	if err != nil {
		l.ErrorContext(ctx, fmt.Sprintf("failed to initialize storage: %v", err))
		return fmt.Errorf("failed to initialize storage: %w", err)
	}
	defer s.Close()

	svc := service.NewService(s, l, cfg)
	h := handlers.NewHandler(svc, l, cfg)
	r := router.Router(h, l, cfg)

	svc.StartOrderAccrualStatusJob(ctx)
	defer svc.StopOrderAccrualStatusJob()

	server := &http.Server{
		Addr:    cfg.RunAddr,
		Handler: r,
	}

	go func() {
		l.InfoContext(ctx, "Running server", slog.String("address", cfg.RunAddr))
		if err = server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			l.ErrorContext(ctx, fmt.Sprintf("HTTP server has encountered an error: %v", err))
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), serverShutdownTimeout)
	defer shutdownCancel()

	if err = server.Shutdown(shutdownCtx); err != nil {
		l.ErrorContext(ctx, fmt.Sprintf("HTTP server shutdown failed: %v", err))
		return fmt.Errorf("HTTP server shutdown failed: %w", err)
	}

	l.InfoContext(ctx, "Server stopped gracefully")
	return nil
}
