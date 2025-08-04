package services

import (
	"context"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/payment"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
)

type BillService interface {
	CreateBill(ctx context.Context, userID, apartmentID int, req dto.CreateBillRequest, file io.ReadCloser, handler *multipart.FileHeader) (map[string]interface{}, error)
	GetBillByID(ctx context.Context, id int) (map[string]interface{}, error)
	GetBillsByApartmentID(ctx context.Context, apartmentID int) ([]models.Bill, error)
	UpdateBill(ctx context.Context, id, apartmentID int, billType string, totalAmount float64, dueDate, billingDeadline, description string) error
	DeleteBill(ctx context.Context, id int) error
	PayBills(ctx context.Context, userID int, billIDs []int) error
	PayBatchBills(ctx context.Context, userID int, billIDs []int) (map[string]interface{}, error)
	GetUnpaidBills(ctx context.Context, userID int) ([]models.Bill, error)
	GetBillWithPaymentStatus(ctx context.Context, userID, billID int) (map[string]interface{}, error)
	GetUserPaymentHistory(ctx context.Context, userID int) ([]PaymentHistoryItem, error)
}

type PaymentHistoryItem struct {
	Bill          models.Bill    `json:"bill"`
	Payment       models.Payment `json:"payment"`
	ApartmentName string         `json:"apartment_name"`
}

type billServiceImpl struct {
	repo                repositories.BillRepository
	userRepo            repositories.UserRepository
	apartmentRepo       repositories.ApartmentRepository
	userApartmentRepo   repositories.UserApartmentRepository
	paymentRepo         repositories.PaymentRepository
	imageService        image.Image
	paymentService      payment.Payment
	notificationService notification.Notification
}

