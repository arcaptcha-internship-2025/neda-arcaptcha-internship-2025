package services_test

import (
	"context"
	"errors"
	"io"
	"mime/multipart"
	"strings"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
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
		name        string
		userID      int
		apartmentID int
		req         dto.CreateBillRequest
		file        io.ReadCloser
		fileHeader  *multipart.FileHeader
		setupMocks  func(
			*repositories.MockApartmentRepo,
			*repositories.MockUserApartmentRepository,
			*repositories.MockBillRepository,
			*image.MockImage,
		)
		expectedError error
	}{
		{
			name:        "successful bill creation with image",
			userID:      1,
			apartmentID: 1,
			req: dto.CreateBillRequest{
				BillType:        models.WaterBill,
				TotalAmount:     100.50,
				DueDate:         "2023-12-31",
				BillingDeadline: "2023-12-25",
				Description:     "Test bill",
			},
			file:       io.NopCloser(strings.NewReader("test image content")),
			fileHeader: &multipart.FileHeader{Filename: "test.jpg"},
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				imgService.On("SaveImage", mock.Anything, []byte("test image content"), "test.jpg").Return("image123", nil)
				billRepo.On("CreateBill", mock.Anything, mock.Anything).Return(1, nil)
			},
			expectedError: nil,
		},
		{
			name:        "invalid apartment ID",
			userID:      1,
			apartmentID: 999,
			req: dto.CreateBillRequest{
				BillType:    models.WaterBill,
				TotalAmount: 100.50,
				DueDate:     "2023-12-31",
			},
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				_ *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 999).Return(nil, errors.New("apartment not found"))
			},
			expectedError: errors.New("the apartment id is incorrect"),
		},
		{
			name:        "non-manager user",
			userID:      2,
			apartmentID: 1,
			req: dto.CreateBillRequest{
				BillType:    models.WaterBill,
				TotalAmount: 100.50,
				DueDate:     "2023-12-31",
			},
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 2, 1).Return(false, nil)
			},
			expectedError: errors.New("only apartment managers can create bills"),
		},
		{
			name:        "invalid bill type",
			userID:      1,
			apartmentID: 1,
			req: dto.CreateBillRequest{
				BillType:    "invalid_type",
				TotalAmount: 100.50,
				DueDate:     "2023-12-31",
			},
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
			},
			expectedError: errors.New("invalid bill type"),
		},
		{
			name:        "invalid due date format",
			userID:      1,
			apartmentID: 1,
			req: dto.CreateBillRequest{
				BillType:    models.WaterBill,
				TotalAmount: 100.50,
				DueDate:     "31-12-2023",
			},
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
			},
			expectedError: errors.New("invalid due date format"),
		},
		{
			name:        "image upload failure",
			userID:      1,
			apartmentID: 1,
			req: dto.CreateBillRequest{
				BillType:    models.WaterBill,
				TotalAmount: 100.50,
				DueDate:     "2023-12-31",
			},
			file:       io.NopCloser(strings.NewReader("test image content")),
			fileHeader: &multipart.FileHeader{Filename: "test.jpg"},
			setupMocks: func(
				apartmentRepo *repositories.MockApartmentRepo,
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				imgService *image.MockImage,
			) {
				apartmentRepo.On("GetApartmentByID", 1).Return(&models.Apartment{}, nil)
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				imgService.On("SaveImage", mock.Anything, mock.Anything, "test.jpg").Return("", errors.New("upload failed"))
			},
			expectedError: errors.New("failed to save image"),
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

			response, err := billService.CreateBill(context.Background(), tt.userID, tt.apartmentID, tt.req, tt.file, tt.fileHeader)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, 1, response["id"])
				assert.Equal(t, tt.req.TotalAmount, response["total_amount"])
				if tt.file != nil {
					assert.True(t, response["image_uploaded"].(bool))
				}
			}

			mockApartmentRepo.AssertExpectations(t)
			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockImageService.AssertExpectations(t)
		})
	}
}

