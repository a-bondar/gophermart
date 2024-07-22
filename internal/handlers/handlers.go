package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/a-bondar/gophermart/internal/models"
)

type Service interface {
	CreateUser(ctx context.Context, login, password string) error
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

func (h *Handler) HandleUserRegister(w http.ResponseWriter, r *http.Request) {
	var request models.HandleRegisterUserRequest
	var buf bytes.Buffer

	_, err := buf.ReadFrom(r.Body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(buf.Bytes(), &request)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if request.Login == "" || request.Password == "" {
		h.logger.ErrorContext(r.Context(), "Missing required fields: login or password")
		http.Error(w, "Missing required fields: login or password", http.StatusBadRequest)
		return
	}

	err = h.service.CreateUser(r.Context(), request.Login, request.Password)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		var message string
		status := http.StatusInternalServerError

		if errors.Is(err, models.ErrUserDuplicateLogin) {
			message = "login already exists"
			status = http.StatusConflict
		}

		http.Error(w, message, status)
		return
	}

	// @TODO: add auth token generation and return it in response

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	h.logger.InfoContext(r.Context(), "User logged in")
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
