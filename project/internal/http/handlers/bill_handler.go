package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/payment"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type BillHandler struct {
	repo                repositories.BillRepository
	userRepo            repositories.UserRepository
	apartmentRepo       repositories.ApartmentRepository
	userApartmentRepo   repositories.UserApartmentRepository
	paymentRepo         repositories.PaymentRepository
	imageService        image.Image
	paymentService      payment.Payment
	notificationService notification.Notification
}

func NewBillHandler(
	repo repositories.BillRepository,
	userRepo repositories.UserRepository,
	apartmentRepo repositories.ApartmentRepository,
	userApartmentRepo repositories.UserApartmentRepository,
	paymentRepo repositories.PaymentRepository,
	imageService image.Image,
	paymentService payment.Payment,
	notificationService notification.Notification,
) *BillHandler {
	return &BillHandler{
		repo:                repo,
		userRepo:            userRepo,
		apartmentRepo:       apartmentRepo,
		userApartmentRepo:   userApartmentRepo,
		paymentRepo:         paymentRepo,
		imageService:        imageService,
		paymentService:      paymentService,
		notificationService: notificationService,
	}
}

type CreateBillRequest struct {
	BillType        models.BillType `json:"bill_type"`
	TotalAmount     float64         `json:"total_amount"`
	DueDate         string          `json:"due_date"`
	BillingDeadline string          `json:"billing_deadline"`
	Description     string          `json:"description"`
}

type PayBillsRequest struct {
	BillIDs []int `json:"bill_ids"`
}

type BatchPaymentRequest struct {
	UserID  int     `json:"user_id"`
	BillIDs []int   `json:"bill_ids"`
	Amount  float64 `json:"total_amount"`
}

func (h *BillHandler) CreateBill(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.PathValue("apartment-id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "invalid apartment ID", http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "failed to get user ID from context", http.StatusInternalServerError)
		return
	}

	isManager, err := h.userApartmentRepo.IsUserManagerOfApartment(r.Context(), userID, apartmentID)
	if err != nil {
		http.Error(w, "failed to verify manager status", http.StatusInternalServerError)
		return
	}
	if !isManager {
		http.Error(w, "only apartment managers can create bills", http.StatusForbidden)
		return
	}

	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "failed to parse form data", http.StatusBadRequest)
		return
	}

	var req CreateBillRequest
	req.BillType = models.BillType(r.FormValue("bill_type"))
	req.TotalAmount, _ = strconv.ParseFloat(r.FormValue("total_amount"), 64)
	req.DueDate = r.FormValue("due_date")
	req.BillingDeadline = r.FormValue("billing_deadline")
	req.Description = r.FormValue("description")

	if req.BillType == "" || req.TotalAmount <= 0 || req.DueDate == "" {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	//get the actual residents in that apartment count
	residents, err := h.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		http.Error(w, "failed to get residents", http.StatusInternalServerError)
		return
	}

	if len(residents) == 0 {
		http.Error(w, "no residents found in apartment", http.StatusBadRequest)
		return
	}

	//bill devision
	amountPerResident := req.TotalAmount / float64(len(residents))

	var imageURL string
	file, handler, err := r.FormFile("bill_image")
	if err == nil {
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "failed to read file", http.StatusInternalServerError)
			return
		}

		imageURL, err = h.imageService.SaveImage(r.Context(), fileBytes, handler.Filename)
		if err != nil {
			http.Error(w, "failed to save image", http.StatusInternalServerError)
			return
		}
	}

	bill := models.Bill{
		ApartmentID:     apartmentID,
		BillType:        req.BillType,
		TotalAmount:     req.TotalAmount,
		DueDate:         req.DueDate,
		BillingDeadline: req.BillingDeadline,
		Description:     req.Description,
		ImageURL:        imageURL,
	}

	billID, err := h.repo.CreateBill(r.Context(), bill)
	if err != nil {
		http.Error(w, "Failed to create bill", http.StatusInternalServerError)
		return
	}

	bill.ID = billID

	for _, resident := range residents {
		payment := models.Payment{
			BillID:        billID,
			UserID:        resident.ID,
			Amount:        fmt.Sprintf("%.2f", amountPerResident),
			PaymentStatus: models.Pending,
		}
		_, err := h.paymentRepo.CreatePayment(r.Context(), payment)
		if err != nil {
			http.Error(w, "failed to create payment record", http.StatusInternalServerError)
			return
		}

		//send notif
		if err := h.notificationService.SendBillNotification(r.Context(), resident.ID, bill, amountPerResident); err != nil {
			log.Printf("failed to send notification to user %d: %v", resident.ID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":                billID,
		"residents_count":   len(residents),
		"amount_per_person": amountPerResident,
	})
}

func (h *BillHandler) GetBillByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid bill ID", http.StatusBadRequest)
		return
	}

	bill, err := h.repo.GetBillByID(id)
	if err != nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bill)
}