func TestDivideBillByType(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		apartmentID int
		billType    models.BillType
		setupMocks  func(
			*repositories.MockUserApartmentRepository,
			*repositories.MockBillRepository,
			*repositories.MockPaymentRepository,
			*notification.MockNotification,
		)
		expectedError error
	}{
		{
			name:        "successful division by type",
			userID:      1,
			apartmentID: 1,
			billType:    models.WaterBill,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}, {BaseModel: models.BaseModel{ID: 3}}}, nil)
				billRepo.On("GetUndividedBillsByTypeAndApartment", 1, models.WaterBill).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, TotalAmount: 100, BillType: models.WaterBill},
				}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 2).Return(nil, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 3).Return(nil, nil)
				paymentRepo.On("CreatePayment", mock.Anything, mock.MatchedBy(func(p models.Payment) bool {
					return p.BillID == 1 && (p.UserID == 2 || p.UserID == 3) && p.PaymentStatus == models.Pending
				})).Return(1, nil).Times(2)
				notifService.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(2)
			},
			expectedError: nil,
		},
		{
			name:        "non-manager user",
			userID:      2,
			apartmentID: 1,
			billType:    models.WaterBill,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *repositories.MockPaymentRepository,
				_ *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 2, 1).Return(false, nil)
			},
			expectedError: errors.New("only apartment managers can divide bills"),
		},
		{
			name:        "no residents in apartment",
			userID:      1,
			apartmentID: 1,
			billType:    models.WaterBill,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *repositories.MockPaymentRepository,
				_ *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{}, nil)
			},
			expectedError: errors.New("no residents found in apartment"),
		},
		{
			name:        "no undivided bills of type",
			userID:      1,
			apartmentID: 1,
			billType:    models.WaterBill,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				_ *repositories.MockPaymentRepository,
				_ *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}}, nil)
				billRepo.On("GetUndividedBillsByTypeAndApartment", 1, models.WaterBill).Return([]models.Bill{}, nil)
			},
			expectedError: errors.New("no undivided bills of type water found"),
		},
		{
			name:        "payment creation failure",
			userID:      1,
			apartmentID: 1,
			billType:    models.WaterBill,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}, {BaseModel: models.BaseModel{ID: 3}}}, nil)
				billRepo.On("GetUndividedBillsByTypeAndApartment", 1, models.WaterBill).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, TotalAmount: 100, BillType: models.WaterBill},
				}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 2).Return(nil, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 3).Return(nil, nil)
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(0, errors.New("db error")).Times(2)
			},
			expectedError: nil,
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

			response, err := billService.DivideBillByType(context.Background(), tt.userID, tt.apartmentID, tt.billType)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.billType, response["bill_type"])

				if strings.Contains(tt.name, "failure") {
					assert.Contains(t, response, "warning")
				}
			}

			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			mockNotificationService.AssertExpectations(t)
		})
	}
}

func TestDivideAllBills(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		apartmentID int
		setupMocks  func(
			*repositories.MockUserApartmentRepository,
			*repositories.MockBillRepository,
			*repositories.MockPaymentRepository,
			*notification.MockNotification,
		)
		expectedError error
	}{
		{
			name:        "successful division of all bills",
			userID:      1,
			apartmentID: 1,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
				notifService *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}, {BaseModel: models.BaseModel{ID: 3}}}, nil)
				billRepo.On("GetUndividedBillsByApartment", 1).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, TotalAmount: 100, BillType: models.WaterBill},
					{BaseModel: models.BaseModel{ID: 2}, TotalAmount: 200, BillType: models.ElectricityBill},
				}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", mock.Anything, mock.Anything).Return(nil, nil)
				paymentRepo.On("CreatePayment", mock.Anything, mock.Anything).Return(1, nil).Times(4)
				notifService.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil).Times(4)
			},
			expectedError: nil,
		},
		{
			name:        "no undivided bills",
			userID:      1,
			apartmentID: 1,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				_ *repositories.MockPaymentRepository,
				_ *notification.MockNotification,
			) {
				userApartmentRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userApartmentRepo.On("GetResidentsInApartment", 1).Return([]models.User{{BaseModel: models.BaseModel{ID: 2}}}, nil)
				billRepo.On("GetUndividedBillsByApartment", 1).Return([]models.Bill{}, nil)
			},
			expectedError: errors.New("no undivided bills found in apartment"),
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

			response, err := billService.DivideAllBills(context.Background(), tt.userID, tt.apartmentID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Contains(t, response, "bill_types_processed")
				assert.Equal(t, 2, response["residents_count"])
			}

			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
			mockNotificationService.AssertExpectations(t)
		})
	}
}

