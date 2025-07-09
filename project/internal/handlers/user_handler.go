package handlers

import (
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/repositories"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type CreateUserRequest struct {
	Username string          `json:"username"`
	Password string          `json:"password"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	FullName string          `json:"full_name"`
	UserType models.UserType `json:"user_type"`
}

type UpdateUserRequest struct {
	ID       int             `json:"id"`
	Username string          `json:"username"`
	Email    string          `json:"email"`
	Phone    string          `json:"phone"`
	FullName string          `json:"full_name"`
	UserType models.UserType `json:"user_type"`
}

type UpdateProfileRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
}

func NewUserHandler(userRepo repositories.UserRepository) *UserHandler {
	return &UserHandler{
		userRepo: userRepo,
	}
}
