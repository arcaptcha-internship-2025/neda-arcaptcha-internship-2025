package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type ApartmentHandler struct {
	apartmentRepo     repositories.ApartmentRepository
	userApartmentRepo repositories.UserApartmentRepository
	inviteLinkRepo    repositories.InviteLinkFlagRepo
}

func NewApartmentHandler(apartmentRepo repositories.ApartmentRepository, userApartmentRepo repositories.UserApartmentRepository, inviteLinkRepo repositories.InviteLinkFlagRepo) *ApartmentHandler {
	return &ApartmentHandler{
		apartmentRepo:     apartmentRepo,
		userApartmentRepo: userApartmentRepo,
		inviteLinkRepo:    inviteLinkRepo,
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
		http.Error(w, "Failed to create apartment: "+err.Error(), http.StatusInternalServerError)
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

func (h *ApartmentHandler) GetAllApartments(w http.ResponseWriter, r *http.Request) {
	// Note: You'll need to add a GetAllApartments method to your repository interface
	apartments, err := h.apartmentRepo.GetAllApartments(r.Context())
	if err != nil {
		http.Error(w, "Failed to get apartments: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apartments)
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
