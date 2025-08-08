package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
	"github.com/sirupsen/logrus"
)

type BillHandler struct {
	billService services.BillService
}

func NewBillHandler(billService services.BillService) *BillHandler {
	return &BillHandler{
		billService: billService,
	}
}

func (h *BillHandler) CreateBill(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.PathValue("apartment_id")
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

	err = r.ParseMultipartForm(32 << 20) // 32 MB max
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	var req dto.CreateBillRequest
	req.BillType = models.BillType(r.FormValue("bill_type"))
	req.TotalAmount, _ = strconv.ParseFloat(r.FormValue("total_amount"), 64)
	req.DueDate = r.FormValue("due_date")
	req.BillingDeadline = r.FormValue("billing_deadline")
	req.Description = r.FormValue("description")

	file, handler, _ := r.FormFile("bill_image")

	response, err := h.billService.CreateBill(r.Context(), userID, apartmentID, req, file, handler)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) DivideBillByType(w http.ResponseWriter, r *http.Request) {
	billTypeStr := r.PathValue("bill_type")
	apartmentIDStr := r.PathValue("apartment_id")

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

	validBillTypes := map[string]models.BillType{
		"water":       models.WaterBill,
		"electricity": models.ElectricityBill,
		"gas":         models.GasBill,
		"maintenance": models.MaintenanceBill,
		"other":       models.OtherBill,
	}

	billType, valid := validBillTypes[billTypeStr]
	if !valid {
		http.Error(w, "Invalid bill type. Valid types: water, electricity, gas, maintenance, other", http.StatusBadRequest)
		return
	}

	response, err := h.billService.DivideBillByType(r.Context(), userID, apartmentID, billType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) DivideAllBills(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.PathValue("apartment_id")
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

	response, err := h.billService.DivideAllBills(r.Context(), userID, apartmentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) GetBillByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid bill ID", http.StatusBadRequest)
		return
	}

	response, err := h.billService.GetBillByID(r.Context(), id)
	if err != nil {
		http.Error(w, "Bill not found: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) GetBillsByApartment(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.URL.Query().Get("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	bills, err := h.billService.GetBillsByApartmentID(r.Context(), apartmentID)
	if err != nil {
		http.Error(w, "Failed to get bills: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bills)
}

func (h *BillHandler) UpdateBill(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID              int     `json:"id"`
		ApartmentID     int     `json:"apartment_id"`
		BillType        string  `json:"bill_type"`
		TotalAmount     float64 `json:"total_amount"`
		DueDate         string  `json:"due_date"`
		BillingDeadline string  `json:"billing_deadline"`
		Description     string  `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.billService.UpdateBill(r.Context(), req.ID, req.ApartmentID, req.BillType, req.TotalAmount, req.DueDate, req.BillingDeadline, req.Description); err != nil {
		http.Error(w, "Failed to update bill", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillHandler) DeleteBill(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid bill ID", http.StatusBadRequest)
		return
	}

	if err := h.billService.DeleteBill(r.Context(), id); err != nil {
		logrus.Error("Failed to delete bill:", err)
		http.Error(w, "Failed to delete bill", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *BillHandler) PayBill(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	path := r.URL.Path
	log.Printf("Full path: %s", path)

	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 3 {
		http.Error(w, "Invalid path format", http.StatusBadRequest)
		return
	}

	paymentStr := parts[len(parts)-1]
	log.Printf("Extracted payment_id: '%s'", paymentStr)

	paymentID, err := strconv.Atoi(paymentStr)
	if err != nil {
		log.Printf("Failed to convert payment_id '%s' to int: %v", paymentStr, err)
		http.Error(w, "Failed to get payment ID", http.StatusBadRequest)
		return
	}

	payments := []int{paymentID}

	if err := h.billService.PayBills(r.Context(), userID, payments, r.Context().Value(middleware.IdempotentKey).(string)); err != nil {
		http.Error(w, "Payment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "payment successful"})
}

func (h *BillHandler) PayBatchBills(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	response, err := h.billService.PayBatchBills(r.Context(), userID, r.Context().Value(middleware.IdempotentKey).(string))
	if err != nil {
		http.Error(w, "Batch payment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) GetUnpaidBills(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	bills, err := h.billService.GetUnpaidBills(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get unpaid bills: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bills)
}

func (h *BillHandler) GetUserPaymentHistory(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.Atoi(r.Context().Value(middleware.UserIDKey).(string))

	history, err := h.billService.GetUserPaymentHistory(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get payment history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
