package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
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
	apartmentIDStr := r.PathValue("apartmentId")
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
		ImageURL        string  `json:"image_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.billService.UpdateBill(r.Context(), req.ID, req.ApartmentID, req.BillType, req.TotalAmount, req.DueDate, req.BillingDeadline, req.Description, req.ImageURL); err != nil {
		http.Error(w, "Failed to update bill: "+err.Error(), http.StatusInternalServerError)
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
		http.Error(w, "Failed to delete bill: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BillHandler) PayBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	var req dto.PayBillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.billService.PayBills(r.Context(), userID, req.BillIDs); err != nil {
		http.Error(w, "Payment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "payment successful"})
}

func (h *BillHandler) PayBatchBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	var req dto.BatchPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	response, err := h.billService.PayBatchBills(r.Context(), userID, req.BillIDs)
	if err != nil {
		http.Error(w, "Batch payment failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) GetUnpaidBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	bills, err := h.billService.GetUnpaidBills(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get unpaid bills: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bills)
}

func (h *BillHandler) GetBillWithPaymentStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	billIDStr := r.URL.Query().Get("bill_id")
	billID, err := strconv.Atoi(billIDStr)
	if err != nil {
		http.Error(w, "Invalid bill ID", http.StatusBadRequest)
		return
	}

	response, err := h.billService.GetBillWithPaymentStatus(r.Context(), userID, billID)
	if err != nil {
		http.Error(w, "Failed to get bill with payment status: "+err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *BillHandler) GetUserPaymentHistory(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	history, err := h.billService.GetUserPaymentHistory(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to get payment history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
