package middleware_test

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/middleware"
	"github.com/a-bondar/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func generateToken(secret string, userID int, expired bool) string {
	expirationTime := time.Now().Add(time.Hour)
	if expired {
		expirationTime = time.Now().Add(-time.Hour)
	}

	claims := models.Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte(secret))
	return tokenString
}

func TestWithAuth(t *testing.T) {
	logger := slog.Default()
	cfg := &config.Config{JWTSecret: "test-secret"}

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{
			name:           "Valid token",
			token:          generateToken(cfg.JWTSecret, 123, false),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid token",
			token:          "invalid-token",
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Expired token",
			token:          generateToken(cfg.JWTSecret, 123, true),
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "No token",
			token:          "",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodGet, "/test", http.NoBody)
			if tt.token != "" {
				req.AddCookie(&http.Cookie{
					Name:  "auth_token",
					Value: tt.token,
				})
			}

			rr := httptest.NewRecorder()

			handler := middleware.WithAuth(logger, cfg)(mockHandler)
			handler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
		})
	}
}
