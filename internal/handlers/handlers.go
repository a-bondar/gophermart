package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/a-bondar/gophermart/internal/middleware"

	"github.com/a-bondar/gophermart/internal/config"

	"github.com/a-bondar/gophermart/internal/models"
)

const (
	ContentType     = "Content-Type"
	ApplicationJSON = "application/json"
)

type Service interface {
	CreateUser(ctx context.Context, login, password string) error
	AuthenticateUser(ctx context.Context, login, password string) (string, error)
	GetUserBalance(ctx context.Context, userID int) (*models.Balance, error)
	CreateOrder(ctx context.Context, userID int, orderNumber string) (*models.Order, bool, error)
	GetUserOrders(ctx context.Context, userID int) ([]models.UserOrderResult, error)
	GetUserWithdrawals(ctx context.Context, userID int) ([]models.UserWithdrawalResult, error)
	UserWithdrawBonuses(ctx context.Context, userID int, orderNumber string, sum float64) error
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

func (h *Handler) handleUserAuth(
	w http.ResponseWriter,
	r *http.Request,
	isRegistration bool,
) {
	var buf bytes.Buffer

	if _, err := buf.ReadFrom(r.Body); err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Failed to read request body", http.StatusBadRequest)
		return
	}

	var request models.HandleUserAuthRequest
	if err := json.Unmarshal(buf.Bytes(), &request); err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Failed to decode JSON", http.StatusInternalServerError)
		return
	}

	if request.Login == "" || request.Password == "" {
		h.logger.ErrorContext(r.Context(), "Missing required fields: login or password")
		http.Error(w, "Missing required fields: login or password", http.StatusBadRequest)
		return
	}

	if isRegistration {
		if err := h.service.CreateUser(r.Context(), request.Login, request.Password); err != nil {
			h.handleRegistrationError(r.Context(), w, err)
			return
		}
	}

	token, err := h.service.AuthenticateUser(r.Context(), request.Login, request.Password)
	if err != nil {
		h.handleAuthError(r.Context(), w, err)
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

func (h *Handler) HandleUserRegister(w http.ResponseWriter, r *http.Request) {
	h.handleUserAuth(w, r, true)
}

func (h *Handler) HandleUserLogin(w http.ResponseWriter, r *http.Request) {
	h.handleUserAuth(w, r, false)
}

func (h *Handler) handleRegistrationError(ctx context.Context, w http.ResponseWriter, err error) {
	var message string
	status := http.StatusInternalServerError

	if errors.Is(err, models.ErrUserDuplicateLogin) {
		message = "Login already exists"
		status = http.StatusConflict
	} else {
		message = "Internal server error"
	}

	h.logger.ErrorContext(ctx, err.Error())
	http.Error(w, message, status)
}

func (h *Handler) handleAuthError(ctx context.Context, w http.ResponseWriter, err error) {
	var message string
	status := http.StatusInternalServerError

	if errors.Is(err, models.ErrUserInvalidCredentials) {
		message = "Invalid login or password"
		status = http.StatusUnauthorized
	} else {
		message = "Internal server error"
	}

	h.logger.ErrorContext(ctx, err.Error())
	http.Error(w, message, status)
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

	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(balance); err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandlePostUserOrders(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get(ContentType) != "text/plain" {
		http.Error(w, "Invalid Content-Type", http.StatusBadRequest)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		http.Error(w, "Order number is required", http.StatusBadRequest)
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
		status = http.StatusOK
	}

	w.WriteHeader(status)
}

func handleUserResponse[T any](
	w http.ResponseWriter,
	r *http.Request,
	getDataFunc func(ctx context.Context, userID int) ([]T, error),
	noDataErr error,
	logger *slog.Logger,
) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	data, err := getDataFunc(r.Context(), userID)
	if err != nil {
		if errors.Is(err, noDataErr) {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(data); err != nil {
		logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "Failed to encode data", http.StatusInternalServerError)
		return
	}
}

func (h *Handler) HandleGetUserOrders(w http.ResponseWriter, r *http.Request) {
	handleUserResponse[models.UserOrderResult](
		w,
		r,
		h.service.GetUserOrders,
		models.ErrUserHasNoOrders,
		h.logger,
	)
}

func (h *Handler) HandleGetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	handleUserResponse[models.UserWithdrawalResult](
		w,
		r,
		h.service.GetUserWithdrawals,
		models.ErrUserHasNoWithdrawals,
		h.logger,
	)
}

func (h *Handler) HandleUserWithdraw(w http.ResponseWriter, r *http.Request) {
	var request models.HandleUserWithdrawRequest
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

	if request.Order == "" || request.Sum <= 0 {
		http.Error(w, "", http.StatusBadRequest)
		return
	}

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	err = h.service.UserWithdrawBonuses(r.Context(), userID, request.Order, request.Sum)
	if err != nil {
		if errors.Is(err, models.ErrInvalidOrderNumber) {
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
			return
		}

		if errors.Is(err, models.ErrUserInsufficientFunds) {
			http.Error(w, "Not enough bonuses", http.StatusPaymentRequired)
			return
		}

		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) HandlePing(w http.ResponseWriter, r *http.Request) {
	err := h.service.Ping(r.Context())
	if err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set(ContentType, ApplicationJSON)
	w.WriteHeader(http.StatusOK)

	if _, err = w.Write([]byte(`{"status": "ok"}`)); err != nil {
		h.logger.ErrorContext(r.Context(), err.Error())
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
}
