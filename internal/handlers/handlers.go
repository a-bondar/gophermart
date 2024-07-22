package handlers

import (
	"context"
	"log/slog"
	"net/http"
)

type Service interface {
	Ping(ctx context.Context) error
}

type Handler struct {
	service Service
	logger  *slog.Logger
}

func NewHandler(service Service, logger *slog.Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write([]byte(`{"status": "ok"}`)); err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
