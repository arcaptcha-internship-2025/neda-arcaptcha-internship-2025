package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025.git/internal/repositories"
	"golang.org/x/crypto/bcrypt"
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

func (h *UserHandler) writeJSONResponse(w http.ResponseWriter, statusCode int, response APIResponse) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

func (h *UserHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, message string) {
	h.writeJSONResponse(w, statusCode, APIResponse{
		Success: false,
		Message: message,
		Error:   message,
	})
}

func (h *UserHandler) writeSuccessResponse(w http.ResponseWriter, message string, data interface{}) {
	h.writeJSONResponse(w, http.StatusOK, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func (h *UserHandler) parseIDFromPath(path string) (int, error) {
	parts := strings.Split(strings.TrimSuffix(path, "/"), "/")
	if len(parts) == 0 {
		return 0, fmt.Errorf("invalid path")
	}

	idStr := parts[len(parts)-1]
	return strconv.Atoi(idStr)
}

// extract user id from middleware
func (h *UserHandler) getCurrentUserID(r *http.Request) (int, error) {
	userIDValue := r.Context().Value(middleware.UserIDKey)
	if userIDValue == nil {
		return 0, fmt.Errorf("user ID not found in context")
	}

	userIDStr, ok := userIDValue.(string)
	if !ok {
		return 0, fmt.Errorf("invalid user ID format in context")
	}

	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		return 0, fmt.Errorf("invalid user ID format: %v", err)
	}

	return userID, nil
}

func (h *UserHandler) SignUp(w http.ResponseWriter, r *http.Request) {
	var req models.User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" || req.Email == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "username, password, and email are required")
		return
	}

	if req.UserType != models.Manager && req.UserType != models.Resident {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid user type")
		return
	}

	existingUser, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil && err != sql.ErrNoRows {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to check existing user")
		return
	}
	if existingUser != nil {
		h.writeErrorResponse(w, http.StatusConflict, "username already exists")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to hash password")
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
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create user")
		return
	}
	user.ID = userID

	// generating token for immediate login after signup
	token, err := middleware.GenerateToken(strconv.Itoa(userID), user.UserType)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	responseData := map[string]interface{}{
		"user":  user,
		"token": token,
	}

	h.writeSuccessResponse(w, "user created successfully", responseData)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Username == "" || req.Password == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "username and password are required")
		return
	}

	existingUser, err := h.userRepo.GetUserByUsername(req.Username)
	if err != nil {
		if err == sql.ErrNoRows {
			h.writeErrorResponse(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to retrieve user")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(existingUser.Password), []byte(req.Password)); err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "invalid username or password")
		return
	}

	token, err := middleware.GenerateToken(strconv.Itoa(existingUser.ID), existingUser.UserType)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to generate token")
		return
	}

	responseData := map[string]interface{}{
		"token":     token,
		"user_id":   strconv.Itoa(existingUser.ID),
		"user_type": string(existingUser.UserType),
		"username":  existingUser.Username,
		"email":     existingUser.Email,
		"full_name": existingUser.FullName,
	}

	h.writeSuccessResponse(w, "login successful", responseData)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	h.writeSuccessResponse(w, "profile retrieved successfully", user)
}

func (h *UserHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existingUser, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "user not found")
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
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update profile")
		return
	}

	h.writeSuccessResponse(w, "profile updated successfully", existingUser)
}

func (h *UserHandler) UploadProfilePicture(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		h.writeErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}
	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "failed to parse form")
		return
	}

	file, header, err := r.FormFile("profile_picture")
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "no file uploaded")
		return
	}
	defer file.Close()

	allowedTypes := []string{".jpg", ".jpeg", ".png", ".gif"}
	fileExt := strings.ToLower(filepath.Ext(header.Filename))
	isValidType := false
	for _, allowedType := range allowedTypes {
		if fileExt == allowedType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid file type. only JPG, PNG, and GIF are allowed")
		return
	}

	// todo: save file to minio
	// for now just read the file to validate it
	_, err = io.ReadAll(file)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	// todo: update user profile with image URL
	fileName := fmt.Sprintf("profile_%d_%d%s", userID, time.Now().Unix(), fileExt)

	h.writeSuccessResponse(w, "profile picture uploaded successfully", map[string]string{
		"filename": fileName,
		"message":  "profile picture uploaded successfully",
	})
}

// manager-only
func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	// auth is handled by middleware
	userID, err := h.parseIDFromPath(r.URL.Path)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	user, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	h.writeSuccessResponse(w, "user retrieved successfully", user)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.userRepo.GetAllUsers(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	h.writeSuccessResponse(w, "users retrieved successfully", users)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.parseIDFromPath(r.URL.Path)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.userRepo.DeleteUser(userID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	h.writeSuccessResponse(w, "user deleted successfully", nil)
}
