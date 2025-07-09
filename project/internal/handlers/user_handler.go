package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

func (h *UserHandler) getCurrentUserID(r *http.Request) (int, error) {
	// todo: Extract user id from JWT token or session
	// for now returning a placeholder
	userIDStr := r.Header.Get("X-User-ID")
	if userIDStr == "" {
		return 0, fmt.Errorf("user not authenticated")
	}
	return strconv.Atoi(userIDStr)
}

func (h *UserHandler) validateUserType(r *http.Request, allowedTypes ...models.UserType) error {
	// todo: extract user type from JWT token or session
	// for now checking header
	userType := r.Header.Get("X-User-Type")
	if userType == "" {
		return fmt.Errorf("user type not found")
	}

	for _, allowedType := range allowedTypes {
		if models.UserType(userType) == allowedType {
			return nil
		}
	}
	return fmt.Errorf("insufficient permissions")
}

// resident
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

	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.Phone = req.Phone
	existingUser.FullName = req.FullName

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

	//parsign multipart form
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

	// validating file type
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

	// todo: Save file to minio
	// for no just reasd the file to validate it
	_, err = io.ReadAll(file)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to process file")
		return
	}

	// todo:update user profile with image URL
	fileName := fmt.Sprintf("profile_%d_%d%s", userID, time.Now().Unix(), fileExt)

	h.writeSuccessResponse(w, "profile picture uploaded successfully", map[string]string{
		"filename": fileName,
		"message":  "profile picture uploaded successfully",
	})
}

// manager
func (h *UserHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	if err := h.validateUserType(r, models.Manager); err != nil {
		h.writeErrorResponse(w, http.StatusForbidden, "manager access required")
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	//basic validation
	if req.Username == "" || req.Password == "" || req.Email == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "username, password, and email are required")
		return
	}

	// todo: hash password
	passwordHash := req.Password // this should be hashed babe

	user := models.User{
		Username:     req.Username,
		PasswordHash: passwordHash,
		Email:        req.Email,
		Phone:        req.Phone,
		FullName:     req.FullName,
		UserType:     req.UserType,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	userID, err := h.userRepo.CreateUser(ctx, user)
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	user.ID = userID
	h.writeSuccessResponse(w, "user created successfully", user)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if err := h.validateUserType(r, models.Manager); err != nil {
		h.writeErrorResponse(w, http.StatusForbidden, "manager access required")
		return
	}

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
	if err := h.validateUserType(r, models.Manager); err != nil {
		h.writeErrorResponse(w, http.StatusForbidden, "manager access required")
		return
	}
	users, err := h.userRepo.GetAllUsers(r.Context())
	if err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	h.writeSuccessResponse(w, "users retrieved successfully", users)
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if err := h.validateUserType(r, models.Manager); err != nil {
		h.writeErrorResponse(w, http.StatusForbidden, "manager access required")
		return
	}

	userID, err := h.parseIDFromPath(r.URL.Path)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid request body")
		return
	}

	existingUser, err := h.userRepo.GetUserByID(userID)
	if err != nil {
		h.writeErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	existingUser.Username = req.Username
	existingUser.Email = req.Email
	existingUser.Phone = req.Phone
	existingUser.FullName = req.FullName
	existingUser.UserType = req.UserType

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := h.userRepo.UpdateUser(ctx, *existingUser); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to update user")
		return
	}

	h.writeSuccessResponse(w, "user updated successfully", existingUser)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	if err := h.validateUserType(r, models.Manager); err != nil {
		h.writeErrorResponse(w, http.StatusForbidden, "manager access required")
		return
	}

	userID, err := h.parseIDFromPath(r.URL.Path)
	if err != nil {
		h.writeErrorResponse(w, http.StatusBadRequest, "invalid user id")
		return
	}

	if err := h.userRepo.DeleteUser(userID); err != nil {
		h.writeErrorResponse(w, http.StatusInternalServerError, "failed to delete user")
		return
	}

	h.writeSuccessResponse(w, "user deleted successfully", nil)
}
