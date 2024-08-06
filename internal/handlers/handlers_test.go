package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/a-bondar/gophermart/mocks"

	"github.com/a-bondar/gophermart/internal/middleware"
	"github.com/stretchr/testify/require"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/handlers"
	"github.com/a-bondar/gophermart/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestHandler_HandleUserRegister(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExp:    1 * time.Hour,
	}

	mockService := new(mocks.Service)
	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name               string
		requestBody        models.HandleUserAuthRequest
		setupMockService   func()
		expectedStatusCode int
		expectedCookies    []*http.Cookie
	}{
		{
			name: "Successful registration",
			requestBody: models.HandleUserAuthRequest{
				Login:    "newuser",
				Password: "newpass",
			},
			setupMockService: func() {
				mockService.On("CreateUser", mock.Anything, "newuser", "newpass").Return(nil)
				mockService.On("AuthenticateUser", mock.Anything, "newuser", "newpass").Return("valid-token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedCookies: []*http.Cookie{
				{
					Name:  "auth_token",
					Value: "valid-token",
				},
			},
		},
		{
			name: "Duplicate login",
			requestBody: models.HandleUserAuthRequest{
				Login:    "existinguser",
				Password: "password",
			},
			setupMockService: func() {
				mockService.On("CreateUser", mock.Anything, "existinguser", "password").Return(models.ErrUserDuplicateLogin)
			},
			expectedStatusCode: http.StatusConflict,
		},
		{
			name: "Missing fields",
			requestBody: models.HandleUserAuthRequest{
				Login:    "",
				Password: "",
			},
			setupMockService:   func() {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Error during registration",
			requestBody: models.HandleUserAuthRequest{
				Login:    "erroruser",
				Password: "password",
			},
			setupMockService: func() {
				mockService.On("CreateUser", mock.Anything, "erroruser", "password").Return(errors.New("some error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests { //nolint:dupl //test case
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMockService()

			requestBodyBytes, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader(requestBodyBytes))
			rr := httptest.NewRecorder()

			handler.HandleUserRegister(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if len(tt.expectedCookies) > 0 {
				result := rr.Result()
				actualCookies := result.Cookies()
				defer result.Body.Close() //nolint:errcheck //test

				for _, expectedCookie := range tt.expectedCookies {
					assert.Condition(t, func() bool {
						for _, c := range actualCookies {
							if c.Name == expectedCookie.Name && c.Value == expectedCookie.Value {
								return true
							}
						}
						return false
					}, "Expected cookie not found: %v", expectedCookie)
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestHandler_HandleUserLogin(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExp:    1 * time.Hour,
	}

	mockService := mocks.NewService(t)
	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name               string
		requestBody        models.HandleLoginUserRequest
		setupMockService   func()
		expectedStatusCode int
		expectedCookies    []*http.Cookie
	}{
		{
			name: "Successful login",
			requestBody: models.HandleLoginUserRequest{
				Login:    "existinguser",
				Password: "correctpassword",
			},
			setupMockService: func() {
				mockService.On("AuthenticateUser", mock.Anything, "existinguser", "correctpassword").Return("valid-token", nil)
			},
			expectedStatusCode: http.StatusOK,
			expectedCookies: []*http.Cookie{
				{
					Name:  "auth_token",
					Value: "valid-token",
					Path:  "/",
				},
			},
		},
		{
			name: "Invalid credentials",
			requestBody: models.HandleLoginUserRequest{
				Login:    "existinguser",
				Password: "wrongpassword",
			},
			setupMockService: func() {
				mockService.On("AuthenticateUser", mock.Anything, "existinguser", "wrongpassword").
					Return("", models.ErrUserInvalidCredentials)
			},
			expectedStatusCode: http.StatusUnauthorized,
		},
		{
			name: "Missing fields",
			requestBody: models.HandleLoginUserRequest{
				Login:    "",
				Password: "",
			},
			setupMockService: func() {
			},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name: "Error during authentication",
			requestBody: models.HandleLoginUserRequest{
				Login:    "erroruser",
				Password: "password",
			},
			setupMockService: func() {
				mockService.On("AuthenticateUser", mock.Anything, "erroruser", "password").Return("", errors.New("some error"))
			},
			expectedStatusCode: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests { //nolint:dupl //test case
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMockService()

			requestBodyBytes, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/api/user/login", bytes.NewReader(requestBodyBytes))
			rr := httptest.NewRecorder()

			handler.HandleUserLogin(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if len(tt.expectedCookies) > 0 {
				result := rr.Result()
				actualCookies := result.Cookies()
				defer result.Body.Close() //nolint:errcheck //test

				for _, expectedCookie := range tt.expectedCookies {
					assert.Condition(t, func() bool {
						for _, c := range actualCookies {
							if c.Name == expectedCookie.Name && c.Value == expectedCookie.Value {
								return true
							}
						}
						return false
					}, "Expected cookie not found: %v", expectedCookie)
				}
			}
		})
	}
}

func TestHandler_HandleUserBalance(t *testing.T) {
	mockService := mocks.NewService(t)
	logger := slog.Default()
	cfg := &config.Config{
		JWTSecret: "test-secret",
		JWTExp:    1 * time.Hour,
	}
	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedBody   *models.Balance
		expectedError  error
	}{
		{
			name: "Successful balance retrieval",
			mockSetup: func() {
				mockService.On("GetUserBalance", mock.Anything, 123).Return(&models.Balance{
					Current:   150.5,
					Withdrawn: 50,
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: &models.Balance{
				Current:   150.5,
				Withdrawn: 50,
			},
		},
		{
			name: "Internal server error",
			mockSetup: func() {
				mockService.On("GetUserBalance", mock.Anything, 123).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			req, err := http.NewRequest(http.MethodGet, "/api/user/balance", http.NoBody)
			assert.NoError(t, err)

			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123))

			rr := httptest.NewRecorder()
			handler.HandleUserBalance(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var actualBody models.Balance
				err = json.NewDecoder(rr.Body).Decode(&actualBody)
				assert.NoError(t, err)
				assert.Equal(t, *tt.expectedBody, actualBody)
			}
		})
	}
}

func TestHandler_HandlePostUserOrders(t *testing.T) {
	mockService := mocks.NewService(t)
	logger := slog.Default()
	cfg := &config.Config{}

	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name               string
		contentType        string
		body               string
		userID             int
		setupMockService   func()
		expectedStatusCode int
	}{
		{
			name:               "Invalid content type",
			contentType:        "application/json",
			body:               "123456789",
			userID:             1,
			setupMockService:   func() {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "Empty order number",
			contentType:        "text/plain",
			body:               "",
			userID:             1,
			setupMockService:   func() {},
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:        "Invalid order number",
			contentType: "text/plain",
			body:        "123456789",
			userID:      1,
			setupMockService: func() {
				mockService.On("CreateOrder", mock.Anything, 1, "123456789").Return(nil, false, models.ErrInvalidOrderNumber)
			},
			expectedStatusCode: http.StatusUnprocessableEntity,
		},
		{
			name:        "Successful new order creation",
			contentType: "text/plain",
			body:        "123456789",
			userID:      1,
			setupMockService: func() {
				mockService.On("CreateOrder", mock.Anything, 1, "123456789").Return(&models.Order{
					OrderNumber: "123456789",
					UserID:      1,
				}, true, nil)
			},
			expectedStatusCode: http.StatusAccepted,
		},
		{
			name:        "Order already created by same user",
			contentType: "text/plain",
			body:        "123456789",
			userID:      1,
			setupMockService: func() {
				mockService.On("CreateOrder", mock.Anything, 1, "123456789").Return(&models.Order{
					OrderNumber: "123456789",
					UserID:      1,
				}, false, nil)
			},
			expectedStatusCode: http.StatusOK,
		},
		{
			name:        "Order already created by different user",
			contentType: "text/plain",
			body:        "123456789",
			userID:      1,
			setupMockService: func() {
				mockService.On("CreateOrder", mock.Anything, 1, "123456789").Return(&models.Order{
					OrderNumber: "123456789",
					UserID:      2,
				}, false, nil)
			},
			expectedStatusCode: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil

			tt.setupMockService()

			req, err := http.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.body))
			require.NoError(t, err)

			req.Header.Set("Content-Type", tt.contentType)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.HandlePostUserOrders(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)
		})
	}
}

func TestHandler_HandleGetUserOrders(t *testing.T) {
	mockService := mocks.NewService(t)
	logger := slog.Default()
	cfg := &config.Config{}
	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedBody   []models.UserOrderResult
	}{
		{
			name: "Successful order retrieval",
			mockSetup: func() {
				mockService.On("GetUserOrders", mock.Anything, 123).Return([]models.UserOrderResult{
					{
						OrderNumber: "123456",
						Status:      models.OrderStatusProcessed,
						Accrual:     100.5,
						UploadedAt:  time.Date(2024, 8, 1, 10, 0, 0, 0, time.UTC),
					},
					{
						OrderNumber: "789012",
						Status:      models.OrderStatusProcessing,
						Accrual:     50.0,
						UploadedAt:  time.Date(2024, 8, 2, 15, 30, 0, 0, time.UTC),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []models.UserOrderResult{
				{
					OrderNumber: "123456",
					Status:      models.OrderStatusProcessed,
					Accrual:     100.5,
					UploadedAt:  time.Date(2024, 8, 1, 10, 0, 0, 0, time.UTC),
				},
				{
					OrderNumber: "789012",
					Status:      models.OrderStatusProcessing,
					Accrual:     50.0,
					UploadedAt:  time.Date(2024, 8, 2, 15, 30, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "No orders for user",
			mockSetup: func() {
				mockService.On("GetUserOrders", mock.Anything, 123).Return(nil, models.ErrUserHasNoOrders)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Internal server error",
			mockSetup: func() {
				mockService.On("GetUserOrders", mock.Anything, 123).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests { //nolint:dupl //test
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			req, err := http.NewRequest(http.MethodGet, "/api/user/orders", http.NoBody)
			assert.NoError(t, err)

			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123))

			rr := httptest.NewRecorder()
			handler.HandleGetUserOrders(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var actualBody []models.UserOrderResult
				err = json.NewDecoder(rr.Body).Decode(&actualBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, actualBody)
			}
		})
	}
}

func TestHandler_HandleGetUserWithdrawals(t *testing.T) {
	mockService := mocks.NewService(t)
	logger := slog.Default()
	cfg := &config.Config{}
	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name           string
		mockSetup      func()
		expectedStatus int
		expectedBody   []models.UserWithdrawalResult
	}{
		{
			name: "Successful withdrawals retrieval",
			mockSetup: func() {
				mockService.On("GetUserWithdrawals", mock.Anything, 123).Return([]models.UserWithdrawalResult{
					{
						Order:       "123456",
						Sum:         50.0,
						ProcessedAt: time.Date(2024, 8, 2, 15, 30, 0, 0, time.UTC),
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: []models.UserWithdrawalResult{
				{
					Order:       "123456",
					Sum:         50.0,
					ProcessedAt: time.Date(2024, 8, 2, 15, 30, 0, 0, time.UTC),
				},
			},
		},
		{
			name: "No withdrawals for user",
			mockSetup: func() {
				mockService.On("GetUserWithdrawals", mock.Anything, 123).Return(nil, models.ErrUserHasNoWithdrawals)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "Internal server error",
			mockSetup: func() {
				mockService.On("GetUserWithdrawals", mock.Anything, 123).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests { //nolint:dupl //test
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			req, err := http.NewRequest(http.MethodGet, "/api/user/withdrawals", http.NoBody)
			assert.NoError(t, err)

			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123))

			rr := httptest.NewRecorder()
			handler.HandleGetUserWithdrawals(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedStatus == http.StatusOK {
				var actualBody []models.UserWithdrawalResult
				err = json.NewDecoder(rr.Body).Decode(&actualBody)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, actualBody)
			}
		})
	}
}

func TestHandler_HandleUserWithdraw(t *testing.T) {
	mockService := mocks.NewService(t)
	logger := slog.Default()
	cfg := &config.Config{}
	handler := handlers.NewHandler(mockService, logger, cfg)

	tests := []struct {
		name           string
		requestBody    models.HandleUserWithdrawRequest
		mockSetup      func()
		expectedStatus int
		expectedError  string
	}{
		{
			name: "Successful withdrawal",
			requestBody: models.HandleUserWithdrawRequest{
				Order: "123456",
				Sum:   50.0,
			},
			mockSetup: func() {
				mockService.On("UserWithdrawBonuses", mock.Anything, 123, "123456", 50.0).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid order number",
			requestBody: models.HandleUserWithdrawRequest{
				Order: "123",
				Sum:   50.0,
			},
			mockSetup: func() {
				mockService.On("UserWithdrawBonuses", mock.Anything, 123, "123", 50.0).Return(models.ErrInvalidOrderNumber)
			},
			expectedStatus: http.StatusUnprocessableEntity,
			expectedError:  "Invalid order number",
		},
		{
			name: "Insufficient funds",
			requestBody: models.HandleUserWithdrawRequest{
				Order: "123456",
				Sum:   5000.0,
			},
			mockSetup: func() {
				mockService.On("UserWithdrawBonuses", mock.Anything, 123, "123456", 5000.0).Return(models.ErrUserInsufficientFunds)
			},
			expectedStatus: http.StatusPaymentRequired,
			expectedError:  "Not enough bonuses",
		},
		{
			name: "Missing order or sum",
			requestBody: models.HandleUserWithdrawRequest{
				Order: "",
				Sum:   0,
			},
			mockSetup:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Internal server error",
			requestBody: models.HandleUserWithdrawRequest{
				Order: "123456",
				Sum:   50.0,
			},
			mockSetup: func() {
				mockService.On("UserWithdrawBonuses", mock.Anything, 123, "123456", 50.0).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService.ExpectedCalls = nil
			tt.mockSetup()

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest(http.MethodPost, "/api/user/balance/withdraw", bytes.NewBuffer(body))
			assert.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, 123))

			rr := httptest.NewRecorder()
			handler.HandleUserWithdraw(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedError != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedError)
			}
		})
	}
}
