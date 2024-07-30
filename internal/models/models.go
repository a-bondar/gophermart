package models

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrUserDuplicateLogin     = errors.New("user: duplicate login")
	ErrUserNotFound           = errors.New("user: not found")
	ErrUserInvalidCredentials = errors.New("user: invalid credentials")
	ErrInvalidOrderNumber     = errors.New("invalid order number")
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

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "NEW"
	OrderStatusProcessing OrderStatus = "PROCESSING"
	OrderStatusInvalid    OrderStatus = "INVALID"
	OrderStatusProcessed  OrderStatus = "PROCESSED"
)

type Order struct {
	UploadedAt  time.Time   `json:"uploaded_at"`
	Status      OrderStatus `json:"status"`
	OrderNumber int         `json:"order_number"`
	UserID      int         `json:"user_id"`
	ID          int         `json:"id"`
	Accrual     float64     `json:"accrual"`
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
