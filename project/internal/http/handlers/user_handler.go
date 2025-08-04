package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
	log "github.com/sirupsen/logrus"
)

type UserHandler struct {
	userService services.UserService
	botAddress  string
}

func NewUserHandler(userService services.UserService, botAddress string) *UserHandler {
	return &UserHandler{
		userService: userService,
		botAddress:  botAddress,
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

	//calling service
	response, err := h.userService.CreateUser(r.Context(), req, h.botAddress)
	if err != nil {
		log.WithError(err).Error("failed to create user")
		switch err.Error() {
		case "invalid user type":
			utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user type")
		case "invalid Telegram username format":
			utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid Telegram username format")
		case "username already exists":
			utils.WriteErrorResponse(w, http.StatusConflict, "username already exists")
		case "Telegram username already in use":
			utils.WriteErrorResponse(w, http.StatusConflict, "Telegram username already in use")
		default:
			if err.Error() == "failed to check existing user" ||
				err.Error() == "failed to check Telegram username" ||
				err.Error() == "failed to hash password" ||
				err.Error() == "failed to create user" ||
				err.Error() == "failed to generate token" {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, err.Error())
			} else {
				utils.WriteErrorResponse(w, http.StatusInternalServerError, "internal server error")
			}
			return
		}
	}

	utils.WriteSuccessResponse(w, "user created successfully", response)
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

	response, err := h.userService.AuthenticateUser(r.Context(), req)
	if err != nil {
		if err.Error() == "invalid username or password" {
			utils.WriteErrorResponse(w, http.StatusUnauthorized, "invalid username or password")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to authenticate user")
		}
		log.WithError(err).Error("failed to authenticate user")
		return
	}

	utils.WriteSuccessResponse(w, "login successful", response)
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getCurrentUserID(r)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusUnauthorized, "authentication required")
		return
	}

	response, err := h.userService.GetUserProfile(r.Context(), userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	utils.WriteSuccessResponse(w, "profile retrieved successfully", response)
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

	response, err := h.userService.UpdateUserProfile(r.Context(), userID, req)
	if err != nil {
		switch err.Error() {
		case "user not found":
			utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		case "invalid Telegram username format":
			utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid Telegram username format")
		case "Telegram username already in use":
			utils.WriteErrorResponse(w, http.StatusConflict, "Telegram username already in use")
		default:
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to update profile")
		}
		return
	}

	utils.WriteSuccessResponse(w, "profile updated successfully", response)
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	response, err := h.userService.GetPublicUser(r.Context(), userID)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		return
	}

	utils.WriteSuccessResponse(w, "user retrieved successfully", response)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	response, err := h.userService.GetAllPublicUsers(r.Context())
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to retrieve users")
		return
	}

	utils.WriteSuccessResponse(w, "users retrieved successfully", response)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("user_id")
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	if err := h.userService.DeleteUser(r.Context(), userID); err != nil {
		if err.Error() == "user not found" {
			utils.WriteErrorResponse(w, http.StatusNotFound, "user not found")
		} else {
			utils.WriteErrorResponse(w, http.StatusInternalServerError, "failed to delete user")
		}
		return
	}

	utils.WriteSuccessResponse(w, "user deleted successfully", nil)
}
