package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/a-bondar/gophermart/internal/config"

	"github.com/a-bondar/gophermart/internal/models"
)

const missingRequiredFields = "Missing required fields: login or password"

type Service interface {
	CreateUser(ctx context.Context, login, password string) error
	AuthenticateUser(ctx context.Context, login, password string) (string, error)
	Ping(ctx context.Context) error
}

type Handler struct {
	service Service
	logger  *slog.Logger
	cfg     *config.Config
}

func NewHandler(service Service, logger *slog.Logger, cfg *config.Config) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
		cfg:     cfg,
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
		h.logger.ErrorContext(r.Context(), missingRequiredFields)
		http.Error(w, missingRequiredFields, http.StatusBadRequest)
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

	token, err := h.service.AuthenticateUser(r.Context(), request.Login, request.Password)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(h.cfg.JWTExp),
		HttpOnly: true,
	})

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	var request models.HandleLoginUserRequest
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
		h.logger.ErrorContext(r.Context(), missingRequiredFields)
		http.Error(w, missingRequiredFields, http.StatusBadRequest)
		return
	}

	token, err := h.service.AuthenticateUser(r.Context(), request.Login, request.Password)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		var message string
		status := http.StatusInternalServerError

		if errors.Is(err, models.ErrUserInvalidCredentials) {
			message = "invalid login or password"
			status = http.StatusUnauthorized
		}

		http.Error(w, message, status)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Expires:  time.Now().Add(h.cfg.JWTExp),
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
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
