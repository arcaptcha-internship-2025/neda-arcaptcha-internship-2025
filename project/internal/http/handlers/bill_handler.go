package handlers

import (
	"encoding/json"
	"fmt"
	"io"
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
	ApartmentID     int             `json:"apartment_id"`
	BillType        models.BillType `json:"bill_type"`
	TotalAmount     float64         `json:"total_amount"`
	DueDate         string          `json:"due_date"`
	BillingDeadline string          `json:"billing_deadline"`
	Description     string          `json:"description"`
}

type PayBillsRequest struct {
	BillIDs []int `json:"bill_ids"`
}

func (h *BillHandler) CreateBill(w http.ResponseWriter, r *http.Request) {
	//parse form data (for file upload)
	err := r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		http.Error(w, "Failed to parse form data", http.StatusBadRequest)
		return
	}

	var req CreateBillRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	apartment, err := h.apartmentRepo.GetApartmentByID(req.ApartmentID)
	if err != nil {
		http.Error(w, "Apartment not found", http.StatusNotFound)
		return
	}

	amountPerUnit := req.TotalAmount / float64(apartment.UnitsCount)

	//handling file upload if exists
	var imageURL string
	file, handler, err := r.FormFile("bill_image")
	if err == nil {
		defer file.Close()

		fileBytes, err := io.ReadAll(file)
		if err != nil {
			http.Error(w, "Failed to read file", http.StatusInternalServerError)
			return
		}

		imageURL, err = h.imageService.SaveImage(r.Context(), fileBytes, handler.Filename)
		if err != nil {
			http.Error(w, "Failed to save image", http.StatusInternalServerError)
			return
		}
	}

	bill := models.Bill{
		ApartmentID:     req.ApartmentID,
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

	residents, err := h.userApartmentRepo.GetResidentsInApartment(req.ApartmentID)
	if err != nil {
		http.Error(w, "Failed to get residents", http.StatusInternalServerError)
		return
	}

	for _, resident := range residents {
		payment := models.Payment{
			BillID:        billID,
			UserID:        resident.ID,
			Amount:        fmt.Sprintf("%.2f", amountPerUnit),
			PaymentStatus: models.Pending,
		}
		_, err := h.paymentRepo.CreatePayment(r.Context(), payment)
		if err != nil {
			http.Error(w, "Failed to create payment record", http.StatusInternalServerError)
			return
		}

		//sending notification to resident
		message := fmt.Sprintf(
			"New bill created:\nType: %s\nAmount: %.2f\nDue Date: %s\nDescription: %s",
			bill.BillType, amountPerUnit, bill.DueDate, bill.Description,
		)
		if bill.ImageURL != "" {
			message += "\nBill image available in your dashboard"
		}

		_ = h.notificationService.SendNotification(r.Context(), resident.ID, message)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]int{"id": billID})
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

	//updating payment statuses
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
