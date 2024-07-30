package models

import (
	"errors"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUserDuplicateLogin     = errors.New("user: duplicate login")
	ErrUserNotFound           = errors.New("user: not found")
	ErrUserInvalidCredentials = errors.New("user: invalid credentials")
)

type Claims struct {
	jwt.RegisteredClaims
	UserID int
}

type User struct {
	Login          string `json:"login"`
	HashedPassword string `json:"hashed_password"`
	CreatedAt      string `json:"created_at"`
	ID             int    `json:"id"`
}

type HandleRegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type HandleLoginUserRequest = HandleRegisterUserRequest

type HandleUserBalanceResponse struct {
	Current   float64 `json:"current"`
	Withdrawn int     `json:"withdrawn"`
}
