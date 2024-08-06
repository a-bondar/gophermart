package middleware

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/a-bondar/gophermart/internal/config"
	"github.com/a-bondar/gophermart/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type key int

const (
	UserIDKey key = iota
)

func GetUserIDFromContext(ctx context.Context) (int, error) {
	value := ctx.Value(UserIDKey)
	if value == nil {
		return 0, errors.New("user ID not found in context")
	}

	userID, ok := value.(int)
	if !ok {
		return 0, errors.New("context user ID is not an int")
	}

	return userID, nil
}

func validateToken(secret, tokenString string) (int, error) {
	token, err := jwt.ParseWithClaims(tokenString, &models.Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return 0, fmt.Errorf("failed to parse token: %w", err)
	}

	if claims, ok := token.Claims.(*models.Claims); ok && token.Valid {
		return claims.UserID, nil
	}

	return 0, jwt.ErrSignatureInvalid
}

func WithAuth(logger *slog.Logger, cfg *config.Config) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth_token")
			if err != nil {
				if !errors.Is(err, http.ErrNoCookie) {
					logger.ErrorContext(r.Context(), err.Error())
					http.Error(w, "", http.StatusInternalServerError)
					return
				}

				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			userID, err := validateToken(cfg.JWTSecret, cookie.Value)
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			h.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