func TestGetBillByID(t *testing.T) {
	tests := []struct {
		name          string
		billID        int
		setupMocks    func(*repositories.MockBillRepository, *image.MockImage)
		expectedError error
	}{
		{
			name:   "successful bill retrieval",
			billID: 1,
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(&models.Bill{
					BaseModel:   models.BaseModel{ID: 1},
					ApartmentID: 1,
					BillType:    models.WaterBill,
					TotalAmount: 100.50,
					ImageURL:    "image123",
				}, nil)
				imgService.On("GetImageURL", mock.Anything, "image123").Return("http://example.com/image123", nil)
			},
			expectedError: nil,
		},
		{
			name:   "bill not found",
			billID: 999,
			setupMocks: func(billRepo *repositories.MockBillRepository, _ *image.MockImage) {
				billRepo.On("GetBillByID", 999).Return(nil, errors.New("not found"))
			},
			expectedError: errors.New("failed to get bill"),
		},
		{
			name:   "image URL generation failure",
			billID: 1,
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(&models.Bill{
					BaseModel:   models.BaseModel{ID: 1},
					ApartmentID: 1,
					BillType:    models.WaterBill,
					TotalAmount: 100.50,
					ImageURL:    "image123",
				}, nil)
				imgService.On("GetImageURL", mock.Anything, "image123").Return("", errors.New("url generation failed"))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockBillRepo, mockImageService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)

			response, err := billService.GetBillByID(context.Background(), tt.billID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.billID, response["id"])

				if strings.Contains(tt.name, "image URL generation failure") {
					assert.Equal(t, "", response["image_url"])
				}
			}

			mockBillRepo.AssertExpectations(t)
			mockImageService.AssertExpectations(t)
		})
	}
}

