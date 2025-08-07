package services

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/payment"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/sirupsen/logrus"
)

type BillService interface {
	CreateBill(ctx context.Context, userID, apartmentID int, req dto.CreateBillRequest, file io.ReadCloser, handler *multipart.FileHeader) (map[string]interface{}, error)
	GetBillByID(ctx context.Context, id int) (map[string]interface{}, error)
	GetBillsByApartmentID(ctx context.Context, apartmentID int) ([]models.Bill, error)
	UpdateBill(ctx context.Context, id, apartmentID int, billType string, totalAmount float64, dueDate, billingDeadline, description string) error
	DeleteBill(ctx context.Context, id int) error
	PayBills(ctx context.Context, userID int, billIDs []int, idempotentKey string) error
	PayBatchBills(ctx context.Context, userID int, billIDs []int, idempotentKey string) (map[string]interface{}, error)
	GetUnpaidBills(ctx context.Context, userID int) ([]models.Bill, error)
	GetBillWithPaymentStatus(ctx context.Context, userID, billID int) (map[string]interface{}, error)
	GetUserPaymentHistory(ctx context.Context, userID int) ([]PaymentHistoryItem, error)
	DivideBillByType(ctx context.Context, userID, apartmentID int, billType models.BillType) (map[string]interface{}, error)
	DivideAllBills(ctx context.Context, userID, apartmentID int) (map[string]interface{}, error)
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
	logger := logrus.WithFields(logrus.Fields{
		"user_id":      userID,
		"apartment_id": apartmentID,
		"bill_type":    req.BillType,
		"amount":       req.TotalAmount,
	})

	logger.Info("Starting bill creation")

	//checking if this apartment with this id exists
	_, err := s.apartmentRepo.GetApartmentByID(apartmentID)
	if err != nil {
		logger.WithError(err).Error("Apartment not found")
		return nil, fmt.Errorf("the apartment id is incorrect: %w", err)
	}

	isManager, err := s.userApartmentRepo.IsUserManagerOfApartment(ctx, userID, apartmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to verify manager status")
		return nil, fmt.Errorf("failed to verify manager status: %w", err)
	}
	if !isManager {
		logger.Warn("Non-manager user attempted to create bill")
		return nil, fmt.Errorf("only apartment managers can create bills")
	}

	if req.BillType == "" || req.TotalAmount <= 0 || req.DueDate == "" {
		logger.Error("Missing required fields for bill creation")
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
		logger.WithField("provided_type", req.BillType).Error("Invalid bill type provided")
		return nil, fmt.Errorf("invalid bill type")
	}

	if _, err := time.Parse("2006-01-02", req.DueDate); err != nil {
		logger.WithError(err).Error("Invalid due date format")
		return nil, fmt.Errorf("invalid due date format (use YYYY-MM-DD)")
	}
	if req.BillingDeadline != "" {
		if _, err := time.Parse("2006-01-02", req.BillingDeadline); err != nil {
			logger.WithError(err).Error("Invalid billing deadline format")
			return nil, fmt.Errorf("invalid billing deadline format (use YYYY-MM-DD)")
		}
	}

	var imageKey string
	if file != nil {
		logger.Debug("Processing image upload")
		fileBytes, err := io.ReadAll(file)
		file.Close()
		if err != nil {
			logger.WithError(err).Error("Failed to read uploaded file")
			return nil, fmt.Errorf("failed to read file: %w", err)
		}

		imageKey, err = s.imageService.SaveImage(ctx, fileBytes, handler.Filename)
		if err != nil {
			logger.WithError(err).WithField("filename", handler.Filename).Error("Failed to save image")
			return nil, fmt.Errorf("failed to save image: %w", err)
		}
		logger.WithField("image_key", imageKey).Debug("Image uploaded successfully")
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
		logger.WithError(err).Error("Failed to create bill in database")
		if imageKey != "" {
			if delErr := s.imageService.DeleteImage(ctx, imageKey); delErr != nil {
				logger.WithError(delErr).WithField("image_key", imageKey).Error("Failed to cleanup uploaded image after bill creation failure")
			}
		}
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	logger.WithField("bill_id", billID).Info("Bill created successfully")

	response := map[string]interface{}{
		"id":             billID,
		"total_amount":   req.TotalAmount,
		"image_uploaded": imageKey != "",
		"status":         "Bill created successfully. Use divide endpoints to create payment records for residents.",
	}

	return response, nil
}

func (s *billServiceImpl) DivideBillByType(ctx context.Context, userID, apartmentID int, billType models.BillType) (map[string]interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"user_id":      userID,
		"apartment_id": apartmentID,
		"bill_type":    billType,
	})

	logger.Info("Starting bill division by type")

	isManager, err := s.userApartmentRepo.IsUserManagerOfApartment(ctx, userID, apartmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to verify manager status")
		return nil, fmt.Errorf("failed to verify manager status: %w", err)
	}
	if !isManager {
		logger.Warn("Non-manager user attempted to divide bills")
		return nil, fmt.Errorf("only apartment managers can divide bills")
	}

	residents, err := s.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to get residents")
		return nil, fmt.Errorf("failed to get residents: %w", err)
	}
	if len(residents) == 0 {
		logger.Warn("No residents found in apartment")
		return nil, fmt.Errorf("no residents found in apartment")
	}

	logger.WithField("residents_count", len(residents)).Debug("Retrieved residents for bill division")

	//bills of specific type that haven't been divided yet
	bills, err := s.repo.GetUndividedBillsByTypeAndApartment(apartmentID, billType)
	if err != nil {
		logger.WithError(err).Error("Failed to get undivided bills")
		return nil, fmt.Errorf("failed to get bills: %w", err)
	}
	if len(bills) == 0 {
		logger.WithField("bill_type", billType).Warn("No undivided bills found")
		return nil, fmt.Errorf("no undivided bills of type %s found", billType)
	}

	logger.WithField("bills_count", len(bills)).Info("Processing undivided bills")

	var processedBills []int
	var failedBills []int
	var totalFailedPayments int

	for _, bill := range bills {
		billLogger := logger.WithFields(logrus.Fields{
			"bill_id":     bill.ID,
			"bill_amount": bill.TotalAmount,
		})

		amountPerResident := bill.TotalAmount / float64(len(residents))
		billProcessed := true

		for _, resident := range residents {
			//checking if payment record already exists
			existingPayment, _ := s.paymentRepo.GetPaymentByBillAndUser(bill.ID, resident.ID)
			if existingPayment != nil {
				continue // if payment record already exists
			}

			payment := models.Payment{
				BaseModel: models.BaseModel{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				BillID:        bill.ID,
				UserID:        resident.ID,
				Amount:        fmt.Sprintf("%.2f", amountPerResident),
				PaymentStatus: models.Pending,
			}

			_, err := s.paymentRepo.CreatePayment(ctx, payment)
			if err != nil {
				billLogger.WithError(err).WithField("resident_id", resident.ID).Error("Failed to create payment record")
				totalFailedPayments++
				billProcessed = false
				continue
			}

			//sending notification
			if err := s.notificationService.SendBillNotification(ctx, resident.ID, bill, amountPerResident); err != nil {
				billLogger.WithError(err).WithField("resident_id", resident.ID).Warn("Failed to send notification")
			}
		}

		if billProcessed {
			processedBills = append(processedBills, bill.ID)
			billLogger.Debug("Bill processed successfully")
		} else {
			failedBills = append(failedBills, bill.ID)
			billLogger.Error("Bill processing failed")
		}
	}

	logger.WithFields(logrus.Fields{
		"processed_count": len(processedBills),
		"failed_count":    len(failedBills),
	}).Info("Bill division completed")

	response := map[string]interface{}{
		"bill_type":       billType,
		"residents_count": len(residents),
		"processed_bills": processedBills,
		"processed_count": len(processedBills),
	}

	if len(failedBills) > 0 {
		response["warning"] = fmt.Sprintf("Failed to process %d bills completely", len(failedBills))
		response["failed_bills"] = failedBills
		response["failed_payments_count"] = totalFailedPayments
	}

	return response, nil
}

