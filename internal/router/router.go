package router

import (
	"log/slog"
	"net/http"

	"github.com/a-bondar/gophermart/internal/config"

	"github.com/a-bondar/gophermart/internal/handlers"
	"github.com/a-bondar/gophermart/internal/middleware"
	"github.com/justinas/alice"
)

func Router(h *handlers.Handler, logger *slog.Logger, cfg *config.Config) http.Handler {
	mux := http.NewServeMux()

	withLogger := middleware.WithLog(logger)
	withAuth := middleware.WithAuth(logger, cfg)

	mux.Handle("/ping", alice.New(withLogger, withAuth).ThenFunc(h.HandlePing))
	mux.Handle("/api/user/register", alice.New(withLogger).ThenFunc(h.HandleUserRegister))
	mux.Handle("/api/user/login", alice.New(withLogger).ThenFunc(h.HandleUserLogin))
	mux.Handle("/api/user/balance", alice.New(withLogger, withAuth).ThenFunc(h.HandleUserBalance))

	return mux
}