func TestUpdateBill(t *testing.T) {
	tests := []struct {
		name            string
		billID          int
		apartmentID     int
		billType        string
		totalAmount     float64
		dueDate         string
		billingDeadline string
		description     string
		setupMocks      func(*repositories.MockBillRepository)
		expectedError   error
	}{
		{
			name:            "successful bill update",
			billID:          1,
			apartmentID:     1,
			billType:        "water",
			totalAmount:     100.50,
			dueDate:         "2023-12-31",
			billingDeadline: "2023-12-25",
			description:     "Updated bill",
			setupMocks: func(billRepo *repositories.MockBillRepository) {
				billRepo.On("UpdateBill", mock.Anything, mock.MatchedBy(func(b models.Bill) bool {
					return b.ID == 1 && b.ApartmentID == 1 && b.BillType == models.WaterBill
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:        "update failed",
			billID:      1,
			apartmentID: 1,
			billType:    "water",
			totalAmount: 100.50,
			dueDate:     "2023-12-31",
			setupMocks: func(billRepo *repositories.MockBillRepository) {
				billRepo.On("UpdateBill", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			expectedError: errors.New("failed to update bill"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockBillRepo)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)

			err := billService.UpdateBill(context.Background(), tt.billID, tt.apartmentID, tt.billType, tt.totalAmount, tt.dueDate, tt.billingDeadline, tt.description)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockBillRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteBill(t *testing.T) {
	tests := []struct {
		name          string
		billID        int
		setupMocks    func(*repositories.MockBillRepository, *image.MockImage)
		expectedError error
	}{
		{
			name:   "successful bill deletion",
			billID: 1,
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(&models.Bill{
					BaseModel: models.BaseModel{ID: 1},
					ImageURL:  "image123",
				}, nil)
				billRepo.On("DeleteBill", 1)
				imgService.On("DeleteImage", mock.Anything, "image123").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:   "bill not found",
			billID: 999,
			setupMocks: func(billRepo *repositories.MockBillRepository, _ *image.MockImage) {
				billRepo.On("GetBillByID", 999).Return(nil, errors.New("not found"))
			},
			expectedError: errors.New("failed to get bill"),
		},
		{
			name:   "image deletion failure",
			billID: 1,
			setupMocks: func(billRepo *repositories.MockBillRepository, imgService *image.MockImage) {
				billRepo.On("GetBillByID", 1).Return(&models.Bill{
					BaseModel: models.BaseModel{ID: 1},
					ImageURL:  "image123",
				}, nil)
				billRepo.On("DeleteBill", 1)
				imgService.On("DeleteImage", mock.Anything, "image123").Return(errors.New("delete failed"))
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockBillRepo := new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockBillRepo, mockImageService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)

			err := billService.DeleteBill(context.Background(), tt.billID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockBillRepo.AssertExpectations(t)
			mockImageService.AssertExpectations(t)
		})
	}
}

func TestPayBills(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		paymentIDs    []int
		idempotentKey string
		setupMocks    func(*repositories.MockPaymentRepository, *payment.MockPayment)
		expectedError error
	}{
		{
			name:          "successful payment",
			userID:        1,
			paymentIDs:    []int{1, 2},
			idempotentKey: "idemp123",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentService.On("PayBills", []int{1, 2}, "idemp123").Return(nil)
				paymentRepo.On("UpdatePaymentsStatus", mock.Anything, mock.MatchedBy(func(payments []models.Payment) bool {
					return len(payments) == 2 && payments[0].PaymentStatus == models.Paid
				})).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "payment service failure",
			userID:        1,
			paymentIDs:    []int{1},
			idempotentKey: "idemp123",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentService.On("PayBills", []int{1}, "idemp123").Return(errors.New("payment failed"))
			},
			expectedError: errors.New("payment failed"),
		},
		{
			name:          "status update failure",
			userID:        1,
			paymentIDs:    []int{1},
			idempotentKey: "idemp123",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentService.On("PayBills", []int{1}, "idemp123").Return(nil)
				paymentRepo.On("UpdatePaymentsStatus", mock.Anything, mock.Anything).Return(errors.New("update failed"))
			},
			expectedError: errors.New("failed to update payments status"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			_ = new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockPaymentRepo, mockPaymentService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)

			err := billService.PayBills(context.Background(), tt.userID, tt.paymentIDs, tt.idempotentKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockPaymentRepo.AssertExpectations(t)
			mockPaymentService.AssertExpectations(t)
		})
	}
}

func TestPayBatchBills(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		idempotentKey string
		setupMocks    func(*repositories.MockPaymentRepository, *payment.MockPayment)
		expectedError error
	}{
		{
			name:          "successful batch payment",
			userID:        1,
			idempotentKey: "idemp123",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, paymentService *payment.MockPayment) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{
					{BaseModel: models.BaseModel{ID: 1}, Amount: "100.50"},
					{BaseModel: models.BaseModel{ID: 2}, Amount: "200.75"},
				}, nil)
				paymentService.On("PayBills", []int{1, 2}, "idemp123").Return(nil)
				paymentRepo.On("UpdatePaymentsStatus", mock.Anything, mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "no pending payments",
			userID:        1,
			idempotentKey: "idemp123",
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository, _ *payment.MockPayment) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{}, nil)
			},
			expectedError: errors.New("no valid unpaid bills found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			_ = new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockPaymentRepo, mockPaymentService)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)

			response, err := billService.PayBatchBills(context.Background(), tt.userID, tt.idempotentKey)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Contains(t, response, "total_amount")
			}

			mockPaymentRepo.AssertExpectations(t)
			mockPaymentService.AssertExpectations(t)
		})
	}
}