func (s *billServiceImpl) DivideAllBills(ctx context.Context, userID, apartmentID int) (map[string]interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"user_id":      userID,
		"apartment_id": apartmentID,
	})

	logger.Info("Starting division of all bills")

	isManager, err := s.userApartmentRepo.IsUserManagerOfApartment(ctx, userID, apartmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to verify manager status")
		return nil, fmt.Errorf("failed to verify manager status: %w", err)
	}
	if !isManager {
		logger.Warn("Non-manager user attempted to divide all bills")
		return nil, fmt.Errorf("only apartment managers can divide bills")
	}

	// Get current residents
	residents, err := s.userApartmentRepo.GetResidentsInApartment(apartmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to get residents")
		return nil, fmt.Errorf("failed to get residents: %w", err)
	}
	if len(residents) == 0 {
		logger.Warn("No residents found in apartment")
		return nil, fmt.Errorf("no residents found in apartment")
	}

	//all undivided bills for the apartment
	bills, err := s.repo.GetUndividedBillsByApartment(apartmentID)
	if err != nil {
		logger.WithError(err).Error("Failed to get undivided bills")
		return nil, fmt.Errorf("failed to get bills: %w", err)
	}
	if len(bills) == 0 {
		logger.Warn("No undivided bills found in apartment")
		return nil, fmt.Errorf("no undivided bills found in apartment")
	}

	logger.WithFields(logrus.Fields{
		"residents_count": len(residents),
		"bills_count":     len(bills),
	}).Info("Processing all undivided bills")

	var processedBills []int
	var failedBills []int
	var totalFailedPayments int
	billTypeCount := make(map[models.BillType]int)

	for _, bill := range bills {
		amountPerResident := bill.TotalAmount / float64(len(residents))
		billProcessed := true
		billTypeCount[bill.BillType]++

		for _, resident := range residents {
			existingPayment, _ := s.paymentRepo.GetPaymentByBillAndUser(bill.ID, resident.ID)
			if existingPayment != nil {
				continue
			}

			payment := models.Payment{
				BaseModel: models.BaseModel{
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
				BillID:        bill.ID,
				UserID:        resident.ID,
				Amount:        fmt.Sprintf("%.2f", amountPerResident),
				PaymentStatus: models.Pending,
			}

			_, err := s.paymentRepo.CreatePayment(ctx, payment)
			if err != nil {
				logger.WithError(err).WithFields(logrus.Fields{
					"bill_id":     bill.ID,
					"resident_id": resident.ID,
				}).Error("Failed to create payment record")
				totalFailedPayments++
				billProcessed = false
				continue
			}

			//sending notification
			if err := s.notificationService.SendBillNotification(ctx, resident.ID, bill, amountPerResident); err != nil {
				logger.WithError(err).WithFields(logrus.Fields{
					"bill_id":     bill.ID,
					"resident_id": resident.ID,
				}).Warn("Failed to send notification")
			}
		}

		if billProcessed {
			processedBills = append(processedBills, bill.ID)
		} else {
			failedBills = append(failedBills, bill.ID)
		}
	}

	logger.WithFields(logrus.Fields{
		"processed_count":      len(processedBills),
		"failed_count":         len(failedBills),
		"bill_types_processed": billTypeCount,
	}).Info("All bills division completed")

	response := map[string]interface{}{
		"residents_count":      len(residents),
		"processed_bills":      processedBills,
		"processed_count":      len(processedBills),
		"bill_types_processed": billTypeCount,
	}

	if len(failedBills) > 0 {
		response["warning"] = fmt.Sprintf("Failed to process %d bills completely", len(failedBills))
		response["failed_bills"] = failedBills
		response["failed_payments_count"] = totalFailedPayments
	}

	return response, nil
}