func NewBillService(
	repo repositories.BillRepository,
	userRepo repositories.UserRepository,
	apartmentRepo repositories.ApartmentRepository,
	userApartmentRepo repositories.UserApartmentRepository,
	paymentRepo repositories.PaymentRepository,
	imageService image.Image,
	paymentService payment.Payment,
	notificationService notification.Notification,
) BillService {
	return &billServiceImpl{
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

func (s *billServiceImpl) CreateBill(ctx context.Context, userID, apartmentID int, req dto.CreateBillRequest, file io.ReadCloser, handler *multipart.FileHeader) (map[string]interface{}, error) {
	isManager, err := s.userApartmentRepo.IsUserManagerOfApartment(ctx, userID, apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify manager status: %w", err)
	}
	if !isManager {
		return nil, fmt.Errorf("only apartment managers can create bills")
	}

	if req.BillType == "" || req.TotalAmount <= 0 || req.DueDate == "" {
		return nil, fmt.Errorf("missing required fields")
	}

	validBillTypes := map[models.BillType]bool{
		models.WaterBill:       true,
		models.ElectricityBill: true,
		models.GasBill:         true,
		models.MaintenanceBill: true,
		models.OtherBill:       true,
	}
	if !validBillTypes[models.BillType(req.BillType)] {
		return nil, fmt.Errorf("invalid bill type")
	}

	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		return nil, fmt.Errorf("invalid due date format (use YYYY-MM-DD)")
	}
	if req.BillingDeadline != "" {
		if _, err := time.Parse("2006-01-02", req.BillingDeadline); err != nil {
			return nil, fmt.Errorf("invalid billing deadline format (use YYYY-MM-DD)")
		}
	}

	residents, err := s.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get residents: %w", err)
	}
	if len(residents) == 0 {
		return nil, fmt.Errorf("no residents found in apartment")
	}

	var imageKey string
	if file != nil {
		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		imageKey, err = s.imageService.SaveImage(ctx, fileBytes, handler.Filename)
		if err != nil {
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
	}

	bill := models.Bill{
		BaseModel: models.BaseModel{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		ApartmentID:     apartmentID,
		BillType:        models.BillType(req.BillType),
		TotalAmount:     req.TotalAmount,
		DueDate:         req.DueDate,
		BillingDeadline: req.BillingDeadline,
		Description:     req.Description,
		ImageURL:        imageKey,
	}

	billID, err := s.repo.CreateBill(ctx, bill)
	if err != nil {
		if imageKey != "" {
			if delErr := s.imageService.DeleteImage(ctx, imageKey); delErr != nil {
				log.Printf("Failed to cleanup uploaded image %s: %v", imageKey, delErr)
			}
		}
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	bill.ID = billID
	amountPerResident := req.TotalAmount / float64(len(residents))

	var failedPayments []int
	for _, resident := range residents {
		payment := models.Payment{
			BaseModel: models.BaseModel{
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			},
			BillID:        billID,
			UserID:        resident.ID,
			Amount:        fmt.Sprintf("%.2f", amountPerResident),
			PaymentStatus: models.Pending,
		}
		_, err := s.paymentRepo.CreatePayment(ctx, payment)
		if err != nil {
			log.Printf("Failed to create payment record for user %d: %v", resident.ID, err)
			failedPayments = append(failedPayments, resident.ID)
			continue
		}

		if err := s.notificationService.SendBillNotification(ctx, resident.ID, bill, amountPerResident); err != nil {
			log.Printf("Failed to send notification to user %d: %v", resident.ID, err)
		}
	}

	response := map[string]interface{}{
		"id":                billID,
		"residents_count":   len(residents),
		"amount_per_person": amountPerResident,
		"image_uploaded":    imageKey != "",
	}
	if len(failedPayments) > 0 {
		response["warning"] = fmt.Sprintf("Failed to create payment records for %d residents", len(failedPayments))
		response["failed_residents"] = failedPayments
	}

	return response, nil
}

func (s *billServiceImpl) GetBillByID(ctx context.Context, id int) (map[string]interface{}, error) {
	bill, err := s.repo.GetBillByID(id)
	if err != nil {
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}

	var imageURL string
	if bill.ImageURL != "" {
		imageURL, err = s.imageService.GetImageURL(ctx, bill.ImageURL)
		if err != nil {
			log.Printf("Failed to generate image URL for bill %d: %v", id, err)
		}
	}

	return map[string]interface{}{
		"id":               bill.ID,
		"apartment_id":     bill.ApartmentID,
		"bill_type":        bill.BillType,
		"total_amount":     bill.TotalAmount,
		"due_date":         bill.DueDate,
		"billing_deadline": bill.BillingDeadline,
		"description":      bill.Description,
		"image_url":        imageURL,
		"created_at":       bill.CreatedAt,
		"updated_at":       bill.UpdatedAt,
	}, nil
}

func (s *billServiceImpl) GetBillsByApartmentID(ctx context.Context, apartmentID int) ([]models.Bill, error) {
	bills, err := s.repo.GetBillsByApartmentID(apartmentID)
	if err != nil {
		return nil, fmt.Errorf("failed to get bills: %w", err)
	}
	return bills, nil
}

func (s *billServiceImpl) UpdateBill(ctx context.Context, id, apartmentID int, billType string, totalAmount float64, dueDate, billingDeadline, description string) error {
	bill := models.Bill{
		BaseModel: models.BaseModel{
			ID:        id,
			UpdatedAt: time.Now(),
		},
		ApartmentID:     apartmentID,
		BillType:        models.BillType(billType),
		TotalAmount:     totalAmount,
		DueDate:         dueDate,
		BillingDeadline: billingDeadline,
		Description:     description,
	}

	if err := s.repo.UpdateBill(ctx, bill); err != nil {
		return fmt.Errorf("failed to update bill: %w", err)
	}
	return nil
}

func (s *billServiceImpl) DeleteBill(ctx context.Context, id int) error {
	if err := s.repo.DeleteBill(id); err != nil {
		return fmt.Errorf("failed to delete bill: %w", err)
	}
	return nil
}

func (s *billServiceImpl) PayBills(ctx context.Context, userID int, billIDs []int) error {
	if err := s.paymentService.PayBills(billIDs); err != nil {
		return fmt.Errorf("payment failed: %w", err)
	}

	var payments []models.Payment
	for _, billID := range billIDs {
		payments = append(payments, models.Payment{
			BaseModel: models.BaseModel{
				UpdatedAt: time.Now(),
			},
			BillID:        billID,
			UserID:        userID,
			PaidAt:        time.Now(),
			PaymentStatus: models.Paid,
		})
	}

	if err := s.paymentRepo.UpdatePaymentsStatus(ctx, payments); err != nil {
		return fmt.Errorf("failed to update payments status: %w", err)
	}
	return nil
}

func (s *billServiceImpl) PayBatchBills(ctx context.Context, userID int, billIDs []int) (map[string]interface{}, error) {
	var totalAmount float64
	var validBills []int

	for _, billID := range billIDs {
		payment, err := s.paymentRepo.GetPaymentByBillAndUser(billID, userID)
		if err != nil {
			continue
		}
		if payment.PaymentStatus == models.Pending {
			amount, _ := strconv.ParseFloat(payment.Amount, 64)
			totalAmount += amount
			validBills = append(validBills, billID)
		}
	}

	if len(validBills) == 0 {
		return nil, fmt.Errorf("no valid unpaid bills found")
	}

	if err := s.paymentService.PayBills(validBills); err != nil {
		return nil, fmt.Errorf("batch payment failed: %w", err)
	}

	var payments []models.Payment
	for _, billID := range validBills {
		payments = append(payments, models.Payment{
			BaseModel: models.BaseModel{
				UpdatedAt: time.Now(),
			},
			BillID:        billID,
			UserID:        userID,
			PaidAt:        time.Now(),
			PaymentStatus: models.Paid,
		})
	}

	if err := s.paymentRepo.UpdatePaymentsStatus(ctx, payments); err != nil {
		return nil, fmt.Errorf("failed to update payments status: %w", err)
	}

	return map[string]interface{}{
		"status":       "batch payment successful",
		"bills_paid":   len(validBills),
		"total_amount": totalAmount,
	}, nil
}

func (s *billServiceImpl) GetUnpaidBills(ctx context.Context, userID int) ([]models.Bill, error) {
	apartments, err := s.userApartmentRepo.GetAllApartmentsForAResident(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments: %w", err)
	}

	var unpaidBills []models.Bill
	for _, apartment := range apartments {
		bills, err := s.repo.GetBillsByApartmentID(apartment.ID)
		if err != nil {
			continue
		}

		for _, bill := range bills {
			payment, err := s.paymentRepo.GetPaymentByBillAndUser(bill.ID, userID)
			if err == nil && payment.PaymentStatus == models.Pending {
				unpaidBills = append(unpaidBills, bill)
			}
		}
	}

	return unpaidBills, nil
}

func (s *billServiceImpl) GetBillWithPaymentStatus(ctx context.Context, userID, billID int) (map[string]interface{}, error) {
	bill, err := s.repo.GetBillByID(billID)
	if err != nil {
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	payment, err := s.paymentRepo.GetPaymentByBillAndUser(billID, userID)
	if err != nil {
		return nil, fmt.Errorf("payment record not found: %w", err)
	}

	return map[string]interface{}{
		"bill":           bill,
		"payment_status": payment.PaymentStatus,
		"amount_due":     payment.Amount,
		"paid_at":        payment.PaidAt,
	}, nil
}

func (s *billServiceImpl) GetUserPaymentHistory(ctx context.Context, userID int) ([]PaymentHistoryItem, error) {
	apartments, err := s.userApartmentRepo.GetAllApartmentsForAResident(userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get apartments: %w", err)
	}

	var history []PaymentHistoryItem
	for _, apartment := range apartments {
		bills, err := s.repo.GetBillsByApartmentID(apartment.ID)
		if err != nil {
			continue
		}

		for _, bill := range bills {
			payment, err := s.paymentRepo.GetPaymentByBillAndUser(bill.ID, userID)
			if err == nil {
				history = append(history, PaymentHistoryItem{
					Bill:          bill,
					Payment:       *payment,
					ApartmentName: apartment.ApartmentName,
				})
			}
		}
	}

	return history, nil
}
