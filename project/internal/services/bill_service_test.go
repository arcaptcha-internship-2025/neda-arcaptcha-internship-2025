package services

import (
	"context"
	"errors"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/image"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/payment"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

			billService := NewBillService(
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