func (s *billServiceImpl) GetBillByID(ctx context.Context, id int) (map[string]interface{}, error) {
	bill, err := s.repo.GetBillByID(id)
	if err != nil {
		logrus.WithError(err).WithField("bill_id", id).Error("Failed to get bill by ID")
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}

	var imageURL string
	if bill.ImageURL != "" {
		imageURL, err = s.imageService.GetImageURL(ctx, bill.ImageURL)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"bill_id":   id,
				"image_key": bill.ImageURL,
			}).Warn("Failed to generate image URL")
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
		logrus.WithError(err).WithField("apartment_id", apartmentID).Error("Failed to get bills by apartment ID")
		return nil, fmt.Errorf("failed to get bills: %w", err)
	}
	return bills, nil
}

func (s *billServiceImpl) UpdateBill(ctx context.Context, id, apartmentID int, billType string, totalAmount float64, dueDate, billingDeadline, description string) error {
	logger := logrus.WithFields(logrus.Fields{
		"bill_id":      id,
		"apartment_id": apartmentID,
		"bill_type":    billType,
		"amount":       totalAmount,
	})

	logger.Info("Updating bill")

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
		logger.WithError(err).Error("Failed to update bill")
		return fmt.Errorf("failed to update bill: %w", err)
	}

	logger.Info("Bill updated successfully")
	return nil
}

