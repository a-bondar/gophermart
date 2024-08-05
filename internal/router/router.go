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
	withGzip := middleware.WithGzip(logger)

	mux.Handle("POST /api/user/register", alice.New(withLogger).ThenFunc(h.HandleUserRegister))
	mux.Handle("POST /api/user/login", alice.New(withLogger).ThenFunc(h.HandleUserLogin))
	mux.Handle("GET /ping", alice.New(withLogger).ThenFunc(h.HandlePing))

	mux.Handle("GET /api/user/balance", alice.New(withLogger, withAuth).ThenFunc(h.HandleUserBalance))
	mux.Handle("POST /api/user/orders", alice.New(withLogger, withAuth).ThenFunc(h.HandlePostUserOrders))
	mux.Handle("GET /api/user/orders", alice.New(withLogger, withGzip, withAuth).ThenFunc(h.HandleGetUserOrders))
	mux.Handle("GET /api/user/withdrawals", alice.New(withLogger, withGzip, withAuth).ThenFunc(h.HandleGetUserWithdrawals))
	mux.Handle("POST /api/user/balance/withdraw", alice.New(withLogger, withAuth).ThenFunc(h.HandleUserWithdraw))

	return mux
}
