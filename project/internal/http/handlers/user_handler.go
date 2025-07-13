package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo repositories.UserRepository
}

type CreateUserRequest struct {
	Username string          `json:"username"`
	Password string          `json:"password"`
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

func (h *UserHandler) getCurrentUserID(r *http.Request) (int, error) {
	userIDValue := r.Context().Value(middleware.UserIDKey)
	if userIDValue == nil {
		return 0, fmt.Errorf("user ID not found in context")
	}

	userIDStr, ok := userIDValue.(string)
	if !ok {
		return 0, fmt.Errorf("invalid user ID format in context")
	}

	return strconv.Atoi(userIDStr)
}

func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req models.User
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	if !utils.ValidateRequiredFields(w, map[string]string{
		"username": req.Username,
		"password": req.Password,
		"email":    req.Email,
	}) {
		return
	}

	if req.UserType != models.Manager && req.UserType != models.Resident {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user type")
		return
	}

	existingUser, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil && err != sql.ErrNoRows {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to check existing user")
		return
	}
	if existingUser != nil {
		utils.WriteErrorResponse(w, http.StatusConflict, "username already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		Email:    req.Email,
		Phone:    req.Phone,
		FullName: req.FullName,
		UserType: req.UserType,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := h.userRepo.CreateUser(ctx, user)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	user.ID = userID

	token, err := middleware.GenerateToken(strconv.Itoa(userID), user.UserType)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.WriteSuccessResponse(w, "user created successfully", map[string]interface{}{
		"user":  user,
		"token": token,
	})
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	if !utils.ValidateRequiredFields(w, map[string]string{
		"username": req.Username,
		"password": req.Password,
	}) {
		return
	}

	existingUser, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, err := middleware.GenerateToken(strconv.Itoa(existingUser.ID), existingUser.UserType)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	utils.WriteSuccessResponse(w, "login successful", map[string]interface{}{
		"token":     token,
		"user_id":   strconv.Itoa(existingUser.ID),
		"user_type": string(existingUser.UserType),
		"username":  existingUser.Username,
		"email":     existingUser.Email,
		"full_name": existingUser.FullName,
	})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	utils.WriteSuccessResponse(w, "profile retrieved successfully", user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req UpdateProfileRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		return
	}

	existingUser, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	if req.Username != "" {
		existingUser.Username = req.Username
	}
	if req.Email != "" {
		existingUser.Email = req.Email
	}
	if req.Phone != "" {
		existingUser.Phone = req.Phone
	}
	if req.FullName != "" {
		existingUser.FullName = req.FullName
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.userRepo.UpdateUser(ctx, *existingUser); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	utils.WriteSuccessResponse(w, "profile updated successfully", existingUser)
}

func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	file, header, err := utils.ParseFileUpload(w, r, "profile_picture", 10<<20)
	if err != nil {
		return
	}
	defer file.Close()

	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif"}
	if !utils.ValidateFileType(w, header.Filename, allowedTypes) {
		return
	}

	_, err = io.ReadAll(file)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	fileName := fmt.Sprintf("profile_%d_%d%s", userID, time.Now().Unix(), filepath.Ext(header.Filename))
	utils.WriteSuccessResponse(w, "profile picture uploaded successfully", map[string]string{
		"filename": fileName,
	})
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ParseIDFromPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	utils.WriteSuccessResponse(w, "user retrieved successfully", user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAllUsers(r.Context())
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	utils.WriteSuccessResponse(w, "users retrieved successfully", users)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ParseIDFromPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.userRepo.DeleteUser(userID); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	utils.WriteSuccessResponse(w, "user deleted successfully", nil)
}
