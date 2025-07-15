package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type ApartmentHandler struct {
	apartmentRepo       repositories.ApartmentRepository
	userApartmentRepo   repositories.UserApartmentRepository
	inviteLinkRepo      repositories.InviteLinkFlagRepo
	notificationService notification.Notification
}

func NewApartmentHandler(apartmentRepo repositories.ApartmentRepository, userApartmentRepo repositories.UserApartmentRepository, inviteLinkRepo repositories.InviteLinkFlagRepo, notificationService notification.Notification) *ApartmentHandler {
	return &ApartmentHandler{
		apartmentRepo:       apartmentRepo,
		userApartmentRepo:   userApartmentRepo,
		inviteLinkRepo:      inviteLinkRepo,
		notificationService: notificationService,
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
	//with the user id that this request was sent from and the apartment id in the body of req
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
	//getting apartment ID from path parameters
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
	//extracting userID from URL
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
	vars := mux.Vars(r)
	apartmentID, err := strconv.Atoi(vars["apartment-id"])
	if err != nil {
		http.Error(w, "invalid apartment ID", http.StatusBadRequest)
		return
	}

	telegramUsername := vars["telegram-username"]
	if telegramUsername == "" {
		http.Error(w, "telegram username is required", http.StatusBadRequest)
		return
	}

	//invitation token
	token, err := generateToken()
	if err != nil {
		http.Error(w, "failed to generate invitation token", http.StatusInternalServerError)
		return
	}

	//storing invitation in redis
	err = h.inviteLinkRepo.Set(r.Context(), telegramUsername, strconv.Itoa(apartmentID), token)
	if err != nil {
		http.Error(w, "failed to store invitation", http.StatusInternalServerError)
		return
	}

	//send a telegram notification
	err = h.notificationService.SendInvitation(r.Context(), telegramUsername, apartmentID, token)
	if err != nil {
		http.Error(w, "failed to send invitation", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "invitation sent"})
}

func (h *ApartmentHandler) JoinApartment(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		http.Error(w, "token is required", http.StatusBadRequest)
		return
	}

	//getting current user from context
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "failed to get user id from context", http.StatusInternalServerError)
		return
	}

	// Verify token and get apartment ID
	apartmentID, err := h.inviteLinkRepo.VerifyToken(r.Context(), token)
	if err != nil {
		http.Error(w, "invalid or expired token", http.StatusBadRequest)
		return
	}

	//add in user_apartment
	userApartment := models.User_apartment{
		UserID:      userID,
		ApartmentID: apartmentID,
		IsManager:   false,
	}

	if err := h.userApartmentRepo.CreateUserApartment(r.Context(), userApartment); err != nil {
		http.Error(w, "failed to join apartment", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "joined apartment"})
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