func TestGetUnpaidBills(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		setupMocks    func(*repositories.MockPaymentRepository)
		expectedError error
	}{
		{
			name:   "successful retrieval of unpaid bills",
			userID: 1,
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{
					{BaseModel: models.BaseModel{ID: 1}, Amount: "100.50"},
					{BaseModel: models.BaseModel{ID: 2}, Amount: "200.75"},
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "no unpaid bills",
			userID: 1,
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return([]models.Payment{}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "database error",
			userID: 1,
			setupMocks: func(paymentRepo *repositories.MockPaymentRepository) {
				paymentRepo.On("GetPendingPaymentsByUser", 1).Return(nil, errors.New("db error"))
			},
			expectedError: errors.New("failed to get unpaid bills"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockPaymentService := new(payment.MockPayment)
			_ = new(repositories.MockBillRepository)
			mockImageService := new(image.MockImage)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockPaymentRepo)

			billService := services.NewBillService(
				nil,
				nil,
				nil,
				nil,
				mockPaymentRepo,
				mockImageService,
				mockPaymentService,
				mockNotificationService,
			)

			payments, err := billService.GetUnpaidBills(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, payments)
			} else {
				assert.NoError(t, err)
				if strings.Contains(tt.name, "no unpaid bills") {
					assert.Empty(t, payments)
				} else {
					assert.NotEmpty(t, payments)
				}
			}

			mockPaymentRepo.AssertExpectations(t)
		})
	}
}

func TestGetUserPaymentHistory(t *testing.T) {
	tests := []struct {
		name       string
		userID     int
		setupMocks func(
			*repositories.MockUserApartmentRepository,
			*repositories.MockBillRepository,
			*repositories.MockPaymentRepository,
		)
		expectedError error
	}{
		{
			name:   "successful retrieval of payment history",
			userID: 1,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				billRepo *repositories.MockBillRepository,
				paymentRepo *repositories.MockPaymentRepository,
			) {
				userApartmentRepo.On("GetAllApartmentsForAResident", 1).Return([]models.Apartment{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentName: "Test Apartment"},
				}, nil)
				billRepo.On("GetBillsByApartmentID", 1).Return([]models.Bill{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentID: 1, BillType: models.WaterBill},
				}, nil)
				paymentRepo.On("GetPaymentByBillAndUser", 1, 1).Return(&models.Payment{
					BaseModel: models.BaseModel{ID: 1},
					BillID:    1,
					UserID:    1,
					Amount:    "100.50",
				}, nil)
			},
			expectedError: nil,
		},
		{
			name:   "no payment history",
			userID: 1,
			setupMocks: func(
				userApartmentRepo *repositories.MockUserApartmentRepository,
				_ *repositories.MockBillRepository,
				_ *repositories.MockPaymentRepository,
			) {
				userApartmentRepo.On("GetAllApartmentsForAResident", 1).Return([]models.Apartment{}, nil)
			},
			expectedError: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockUserApartmentRepo := new(repositories.MockUserApartmentRepository)
			mockBillRepo := new(repositories.MockBillRepository)
			mockPaymentRepo := new(repositories.MockPaymentRepository)
			mockImageService := new(image.MockImage)
			mockPaymentService := new(payment.MockPayment)
			mockNotificationService := new(notification.MockNotification)

			tt.setupMocks(mockUserApartmentRepo, mockBillRepo, mockPaymentRepo)

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

			history, err := billService.GetUserPaymentHistory(context.Background(), tt.userID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError.Error())
				assert.Nil(t, history)
			} else {
				assert.NoError(t, err)
				if strings.Contains(tt.name, "no payment history") {
					assert.Empty(t, history)
				} else {
					assert.NotEmpty(t, history)
					assert.Equal(t, "Test Apartment", history[0].ApartmentName)
				}
			}

			mockUserApartmentRepo.AssertExpectations(t)
			mockBillRepo.AssertExpectations(t)
			mockPaymentRepo.AssertExpectations(t)
		})
	}
}
