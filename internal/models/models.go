package models

import "errors"

var (
	ErrUserDuplicateLogin = errors.New("user: duplicate login")
)

type User struct {
	ID             string `json:"id"`
	Login          string `json:"login"`
	HashedPassword string `json:"hashed_password"`
	CreatedAt      string `json:"created_at"`
}

type HandleRegisterUserRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
