package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/handlers"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/payment"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateBill(t *testing.T) {
	tests := []struct {
		name       string
		setupMocks func(
			*repositories.MockApartmentRepo,
			*repositories.MockUserApartmentRepository,
			*repositories.MockBillRepository,
			*image.MockImage,
		)
		requestBody    map[string]string
		fileUpload     bool
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful bill creation with image",
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				imgService.On("SaveImage", mock.Anything, mock.Anything, "test.jpg").Return("image123", nil)
				billRepo.On("CreateBill", mock.Anything, mock.Anything).Return(1, nil)
			},
			requestBody: map[string]string{
				"bill_type":        "water",
				"total_amount":     "100.50",
				"due_date":         "2023-12-31",
				"billing_deadline": "2023-12-25",
				"description":      "Test bill",
			},
			fileUpload:     true,
			expectedStatus: http.StatusCreated,
			expectedBody:   `"status":"Bill created successfully. Use divide endpoints to create payment records for residents."`,
		},
		{
			name: "successful bill creation without image",
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				billRepo.On("CreateBill", mock.Anything, mock.Anything).Return(1, nil)
			},
			requestBody: map[string]string{
				"bill_type":    "electricity",
				"total_amount": "200.75",
				"due_date":     "2023-12-31",
				"description":  "Test bill",
			},
			fileUpload:     false,
			expectedStatus: http.StatusCreated,
			expectedBody:   `"status":"Bill created successfully. Use divide endpoints to create payment records for residents."`,
		},
		{
			name: "invalid apartment ID",
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(nil, errors.New("not found"))
			},
			requestBody: map[string]string{
				"bill_type":    "water",
				"total_amount": "100.50",
				"due_date":     "2023-12-31",
			},
			fileUpload:     false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"the apartment id is incorrect"`,
		},
		{
			name: "non-manager user",
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			requestBody: map[string]string{
				"bill_type":    "water",
				"total_amount": "100.50",
				"due_date":     "2023-12-31",
			},
			fileUpload:     false,
			expectedStatus: http.StatusForbidden,
			expectedBody:   `"only apartment managers can create bills"`,
		},
		{
			name: "invalid bill type",
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
			},
			requestBody: map[string]string{
				"bill_type":    "invalid_type",
				"total_amount": "100.50",
				"due_date":     "2023-12-31",
			},
			fileUpload:     false,
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"invalid bill type"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockApartmentRepo := new(repositories.MockApartmentRepo)
			mockUserApartmentRepo := new(repositories.MockUserApartmentRepository)
			mockBillRepo := new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockApartmentRepo, mockUserApartmentRepo, mockBillRepo, mockImageService)

			billService := services.NewBillService(
				nil,
				nil,
				mockApartmentRepo,
				mockUserApartmentRepo,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)
			handler := handlers.NewBillHandler(billService)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			for key, value := range tt.requestBody {
				_ = writer.WriteField(key, value)
			}

			if tt.fileUpload {
				part, _ := writer.CreateFormFile("bill_image", "test.jpg")
				_, _ = part.Write([]byte("test image content"))
			}

			writer.Close()

			req := httptest.NewRequest("POST", "/apartments/1/bills", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			req = req.WithContext(ctx)

			req = muxSetPathParams(req, map[string]string{"apartment_id": "1"})

			rr := httptest.NewRecorder()

			handler.CreateBill(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockApartmentRepo.AssertExpectations(t)
			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockImageService.AssertExpectations(t)
		})
	}
}

func TestDivideBillByType(t *testing.T) {
	tests := []struct {
		name           string
		billType       string
		setupMocks     func(*repositories.MockUserApartmentRepository, *repositories.MockBillRepository, *repositories.MockPaymentRepository, *notification.MockNotification)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:     "successful division by type",
			billType: "water",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}, {BaseModel: models.BaseModel{ID: 3}}}, nil)
				billRepo.On("GetUndividedBillsByTypeAndApartment", 1, models.WaterBill).Return([]models.Bill{{BaseModel: models.BaseModel{ID: 1}, TotalAmount: 100}}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 2).Return(nil, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 3).Return(nil, nil)
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(1, nil).Times(2)
				notifService.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"bill_type":"water"`,
		},
		{
			name:     "invalid bill type",
			billType: "invalid_type",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {

			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid bill type"`,
		},
		{
			name:     "non-manager user",
			billType: "water",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `"only apartment managers can divide bills"`,
		},
		{
			name:     "no residents in apartment",
			billType: "water",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"no residents found in apartment"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockUserApartmentRepo := new(repositories.MockUserApartmentRepository)
			mockBillRepo := new(repositories.MockBillRepository)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockNotificationService := new(notification.MockNotification)
			mockImageService := new(image.MockImage)
			mockPaymentService := new(payment.MockPayment)

			tt.setupMocks(mockUserApartmentRepo, mockBillRepo, mockPaymentRepo, mockNotificationService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				mockUserApartmentRepo,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("POST", fmt.Sprintf("/apartments/1/bills/divide/%s", tt.billType), nil)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			req = req.WithContext(ctx)

			req = muxSetPathParams(req, map[string]string{
				"apartment_id": "1",
				"bill_type":    tt.billType,
			})

			rr := httptest.NewRecorder()

			handler.DivideBillByType(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			mockNotificationService.AssertExpectations(t)
		})
	}
}

