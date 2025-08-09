package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
)

type ApartmentHandler struct {
	apartmentService services.ApartmentService
}

func NewApartmentHandler(apartmentService services.ApartmentService) *ApartmentHandler {
	return &ApartmentHandler{
		apartmentService: apartmentService,
	}
}
func (h *ApartmentHandler) CreateApartment(w http.ResponseWriter, r *http.Request) {
	var request struct {
		ApartmentName string `json:"apartment_name"`
		Address       string `json:"address"`
		UnitsCount    int    `json:"units_count"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ApartmentName == "" || request.Address == "" || request.UnitsCount <= 0 {
		http.Error(w, "All fields are required and units count must be positive", http.StatusBadRequest)
		return
	}

	userIDString, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	userID, _ := strconv.Atoi(userIDString)

	id, err := h.apartmentService.CreateApartment(r.Context(), userID, request.ApartmentName, request.Address, request.UnitsCount)
	if err != nil {
		http.Error(w, "Failed to create apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := map[string]int{"id": id}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ApartmentHandler) GetApartmentByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}
	managerId, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	apartment, err := h.apartmentService.GetApartmentByID(r.Context(), id, managerId)

	if err != nil {
		http.Error(w, "Apartment not found"+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(apartment)
}

func (h *ApartmentHandler) GetResidentsInApartment(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.PathValue("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}
	managerId, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	residents, err := h.apartmentService.GetResidentsInApartment(r.Context(), apartmentID, managerId)
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
	residentIDStr := r.PathValue("user_id")
	residentID, err := strconv.Atoi(residentIDStr)
	if err != nil {
		utils.WriteErrorResponse(w, http.StatusBadRequest, "invalid user ID")
		return
	}

	apartments, err := h.apartmentService.GetAllApartmentsForResident(r.Context(), residentID)
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
	var request struct {
		ID            int    `json:"id"`
		ApartmentName string `json:"apartment_name"`
		Address       string `json:"address"`
		UnitsCount    int    `json:"units_count"`
		ManagerID     int    `json:"manager_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if request.ID == 0 || request.ApartmentName == "" || request.Address == "" || request.UnitsCount == 0 || request.ManagerID == 0 {
		http.Error(w, "All fields are required", http.StatusBadRequest)
		return
	}

	if err := h.apartmentService.UpdateApartment(r.Context(), request.ID, request.ApartmentName, request.Address, request.UnitsCount, request.ManagerID); err != nil {
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
	managerId, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	if err := h.apartmentService.DeleteApartment(r.Context(), id, managerId); err != nil {
		http.Error(w, "Failed to delete apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *ApartmentHandler) InviteUserToApartment(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.PathValue("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	telegramUsername := r.PathValue("telegram_username")
	if telegramUsername == "" {
		http.Error(w, "Telegram username is required", http.StatusBadRequest)
		return
	}

	userIDString, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	managerID, _ := strconv.Atoi(userIDString)

	response, err := h.apartmentService.InviteUserToApartment(r.Context(), managerID, apartmentID, telegramUsername)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *ApartmentHandler) JoinApartment(w http.ResponseWriter, r *http.Request) {
	invitationCode := r.PathValue("invitation_code")

	userIDString, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	userID, _ := strconv.Atoi(userIDString)

	response, err := h.apartmentService.JoinApartment(r.Context(), userID, invitationCode)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *ApartmentHandler) LeaveApartment(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.URL.Query().Get("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	userIDString, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok {
		http.Error(w, "Failed to get user ID from context", http.StatusInternalServerError)
		return
	}
	userID, _ := strconv.Atoi(userIDString)

	if err := h.apartmentService.LeaveApartment(r.Context(), userID, apartmentID); err != nil {
		http.Error(w, "Failed to leave apartment: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "left apartment"})
}
