package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/handlers"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBillHandler_CreateBill(t *testing.T) {
	tests := []struct {
		name        string
		requestBody map[string]string
		fileContent string
		fileName    string
		setupMocks  func(
			*MockBillRepository,
			*MockUserRepository,
			*MockApartmentRepository,
			*MockUserApartmentRepository,
			*MockPaymentRepository,
			*MockImageService,
			*MockNotificationService,
		)
		expectedStatus   int
		expectedResponse map[string]interface{}
	}{
		{
			name: "successful bill creation without image",
			requestBody: map[string]string{
				"bill_type":        "water",
				"total_amount":     "100.50",
				"due_date":         "2024-01-15",
				"billing_deadline": "2024-01-10",
				"description":      "Water bill for January",
			},
			setupMocks: func(
				billRepo *MockBillRepository,
				userRepo *MockUserRepository,
				apartmentRepo *MockApartmentRepository,
				userApartmentRepo *MockUserApartmentRepository,
				paymentRepo *MockPaymentRepository,
				imageService *MockImageService,
				notificationService *MockNotificationService,
			) {
				// Setup context with user ID
				userID := 1
				apartmentID := 1

				// Mock manager check
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, userID, apartmentID).Return(true, nil)

				// Mock residents
				residents := []models.User{
					{BaseModel: models.BaseModel{ID: 2}},
					{BaseModel: models.BaseModel{ID: 3}},
				}
				userApartmentRepo.On("GetResidentsInApartment", apartmentID).Return(residents, nil)

				// Mock bill creation
				billRepo.On("CreateBill", mock.Anything, mock.Anything).Return(1, nil)

				// Mock payment creation
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(1, nil).Times(len(residents))

				// Mock notifications
				notificationService.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).
					Return(nil).Times(len(residents))
			},
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"id":                float64(1),
				"residents_count":   float64(2),
				"amount_per_person": 50.25,
				"image_uploaded":    false,
			},
		},
		{
			name: "successful bill creation with image",
			requestBody: map[string]string{
				"bill_type":        "electricity",
				"total_amount":     "200.00",
				"due_date":         "2024-01-20",
				"billing_deadline": "2024-01-15",
				"description":      "Electricity bill",
			},
			fileContent: "fake image content",
			fileName:    "bill.jpg",
			setupMocks: func(
				billRepo *MockBillRepository,
				userRepo *MockUserRepository,
				apartmentRepo *MockApartmentRepository,
				userApartmentRepo *MockUserApartmentRepository,
				paymentRepo *MockPaymentRepository,
				imageService *MockImageService,
				notificationService *MockNotificationService,
			) {
				userID := 1
				apartmentID := 1

				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, userID, apartmentID).Return(true, nil)

				residents := []models.User{{BaseModel: models.BaseModel{ID: 2}}}
				userApartmentRepo.On("GetResidentsInApartment", apartmentID).Return(residents, nil)

				billRepo.On("CreateBill", mock.Anything, mock.Anything).Return(1, nil)
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(1, nil)
				notificationService.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

				// Mock image service
				imageService.On("SaveImage", mock.Anything, mock.Anything, "bill.jpg").Return("bills/123_bill.jpg", nil)
			},
			expectedStatus: http.StatusOK,
			expectedResponse: map[string]interface{}{
				"id":                float64(1),
				"residents_count":   float64(1),
				"amount_per_person": 200.00,
				"image_uploaded":    true,
			},
		},
		{
			name: "invalid bill type",
			requestBody: map[string]string{
				"bill_type":    "invalid",
				"total_amount": "100.50",
				"due_date":     "2024-01-15",
			},
			setupMocks: func(
				billRepo *MockBillRepository,
				userRepo *MockUserRepository,
				apartmentRepo *MockApartmentRepository,
				userApartmentRepo *MockUserApartmentRepository,
				paymentRepo *MockPaymentRepository,
				imageService *MockImageService,
				notificationService *MockNotificationService,
			) {
				userID := 1
				apartmentID := 1
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, userID, apartmentID).Return(true, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "not a manager",
			requestBody: map[string]string{
				"bill_type":    "water",
				"total_amount": "100.50",
				"due_date":     "2024-01-15",
			},
			setupMocks: func(
				billRepo *MockBillRepository,
				userRepo *MockUserRepository,
				apartmentRepo *MockApartmentRepository,
				userApartmentRepo *MockUserApartmentRepository,
				paymentRepo *MockPaymentRepository,
				imageService *MockImageService,
				notificationService *MockNotificationService,
			) {
				userID := 1
				apartmentID := 1
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, userID, apartmentID).Return(false, nil)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "no residents in apartment",
			requestBody: map[string]string{
				"bill_type":    "water",
				"total_amount": "100.50",
				"due_date":     "2024-01-15",
			},
			setupMocks: func(
				billRepo *MockBillRepository,
				userRepo *MockUserRepository,
				apartmentRepo *MockApartmentRepository,
				userApartmentRepo *MockUserApartmentRepository,
				paymentRepo *MockPaymentRepository,
				imageService *MockImageService,
				notificationService *MockNotificationService,
			) {
				userID := 1
				apartmentID := 1
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, userID, apartmentID).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", apartmentID).Return([]models.User{}, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mocks
			billRepo := new(MockBillRepository)
			userRepo := new(MockUserRepository)
			apartmentRepo := new(MockApartmentRepository)
			userApartmentRepo := new(MockUserApartmentRepository)
			paymentRepo := new(MockPaymentRepository)
			imageService := new(MockImageService)
			paymentService := new(MockPaymentService)
			notificationService := new(MockNotificationService)

			// Setup mocks
			tt.setupMocks(
				billRepo,
				userRepo,
				apartmentRepo,
				userApartmentRepo,
				paymentRepo,
				imageService,
				notificationService,
			)

			// Create handler
			handler := handlers.NewBillHandler(
				billRepo,
				userRepo,
				apartmentRepo,
				userApartmentRepo,
				paymentRepo,
				imageService,
				paymentService,
				notificationService,
			)

			// Create request
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			// Add form fields
			for key, value := range tt.requestBody {
				_ = writer.WriteField(key, value)
			}

			// Add file if provided
			if tt.fileContent != "" {
				part, _ := writer.CreateFormFile("bill_image", tt.fileName)
				_, _ = io.WriteString(part, tt.fileContent)
			}

			_ = writer.Close()

			req := httptest.NewRequest("POST", "/apartments/1/bills", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			// Add user ID to context
			ctx := context.WithValue(req.Context(), "user_id", 1)
			req = req.WithContext(ctx)

			// Add apartment ID to path
			req = mux.SetURLVars(req, map[string]string{"apartment-id": "1"})

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call handler
			handler.CreateBill(rr, req)

			// Check status code
			assert.Equal(t, tt.expectedStatus, rr.Code)

			// Check response body if expected
			if tt.expectedResponse != nil {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				for key, expectedValue := range tt.expectedResponse {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			// Assert all expectations were met
			billRepo.AssertExpectations(t)
			userRepo.AssertExpectations(t)
			apartmentRepo.AssertExpectations(t)
			userApartmentRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
			imageService.AssertExpectations(t)
			notificationService.AssertExpectations(t)
		})
	}
}

func TestBillHandler_GetBillByID(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		setupMocks     func(*MockBillRepository, *MockImageService)
		expectedStatus int
		expectedBill   *models.Bill
	}{
		{
			name:   "successful retrieval",
			billID: "1",
			setupMocks: func(billRepo *MockBillRepository, imageService *MockImageService) {
				bill := &models.Bill{
					BaseModel:       models.BaseModel{ID: 1},
					ApartmentID:     1,
					BillType:        models.WaterBill,
					TotalAmount:     100.50,
					DueDate:         "2024-01-15",
					BillingDeadline: "2024-01-10",
					Description:     "Water bill",
					ImageURL:        "bills/123_bill.jpg",
				}
				billRepo.On("GetBillByID", 1).Return(bill, nil)
				imageService.On("GetImageURL", mock.Anything, "bills/123_bill.jpg").Return("http://example.com/bills/123_bill.jpg", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBill: &models.Bill{
				BaseModel:       models.BaseModel{ID: 1},
				ApartmentID:     1,
				BillType:        models.WaterBill,
				TotalAmount:     100.50,
				DueDate:         "2024-01-15",
				BillingDeadline: "2024-01-10",
				Description:     "Water bill",
				ImageURL:        "bills/123_bill.jpg",
			},
		},
		{
			name:   "bill not found",
			billID: "999",
			setupMocks: func(billRepo *MockBillRepository, imageService *MockImageService) {
				billRepo.On("GetBillByID", 999).Return(&models.Bill{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:   "invalid bill ID",
			billID: "invalid",
			setupMocks: func(billRepo *MockBillRepository, imageService *MockImageService) {
				// No expectations needed for invalid ID
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			billRepo := new(MockBillRepository)
			imageService := new(MockImageService)

			tt.setupMocks(billRepo, imageService)

			handler := handlers.NewBillHandler(
				billRepo,
				nil, // other repos not needed for this test
				nil,
				nil,
				nil,
				imageService,
				nil,
				nil,
			)

			req := httptest.NewRequest("GET", "/bills?id="+tt.billID, nil)
			rr := httptest.NewRecorder()

			handler.GetBillByID(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedBill != nil {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				assert.NoError(t, err)

				assert.Equal(t, float64(tt.expectedBill.ID), response["id"])
				assert.Equal(t, string(tt.expectedBill.BillType), response["bill_type"])
				assert.Equal(t, tt.expectedBill.TotalAmount, response["total_amount"])
			}

			billRepo.AssertExpectations(t)
			imageService.AssertExpectations(t)
		})
	}
}

func TestBillHandler_PayBills(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    handlers.PayBillsRequest
		setupMocks     func(*MockPaymentRepository, *MockPaymentService)
		expectedStatus int
	}{
		{
			name: "successful payment",
			requestBody: handlers.PayBillsRequest{
				BillIDs: []int{1, 2, 3},
			},
			setupMocks: func(paymentRepo *MockPaymentRepository, paymentService *MockPaymentService) {
				// Mock payment service
				paymentService.On("PayBills", []int{1, 2, 3}).Return(nil)

				// Mock payment updates
				payments := []models.Payment{
					{BillID: 1, UserID: 1, PaymentStatus: models.Paid, PaidAt: time.Now()},
					{BillID: 2, UserID: 1, PaymentStatus: models.Paid, PaidAt: time.Now()},
					{BillID: 3, UserID: 1, PaymentStatus: models.Paid, PaidAt: time.Now()},
				}
				paymentRepo.On("UpdatePaymentsStatus", mock.Anything, payments).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "payment service failure",
			requestBody: handlers.PayBillsRequest{
				BillIDs: []int{1, 2, 3},
			},
			setupMocks: func(paymentRepo *MockPaymentRepository, paymentService *MockPaymentService) {
				paymentService.On("PayBills", []int{1, 2, 3}).Return(errors.New("payment failed"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			paymentRepo := new(MockPaymentRepository)
			paymentService := new(MockPaymentService)

			tt.setupMocks(paymentRepo, paymentService)

			handler := handlers.NewBillHandler(
				nil, // bill repo not needed
				nil,
				nil,
				nil,
				paymentRepo,
				nil,
				paymentService,
				nil,
			)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/bills/pay", bytes.NewReader(body))

			// Add user ID to context
			ctx := context.WithValue(req.Context(), "user_id", 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.PayBills(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			paymentRepo.AssertExpectations(t)
			paymentService.AssertExpectations(t)
		})
	}
}

func TestBillHandler_GetUnpaidBills(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*MockUserApartmentRepository, *MockBillRepository, *MockPaymentRepository)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful retrieval with unpaid bills",
			setupMocks: func(userApartmentRepo *MockUserApartmentRepository, billRepo *MockBillRepository, paymentRepo *MockPaymentRepository) {
				userID := 1

				// Mock apartments
				apartments := []models.Apartment{
					{BaseModel: models.BaseModel{ID: 1}},
					{BaseModel: models.BaseModel{ID: 2}},
				}
				userApartmentRepo.On("GetAllApartmentsForAResident", userID).Return(apartments, nil)

				// Mock bills for first apartment
				bills1 := []models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentID: 1},
					{BaseModel: models.BaseModel{ID: 2}, ApartmentID: 1},
				}
				billRepo.On("GetBillsByApartmentID", 1).Return(bills1, nil)

				// Mock bills for second apartment
				bills2 := []models.Bill{
					{BaseModel: models.BaseModel{ID: 3}, ApartmentID: 2},
				}
				billRepo.On("GetBillsByApartmentID", 2).Return(bills2, nil)

				// Mock payments - first bill is unpaid, others are paid
				paymentRepo.On("GetPaymentByBillAndUser", 1, userID).Return(
					&models.Payment{PaymentStatus: models.Pending}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 2, userID).Return(
					&models.Payment{PaymentStatus: models.Paid}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 3, userID).Return(
					&models.Payment{PaymentStatus: models.Paid}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  1, // Only one unpaid bill
		},
		{
			name: "no unpaid bills",
			setupMocks: func(userApartmentRepo *MockUserApartmentRepository, billRepo *MockBillRepository, paymentRepo *MockPaymentRepository) {
				userID := 1
				apartments := []models.Apartment{{BaseModel: models.BaseModel{ID: 1}}}
				userApartmentRepo.On("GetAllApartmentsForAResident", userID).Return(apartments, nil)

				bills := []models.Bill{{BaseModel: models.BaseModel{ID: 1}, ApartmentID: 1}}
				billRepo.On("GetBillsByApartmentID", 1).Return(bills, nil)

				paymentRepo.On("GetPaymentByBillAndUser", 1, userID).Return(
					&models.Payment{PaymentStatus: models.Paid}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userApartmentRepo := new(MockUserApartmentRepository)
			billRepo := new(MockBillRepository)
			paymentRepo := new(MockPaymentRepository)

			tt.setupMocks(userApartmentRepo, billRepo, paymentRepo)

			handler := handlers.NewBillHandler(
				billRepo,
				nil,
				nil,
				userApartmentRepo,
				paymentRepo,
				nil,
				nil,
				nil,
			)

			req := httptest.NewRequest("GET", "/bills/unpaid", nil)

			// Add user ID to context
			ctx := context.WithValue(req.Context(), "user_id", 1)
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.GetUnpaidBills(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)

			if tt.expectedCount > 0 {
				var bills []models.Bill
				err := json.Unmarshal(rr.Body.Bytes(), &bills)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(bills))
			}

			userApartmentRepo.AssertExpectations(t)
			billRepo.AssertExpectations(t)
			paymentRepo.AssertExpectations(t)
		})
	}
}

// Additional tests for other handler methods would follow the same pattern
