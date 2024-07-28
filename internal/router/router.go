package router

import (
	"log/slog"
	"net/http"

	"github.com/a-bondar/gophermart/internal/handlers"
	"github.com/a-bondar/gophermart/internal/middleware"
	"github.com/justinas/alice"
)

func Router(h *handlers.Handler, logger *slog.Logger) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /ping", h.HandlePing)

	mux.HandleFunc("POST /api/user/register", h.HandleUserRegister)
	mux.HandleFunc("POST /api/user/login", h.HandleUserLogin)

	chain := alice.New(middleware.WithLog(logger)).Then(mux)

	return chain
}
