package handlers

import (
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

func (h *ApartmentHandler) GetResidentsInApartment(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	residents, err := h.userApartmentRepo.GetResidentsInApartment(id)
	if err != nil {
		http.Error(w, "Failed to get residents: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(residents)
}

func (h *ApartmentHandler) InviteUserToApartment(w http.ResponseWriter, r *http.Request) {
	panic("InviteUserToApartment not implemented yet")
}

func (h *ApartmentHandler) JoinApartment(w http.ResponseWriter, r *http.Request) {
	panic("InviteUserToApartment not implemented yet")

}

func (h *ApartmentHandler) LeaveApartment(w http.ResponseWriter, r *http.Request) {
	panic("InviteUserToApartment not implemented yet")

}
