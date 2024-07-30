package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/a-bondar/gophermart/internal/middleware"

	"github.com/a-bondar/gophermart/internal/config"

	"github.com/a-bondar/gophermart/internal/models"
)

const missingRequiredFields = "Missing required fields: login or password"

type Service interface {
	CreateUser(ctx context.Context, login, password string) error
	AuthenticateUser(ctx context.Context, login, password string) (string, error)
	GetUserBalance(ctx context.Context, userID int) (float64, error)
	CreateOrder(ctx context.Context, userID int, orderNumber int) (*models.Order, bool, error)
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

func (h *Handler) HandleUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	balance, err := h.service.GetUserBalance(r.Context(), userID)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// @TODO: add withdrawn
	response := models.HandleUserBalanceResponse{Current: balance, Withdrawn: 0}

	if err = json.NewEncoder(w).Encode(response); err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandlePostUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("Content-Type") != "text/plain" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	orderNumberStr := strings.TrimSpace(string(body))
	if orderNumberStr == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
		return
	}

	orderNumber, err := strconv.Atoi(orderNumberStr)
	if err != nil {
		http.Error(w, "Invalid order number format", http.StatusUnprocessableEntity)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	order, isNew, err := h.service.CreateOrder(r.Context(), userID, orderNumber)
	if err != nil {
		if errors.Is(err, models.ErrInvalidOrderNumber) {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}

		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	if order.UserID != userID {
		http.Error(w, "Order has been already created", http.StatusConflict)
		return
	}

	var status = http.StatusAccepted

	if !isNew {
		status = http.StatusCreated
	}

	w.WriteHeader(status)
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