func TestDivideAllBills(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*repositories.MockUserApartmentRepository, *repositories.MockBillRepository, *repositories.MockPaymentRepository, *notification.MockNotification)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful division of all bills",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}, {BaseModel: models.BaseModel{ID: 3}}}, nil)
				billRepo.On("GetUndividedBillsByApartment", 1).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, BillType: models.WaterBill, TotalAmount: 100},
					{BaseModel: models.BaseModel{ID: 2}, BillType: models.ElectricityBill, TotalAmount: 200},
				}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", mock.Anything, mock.Anything).Return(nil, nil)
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(1, nil).Times(4)
				notifService.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(4)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"residents_count":2`,
		},
		{
			name: "non-manager user",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   `"only apartment managers can divide bills"`,
		},
		{
			name: "no undivided bills",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}}, nil)
				billRepo.On("GetUndividedBillsByApartment", 1).Return([]models.Bill{}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"no undivided bills found in apartment"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockUserApartmentRepo := new(repositories.MockUserApartmentRepository)
			mockBillRepo := new(repositories.MockBillRepository)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockNotificationService := new(notification.MockNotification)
			mockImageService := new(image.MockImage)
			mockPaymentService := new(payment.MockPayment)

			tt.setupMocks(mockUserApartmentRepo, mockBillRepo, mockPaymentRepo, mockNotificationService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				mockUserApartmentRepo,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("POST", "/apartments/1/bills/divide", nil)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			req = req.WithContext(ctx)

			req = muxSetPathParams(req, map[string]string{"apartment_id": "1"})

			rr := httptest.NewRecorder()

			handler.DivideAllBills(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			mockNotificationService.AssertExpectations(t)
		})
	}
}

func TestGetBillByID(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		setupMocks     func(*repositories.MockBillRepository, *image.MockImage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful bill retrieval",
			billID: "1",
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(&models.Bill{
					BaseModel:   models.BaseModel{ID: 1},
					ApartmentID: 1,
					BillType:    models.WaterBill,
					TotalAmount: 100.50,
					DueDate:     "2023-12-31",
					Description: "Test bill",
					ImageURL:    "image123",
				}, nil)
				imgService.On("GetImageURL", mock.Anything, "image123").Return("http://example.com/image123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"id":1,"bill_type":"water"`,
		},
		{
			name:           "invalid bill ID",
			billID:         "invalid",
			setupMocks:     func(*repositories.MockBillRepository, *image.MockImage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid bill ID"`,
		},
		{
			name:   "bill not found",
			billID: "1",
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"Bill not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)

			tt.setupMocks(mockBillRepo, mockImageService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				nil,
				mockImageService,
				nil,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("GET", fmt.Sprintf("/bills?id=%s", tt.billID), nil)
			rr := httptest.NewRecorder()

			handler.GetBillByID(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockBillRepo.AssertExpectations(t)
			mockImageService.AssertExpectations(t)
		})
	}
}

func TestGetBillsByApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		setupMocks     func(*repositories.MockBillRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:        "successful bills retrieval",
			apartmentID: "1",
			setupMocks: func(billRepo *repositories.MockBillRepository) {
				billRepo.On("GetBillsByApartmentID", 1).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentID: 1, BillType: models.WaterBill},
					{BaseModel: models.BaseModel{ID: 2}, ApartmentID: 1, BillType: models.ElectricityBill},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"apartment_id":1,"bill_type":"water"`,
		},
		{
			name:           "invalid apartment ID",
			apartmentID:    "invalid",
			setupMocks:     func(*repositories.MockBillRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid apartment ID"`,
		},
		{
			name:        "no bills found",
			apartmentID: "1",
			setupMocks: func(billRepo *repositories.MockBillRepository) {
				billRepo.On("GetBillsByApartmentID", 1).Return([]models.Bill{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)

			tt.setupMocks(mockBillRepo)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("GET", fmt.Sprintf("/bills?apartment_id=%s", tt.apartmentID), nil)
			rr := httptest.NewRecorder()

			handler.GetBillsByApartment(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockBillRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateBill(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setupMocks     func(*repositories.MockBillRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful bill update",
			requestBody: map[string]interface{}{
				"id":               1,
				"apartment_id":     1,
				"bill_type":        "water",
				"total_amount":     100.50,
				"due_date":         "2023-12-31",
				"billing_deadline": "2023-12-25",
				"description":      "Updated bill",
			},
			setupMocks: func(billRepo *repositories.MockBillRepository) {
				billRepo.On("UpdateBill", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"id":           "invalid",
				"apartment_id": 1,
			},
			setupMocks:     func(*repositories.MockBillRepository) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid request body"`,
		},
		{
			name: "update failed",
			requestBody: map[string]interface{}{
				"id":               1,
				"apartment_id":     1,
				"bill_type":        "water",
				"total_amount":     100.50,
				"due_date":         "2023-12-31",
				"billing_deadline": "2023-12-25",
				"description":      "Updated bill",
			},
			setupMocks: func(billRepo *repositories.MockBillRepository) {
				billRepo.On("UpdateBill", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"Failed to update bill"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)

			tt.setupMocks(mockBillRepo)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/bills", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			handler.UpdateBill(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}

			mockBillRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteBill(t *testing.T) {
	tests := []struct {
		name           string
		billID         string
		setupMocks     func(*repositories.MockBillRepository, *image.MockImage)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:   "successful bill deletion",
			billID: "1",
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(&models.Bill{BaseModel: models.BaseModel{ID: 1}, ImageURL: "image123"}, nil)
				billRepo.On("DeleteBill", 1)
				imgService.On("DeleteImage", mock.Anything, "image123").Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid bill ID",
			billID:         "invalid",
			setupMocks:     func(*repositories.MockBillRepository, *image.MockImage) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Invalid bill ID"`,
		},
		{
			name:   "bill not found",
			billID: "1",
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(nil, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   `"Bill not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)

			tt.setupMocks(mockBillRepo, mockImageService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				nil,
				mockImageService,
				nil,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("DELETE", fmt.Sprintf("/bills?id=%s", tt.billID), nil)
			rr := httptest.NewRecorder()

			handler.DeleteBill(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBody)
			}

			mockBillRepo.AssertExpectations(t)
			mockImageService.AssertExpectations(t)
		})
	}
}

func TestPayBill(t *testing.T) {
	tests := []struct {
		name           string
		paymentID      string
		setupMocks     func(*repositories.MockPaymentRepository, *payment.MockPayment)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:      "successful payment",
			paymentID: "1",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentService.On("PayBills", []int{1}, mock.Anything).Return(nil)
				paymentRepo.On("UpdatePaymentsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"status":"payment successful"`,
		},
		{
			name:           "invalid payment ID",
			paymentID:      "invalid",
			setupMocks:     func(*repositories.MockPaymentRepository, *payment.MockPayment) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"Failed to get payment ID"`,
		},
		{
			name:      "payment failed",
			paymentID: "1",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentService.On("PayBills", []int{1}, mock.Anything).Return(errors.New("payment failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"Payment failed"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)

			tt.setupMocks(mockPaymentRepo, mockPaymentService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				nil,
				mockPaymentService,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("POST", fmt.Sprintf("/payments/%s", tt.paymentID), nil)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			ctx = context.WithValue(ctx, middleware.IdempotentKey, "idemp123")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.PayBill(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockPaymentRepo.AssertExpectations(t)
			mockPaymentService.AssertExpectations(t)
		})
	}
}

func TestPayBatchBills(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*repositories.MockPaymentRepository, *payment.MockPayment)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful batch payment",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{
					{BaseModel: models.BaseModel{ID: 1}, Amount: "100.50"},
					{BaseModel: models.BaseModel{ID: 2}, Amount: "200.75"},
				}, nil)
				paymentService.On("PayBills", []int{1, 2}, mock.Anything).Return(nil)
				paymentRepo.On("UpdatePaymentsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"total_amount":301.25`,
		},
		{
			name: "no pending payments",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{}, nil)
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   `"no valid unpaid bills found"`,
		},
		{
			name: "payment failed",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{
					{BaseModel: models.BaseModel{ID: 1}, Amount: "100.50"},
				}, nil)
				paymentService.On("PayBills", []int{1}, mock.Anything).Return(errors.New("payment failed"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedBody:   `"batch payment failed"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)

			tt.setupMocks(mockPaymentRepo, mockPaymentService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				nil,
				mockPaymentService,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("POST", "/payments/batch", nil)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			ctx = context.WithValue(ctx, middleware.IdempotentKey, "idemp123")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.PayBatchBills(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockPaymentRepo.AssertExpectations(t)
			mockPaymentService.AssertExpectations(t)
		})
	}
}

func TestGetUnpaidBills(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*repositories.MockPaymentRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful retrieval of unpaid bills",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{
					{BaseModel: models.BaseModel{ID: 1}, Amount: "100.50"},
					{BaseModel: models.BaseModel{ID: 2}, Amount: "200.75"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[{"id":1,"amount":"100.50"`,
		},
		{
			name: "no unpaid bills",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPaymentRepo := new(repositories.MockPaymentRepository)

			tt.setupMocks(mockPaymentRepo)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				nil,
				nil,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("GET", "/bills/unpaid", nil)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.GetUnpaidBills(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserPaymentHistory(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*repositories.MockUserApartmentRepository, *repositories.MockBillRepository, *repositories.MockPaymentRepository)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "successful retrieval of payment history",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
			) {
				userApartmentRepo.On("GetAllApartmentsForAResident", 1).Return([]models.Apartment{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentName: "Test Apartment"},
				}, nil)
				billRepo.On("GetBillsByApartmentID", 1).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentID: 1},
				}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 1).Return(&models.Payment{
					BaseModel: models.BaseModel{ID: 1},
					BillID:    1,
					UserID:    1,
					Amount:    "100.50",
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"apartment_name":"Test Apartment"`,
		},
		{
			name: "no payment history",
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
			) {
				userApartmentRepo.On("GetAllApartmentsForAResident", 1).Return([]models.Apartment{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `[]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockUserApartmentRepo := new(repositories.MockUserApartmentRepository)
			mockBillRepo := new(repositories.MockBillRepository)
			mockPaymentRepo := new(repositories.MockPaymentRepository)

			tt.setupMocks(mockUserApartmentRepo, mockBillRepo, mockPaymentRepo)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				mockUserApartmentRepo,
				mockPaymentRepo,
				nil,
				nil,
				nil,
			)
			handler := handlers.NewBillHandler(billService)

			req := httptest.NewRequest("GET", "/payments/history", nil)

			ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
			req = req.WithContext(ctx)

			rr := httptest.NewRecorder()

			handler.GetUserPaymentHistory(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code)
			assert.Contains(t, rr.Body.String(), tt.expectedBody)

			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func muxSetPathParams(req *http.Request, params map[string]string) *http.Request {
	ctx := req.Context()
	for key, val := range params {
		ctx = context.WithValue(ctx, key, val)
	}
	return req.WithContext(ctx)
}