func (s *billServiceImpl) DeleteBill(ctx context.Context, id int) error {
	logger := logrus.WithField("bill_id", id)
	logger.Info("Deleting bill")

	bill, err := s.repo.GetBillByID(id)
	if err != nil {
		logger.WithError(err).Error("Failed to get bill for deletion")
		return fmt.Errorf("failed to get bill: %w", err)
	}

	if err := s.repo.DeleteBill(id); err != nil {
		logger.WithError(err).Error("Failed to delete bill from database")
		return fmt.Errorf("failed to delete bill: %w", err)
	}

	if bill.ImageURL != "" {
		if err := s.imageService.DeleteImage(ctx, bill.ImageURL); err != nil {
			logger.WithError(err).WithField("image_key", bill.ImageURL).Warn("Failed to delete associated image")
		} else {
			logger.WithField("image_key", bill.ImageURL).Debug("Associated image deleted successfully")
		}
	}

	logger.Info("Bill deleted successfully")
	return nil
}

func (s *billServiceImpl) PayBills(ctx context.Context, userID int, billIDs []int, idempotentKey string) error {
	logger := logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"bill_ids":   billIDs,
		"bill_count": len(billIDs),
	})

	logger.Info("Processing bill payment")

	if err := s.paymentService.PayBills(billIDs, idempotentKey); err != nil {
		logger.WithError(err).Error("Payment processing failed")
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
		logger.WithError(err).Error("Failed to update payment status")
		return fmt.Errorf("failed to update payments status: %w", err)
	}

	logger.Info("Bill payment completed successfully")
	return nil
}

func (s *billServiceImpl) PayBatchBills(ctx context.Context, userID int, billIDs []int, idempotentKey string) (map[string]interface{}, error) {
	logger := logrus.WithFields(logrus.Fields{
		"user_id":    userID,
		"bill_count": len(billIDs),
	})

	logger.Info("Processing batch bill payment")

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
		logger.Warn("No valid unpaid bills found for batch payment")
		return nil, fmt.Errorf("no valid unpaid bills found")
	}

	logger.WithFields(logrus.Fields{
		"valid_bills_count": len(validBills),
		"total_amount":      totalAmount,
	}).Info("Processing batch payment for valid bills")

	if err := s.paymentService.PayBills(validBills, idempotentKey); err != nil {
		logger.WithError(err).Error("Batch payment processing failed")
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
		logger.WithError(err).Error("Failed to update batch payment status")
		return nil, fmt.Errorf("failed to update payments status: %w", err)
	}

	logger.WithFields(logrus.Fields{
		"bills_paid":   len(validBills),
		"total_amount": totalAmount,
	}).Info("Batch payment completed successfully")

	return map[string]interface{}{
		"status":       "batch payment successful",
		"bills_paid":   len(validBills),
		"total_amount": totalAmount,
	}, nil
}

func (s *billServiceImpl) GetUnpaidBills(ctx context.Context, userID int) ([]models.Bill, error) {
	apartments, err := s.userApartmentRepo.GetAllApartmentsForAResident(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get apartments for user")
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

	logrus.WithFields(logrus.Fields{
		"user_id":            userID,
		"unpaid_bills_count": len(unpaidBills),
	}).Debug("Retrieved unpaid bills for user")

	return unpaidBills, nil
}

func (s *billServiceImpl) GetBillWithPaymentStatus(ctx context.Context, userID, billID int) (map[string]interface{}, error) {
	bill, err := s.repo.GetBillByID(billID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"bill_id": billID,
		}).Error("Bill not found")
		return nil, fmt.Errorf("bill not found: %w", err)
	}

	payment, err := s.paymentRepo.GetPaymentByBillAndUser(billID, userID)
	if err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": userID,
			"bill_id": billID,
		}).Error("Payment record not found")
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
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get apartments for payment history")
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

	logrus.WithFields(logrus.Fields{
		"user_id":       userID,
		"history_count": len(history),
	}).Debug("Retrieved payment history for user")

	return history, nil
}
