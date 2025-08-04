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

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UpdateProfileRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FullName     string `json:"full_name"`
	TelegramUser string `json:"telegram_user"`
}

type UserInfo struct {
	ID           int             `json:"id"`
	Username     string          `json:"username"`
	Email        string          `json:"email"`
	Phone        string          `json:"phone"`
	FullName     string          `json:"full_name"`
	UserType     models.UserType `json:"user_type"`
	TelegramUser string          `json:"telegram_user"`
}

type TelegramInfo struct {
	Username  string `json:"username"`
	Connected bool   `json:"connected"`
}

type SignUpResponse struct {
	User                      UserInfo `json:"user"`
	TelegramSetupRequired     bool     `json:"telegram_setup_required"`
	TelegramSetupInstructions string   `json:"telegram_setup_instructions,omitempty"`
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

type ProfileResponse struct {
	ID       int             `json:"id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	FullName string          `json:"full_name"`
	UserType models.UserType `json:"user_type"`
	Telegram TelegramInfo    `json:"telegram"`
}

type PublicUserResponse struct {
	ID       int             `json:"id"`
	Username string          `json:"username"`
	FullName string          `json:"full_name"`
	UserType models.UserType `json:"user_type,omitempty"`
}
