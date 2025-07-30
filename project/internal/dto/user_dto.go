package dto

import "github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"

type CreateUserRequest struct {
	Username     string          `json:"username"`
	Password     string          `json:"password"`
	Email        string          `json:"email"`
	Phone        string          `json:"phone"`
	FullName     string          `json:"full_name"`
	UserType     models.UserType `json:"user_type"`
	TelegramUser string          `json:"telegram_user"`
}

type UpdateProfileRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FullName     string `json:"full_name"`
	TelegramUser string `json:"telegram_user"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string       `json:"token"`
	UserID   string       `json:"user_id"`
	UserType string       `json:"user_type"`
	Username string       `json:"username"`
	Email    string       `json:"email"`
	FullName string       `json:"full_name"`
	Telegram TelegramInfo `json:"telegram"`
}

type TelegramInfo struct {
	Username  string `json:"username"`
	Connected bool   `json:"connected"`
}
