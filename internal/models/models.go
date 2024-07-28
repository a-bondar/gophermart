package models

import "errors"

var (
	ErrUserDuplicateLogin     = errors.New("user: duplicate login")
	ErrUserNotFound           = errors.New("user: not found")
	ErrUserInvalidCredentials = errors.New("user: invalid credentials")
)

type User struct {
	Login          string `json:"login"`
	HashedPassword string `json:"hashed_password"`
	CreatedAt      string `json:"created_at"`
	ID             int64  `json:"id"`
}

type HandleRegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

type HandleLoginUserRequest = HandleRegisterUserRequest
