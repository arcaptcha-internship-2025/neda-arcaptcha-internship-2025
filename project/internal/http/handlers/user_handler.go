package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"golang.org/x/crypto/bcrypt"
)

type UserHandler struct {
	userRepo repositories.UserRepository
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
	var req dto.CreateUserRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "Invalid request body")
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

	//validating telegram username format if provided
	if req.TelegramUser != "" {
		if !isValidTelegramUsername(req.TelegramUser) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid Telegram username format")
			return
		}
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

	//checking if telegram username is already taken
	if req.TelegramUser != "" {
		existingTelegramUser, err := h.userRepo.GetUserByTelegramUser(req.TelegramUser)
		if err != nil && err != sql.ErrNoRows {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to check Telegram username")
			return
		}
		if existingTelegramUser != nil {
			utils.WriteErrorResponse(w, http.StatusConflict, "Telegram username already in use")
			return
		}
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to hash password")
		return
	}

	user := models.User{
		Username:     req.Username,
		Password:     string(hashedPassword),
		Email:        req.Email,
		Phone:        req.Phone,
		FullName:     req.FullName,
		UserType:     req.UserType,
		TelegramUser: req.TelegramUser,
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

	responseData := map[string]interface{}{
		"user":                    user,
		"token":                   token,
		"telegram_setup_required": req.TelegramUser != "",
	}

	//add bot address to this message
	if req.TelegramUser != "" {
		responseData["telegram_setup_instructions"] = "Please start a chat with our bot in Telegram to complete setup : "
	}

	utils.WriteSuccessResponse(w, "user created successfully", responseData)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req dto.LoginRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
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

	response := dto.LoginResponse{
		Token:    token,
		UserID:   strconv.Itoa(existingUser.ID),
		UserType: string(existingUser.UserType),
		Username: existingUser.Username,
		Email:    existingUser.Email,
		FullName: existingUser.FullName,
		Telegram: dto.TelegramInfo{
			Username:  existingUser.TelegramUser,
			Connected: existingUser.TelegramChatID != 0,
		},
	}

	utils.WriteSuccessResponse(w, "login successful", response)
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

	profileResponse := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"email":     user.Email,
		"phone":     user.Phone,
		"full_name": user.FullName,
		"user_type": user.UserType,
		"telegram": map[string]interface{}{
			"username":  user.TelegramUser,
			"connected": user.TelegramChatID != 0,
		},
	}

	utils.WriteSuccessResponse(w, "profile retrieved successfully", profileResponse)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req dto.UpdateProfileRequest
	if err := utils.DecodeJSONBody(w, r, &req); err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existingUser, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	//checking if Telegram username is being updated and validate it
	if req.TelegramUser != "" && req.TelegramUser != existingUser.TelegramUser {
		if !isValidTelegramUsername(req.TelegramUser) {
			utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid Telegram username format")
			return
		}

		//checking if new Telegram username is already taken
		existingTelegramUser, err := h.userRepo.GetUserByTelegramUser(req.TelegramUser)
		if err != nil && err != sql.ErrNoRows {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to check Telegram username")
			return
		}
		if existingTelegramUser != nil && existingTelegramUser.ID != userID {
			utils.WriteErrorResponse(w, http.StatusConflict, "Telegram username already in use")
			return
		}
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
	if req.TelegramUser != "" {
		existingUser.TelegramUser = req.TelegramUser
		//reseting chat id if Telegram username is changed
		if req.TelegramUser != existingUser.TelegramUser {
			existingUser.TelegramChatID = 0
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.userRepo.UpdateUser(ctx, *existingUser); err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	updatedProfile := map[string]interface{}{
		"id":        existingUser.ID,
		"username":  existingUser.Username,
		"email":     existingUser.Email,
		"phone":     existingUser.Phone,
		"full_name": existingUser.FullName,
		"user_type": existingUser.UserType,
		"telegram": map[string]interface{}{
			"username":  existingUser.TelegramUser,
			"connected": existingUser.TelegramChatID != 0,
		},
	}

	utils.WriteSuccessResponse(w, "profile updated successfully", updatedProfile)
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

	publicUser := map[string]interface{}{
		"id":        user.ID,
		"username":  user.Username,
		"full_name": user.FullName,
	}

	utils.WriteSuccessResponse(w, "user retrieved successfully", publicUser)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAllUsers(r.Context())
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	publicUsers := make([]map[string]interface{}, len(users))
	for i, user := range users {
		publicUsers[i] = map[string]interface{}{
			"id":        user.ID,
			"username":  user.Username,
			"full_name": user.FullName,
			"user_type": user.UserType,
		}
	}

	utils.WriteSuccessResponse(w, "users retrieved successfully", publicUsers)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := utils.ParseIDFromPath(r.URL.Path)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.userRepo.DeleteUser(userID); err != nil {
		if err == sql.ErrNoRows {
			utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
			return
		}
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	utils.WriteSuccessResponse(w, "user deleted successfully", nil)
}

func isValidTelegramUsername(username string) bool {
	if len(username) < 5 || len(username) > 32 {
		return false
	}

	//telegram usernames can only contain a-z, 0-9, and underscores
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '_') {
			return false
		}
	}

	return true
}
