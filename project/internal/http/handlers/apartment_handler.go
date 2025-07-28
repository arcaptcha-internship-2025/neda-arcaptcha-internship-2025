package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type ApartmentHandler struct {
	apartmentRepo       repositories.ApartmentRepository
	userRepo            repositories.UserRepository
	userApartmentRepo   repositories.UserApartmentRepository
	inviteLinkRepo      repositories.InviteLinkFlagRepo
	notificationService notification.Notification
	appBaseURL          string // base url for generating invite links
}

func NewApartmentHandler(
	apartmentRepo repositories.ApartmentRepository,
	userRepo repositories.UserRepository,
	userApartmentRepo repositories.UserApartmentRepository,
	inviteLinkRepo repositories.InviteLinkFlagRepo,
	notificationService notification.Notification,
	appBaseURL string,
) *ApartmentHandler {
	return &ApartmentHandler{
		apartmentRepo:       apartmentRepo,
		userRepo:            userRepo,
		userApartmentRepo:   userApartmentRepo,
		inviteLinkRepo:      inviteLinkRepo,
		notificationService: notificationService,
		appBaseURL:          appBaseURL,
	}
}

func (h *ApartmentHandler) CreateApartment(w http.ResponseWriter, r *http.Request) {
	var apartment models.Apartment
	if err := json.NewDecoder(r.Body).Decode(&apartment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	id, err := h.apartmentRepo.CreateApartment(r.Context(), apartment)
	if err != nil {
		http.Error(w, "failed to create apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	//manager should be added to user-apartment repo
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: id,
		IsManager:   true,
	}

	if err := h.userApartmentRepo.CreateUserApartment(r.Context(), userApartment); err != nil {
		http.Error(w, "failed to assign manager to apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *ApartmentHandler) GetApartmentByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	apartment, err := h.apartmentRepo.GetApartmentByID(id)
	if err != nil {
		http.Error(w, "Apartment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apartment)
}

func (h *ApartmentHandler) GetResidentsInApartment(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	apartmentID, err := strconv.Atoi(vars["apartment-id"])
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	residents, err := h.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		http.Error(w, "Failed to get residents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(residents); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ApartmentHandler) GetAllApartmentsForResident(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	residentID, err := strconv.Atoi(vars["userID"])
	if err != nil {
		http.Error(w, "Invalid user ID", http.StatusBadRequest)
		return
	}

	apartments, err := h.userApartmentRepo.GetAllApartmentsForAResident(residentID)
	if err != nil {
		http.Error(w, "Failed to get apartments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(apartments); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (h *ApartmentHandler) UpdateApartment(w http.ResponseWriter, r *http.Request) {
	var apartment models.Apartment
	if err := json.NewDecoder(r.Body).Decode(&apartment); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.apartmentRepo.UpdateApartment(r.Context(), apartment); err != nil {
		http.Error(w, "Failed to update apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ApartmentHandler) DeleteApartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	if err := h.apartmentRepo.DeleteApartment(id); err != nil {
		http.Error(w, "Failed to delete apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ApartmentHandler) InviteUserToApartment(w http.ResponseWriter, r *http.Request) {
	//getting manager id from context
	managerID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	//verifying the inviting user is a manager of the apartment
	vars := mux.Vars(r)
	apartmentID, err := strconv.Atoi(vars["apartment-id"])
	if err != nil {
		http.Error(w, "invalid apartment ID", http.StatusBadRequest)
		return
	}

	//checking if the user is actually a manager of this apartment
	isManager, err := h.userApartmentRepo.IsUserManagerOfApartment(r.Context(), managerID, apartmentID)
	if err != nil {
		http.Error(w, "failed to verify apartment manager status", http.StatusInternalServerError)
		return
	}
	if !isManager {
		http.Error(w, "only apartment managers can send invitations", http.StatusForbidden)
		return
	}

	telegramUsername := vars["telegram-username"]
	if telegramUsername == "" {
		http.Error(w, "telegram username is required", http.StatusBadRequest)
		return
	}

	receiver, err := h.userRepo.GetUserByTelegramUser(telegramUsername)
	if err != nil {
		http.Error(w, "user with this Telegram username not found", http.StatusNotFound)
		return
	}

	isResident, err := h.userApartmentRepo.IsUserInApartment(r.Context(), receiver.ID, apartmentID)
	if err != nil {
		http.Error(w, "failed to check resident status", http.StatusInternalServerError)
		return
	}
	if isResident {
		http.Error(w, "user is already a resident of this apartment", http.StatusConflict)
		return
	}

	token, err := generateToken()
	if err != nil {
		http.Error(w, "failed to generate invitation token", http.StatusInternalServerError)
		return
	}

	//apartment details for the invitation
	apartment, err := h.apartmentRepo.GetApartmentByID(apartmentID)
	if err != nil {
		http.Error(w, "failed to get apartment details", http.StatusInternalServerError)
		return
	}

	//manager details for the invitation
	manager, err := h.userRepo.GetUserByID(managerID)
	if err != nil {
		http.Error(w, "failed to get manager details", http.StatusInternalServerError)
		return
	}

	inviteURL := fmt.Sprintf("%s/apartment/join?token=%s", h.appBaseURL, token)

	invitation := models.InvitationLink{
		SenderID:         managerID,
		SenderUsername:   manager.Username,
		ReceiverUsername: telegramUsername,
		ApartmentID:      apartmentID,
		ApartmentName:    apartment.ApartmentName,
		Token:            token,
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		Status:           models.InvitationStatusPending,
		InviteURL:        inviteURL,
	}

	//storing invitation
	if err := h.inviteLinkRepo.CreateInvitation(r.Context(), invitation); err != nil {
		http.Error(w, "failed to store invitation", http.StatusInternalServerError)
		return
	}

	//sending notification via Telegram
	if err := h.notificationService.SendInvitation(r.Context(), invitation); err != nil {
		http.Error(w, "invitation created but failed to send notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "invitation sent",
		"token":      token,
		"invite_url": inviteURL,
		"expires_at": invitation.ExpiresAt,
	})
}

func (h *ApartmentHandler) JoinApartment(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	//getting invitation by token
	inv, err := h.inviteLinkRepo.GetInvitationByToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
		return
	}

	if inv.Status != models.InvitationStatusPending && inv.Status != models.InvitationStatusNotified {
		http.Error(w, "invitation is no longer valid", http.StatusBadRequest)
		return
	}

	isResident, err := h.userApartmentRepo.IsUserInApartment(r.Context(), userID, inv.ApartmentID)
	if err != nil {
		http.Error(w, "failed to check resident status", http.StatusInternalServerError)
		return
	}
	if isResident {
		http.Error(w, "you are already a resident of this apartment", http.StatusConflict)
		return
	}

	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: inv.ApartmentID,
		IsManager:   false,
	}

	if err := h.userApartmentRepo.CreateUserApartment(r.Context(), userApartment); err != nil {
		http.Error(w, "failed to join apartment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "joined apartment",
		"apartment_id": inv.ApartmentID,
		"apartment":    inv.ApartmentName,
	})
}

func (h *ApartmentHandler) LeaveApartment(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.URL.Query().Get("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	err = h.userApartmentRepo.DeleteUserApartment(userID, apartmentID)
	if err != nil {
		http.Error(w, "Failed to leave apartment", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "left apartment"})
}

func generateToken() (string, error) {
	bytes := make([]byte, 32) // 256-bit token
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