func (h *BillHandler) GetBillsByApartment(w http.ResponseWriter, r *http.Request) {
	apartmentIDStr := r.URL.Query().Get("apartment_id")
	apartmentID, err := strconv.Atoi(apartmentIDStr)
	if err != nil {
		http.Error(w, "Invalid apartment ID", http.StatusBadRequest)
		return
	}

	bills, err := h.repo.GetBillsByApartmentID(apartmentID)
	if err != nil {
		http.Error(w, "Failed to get bills", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bills)
}

func (h *BillHandler) UpdateBill(w http.ResponseWriter, r *http.Request) {
	var bill models.Bill
	if err := json.NewDecoder(r.Body).Decode(&bill); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateBill(r.Context(), bill); err != nil {
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

	h.repo.DeleteBill(id)
	w.WriteHeader(http.StatusOK)
}

func (h *BillHandler) PayBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	var req PayBillsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//mock payment processing
	if err := h.paymentService.PayBills(req.BillIDs); err != nil {
		http.Error(w, "Payment failed", http.StatusInternalServerError)
		return
	}

	var payments []models.Payment
	for _, billID := range req.BillIDs {
		payments = append(payments, models.Payment{
			BillID:        billID,
			UserID:        userID,
			PaidAt:        time.Now(),
			PaymentStatus: models.Paid,
		})
	}

	h.paymentRepo.UpdatePaymentsStatus(r.Context(), payments)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "payment successful"})
}

func (h *BillHandler) PayBatchBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	var req BatchPaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	//validating that all bills belong to user and are unpaid
	var totalAmount float64
	var validBills []int

	for _, billID := range req.BillIDs {
		payment, err := h.paymentRepo.GetPaymentByBillAndUser(billID, userID)
		if err != nil {
			continue //invalid bills
		}

		if payment.PaymentStatus == models.Pending {
			amount, _ := strconv.ParseFloat(payment.Amount, 64)
			totalAmount += amount
			validBills = append(validBills, billID)
		}
	}

	if len(validBills) == 0 {
		http.Error(w, "No valid unpaid bills found", http.StatusBadRequest)
		return
	}

	//batch payment
	if err := h.paymentService.PayBills(validBills); err != nil {
		http.Error(w, "Batch payment failed", http.StatusInternalServerError)
		return
	}

	var payments []models.Payment
	for _, billID := range validBills {
		payments = append(payments, models.Payment{
			BillID:        billID,
			UserID:        userID,
			PaidAt:        time.Now(),
			PaymentStatus: models.Paid,
		})
	}

	h.paymentRepo.UpdatePaymentsStatus(r.Context(), payments)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":       "batch payment successful",
		"bills_paid":   len(validBills),
		"total_amount": totalAmount,
	})
}

func (h *BillHandler) GetUnpaidBills(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("user_id").(int)
	if !ok {
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	apartments, err := h.userApartmentRepo.GetAllApartmentsForAResident(userID)
	if err != nil {
		http.Error(w, "Failed to get apartments", http.StatusInternalServerError)
		return
	}

	var unpaidBills []models.Bill
	for _, apartment := range apartments {
		bills, err := h.repo.GetBillsByApartmentID(apartment.ID)
		if err != nil {
			continue
		}

		for _, bill := range bills {
			//if payment exists and is unpaid
			payment, err := h.paymentRepo.GetPaymentByBillAndUser(bill.ID, userID)
			if err == nil && payment.PaymentStatus == models.Pending {
				unpaidBills = append(unpaidBills, bill)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(unpaidBills)
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

	bill, err := h.repo.GetBillByID(billID)
	if err != nil {
		http.Error(w, "Bill not found", http.StatusNotFound)
		return
	}

	payment, err := h.paymentRepo.GetPaymentByBillAndUser(billID, userID)
	if err != nil {
		http.Error(w, "Payment record not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"bill":           bill,
		"payment_status": payment.PaymentStatus,
		"amount_due":     payment.Amount,
		"paid_at":        payment.PaidAt,
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

	apartments, err := h.userApartmentRepo.GetAllApartmentsForAResident(userID)
	if err != nil {
		http.Error(w, "Failed to get apartments", http.StatusInternalServerError)
		return
	}

	type PaymentHistoryItem struct {
		Bill          models.Bill    `json:"bill"`
		Payment       models.Payment `json:"payment"`
		ApartmentName string         `json:"apartment_name"`
	}

	var history []PaymentHistoryItem

	for _, apartment := range apartments {
		bills, err := h.repo.GetBillsByApartmentID(apartment.ID)
		if err != nil {
			continue
		}

		for _, bill := range bills {
			payment, err := h.paymentRepo.GetPaymentByBillAndUser(bill.ID, userID)
			if err == nil {
				history = append(history, PaymentHistoryItem{
					Bill:          bill,
					Payment:       *payment,
					ApartmentName: apartment.ApartmentName,
				})
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}
