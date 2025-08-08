package repositories

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockPaymentRepository struct {
	mock.Mock
}

func (m *MockPaymentRepository) CreatePayment(ctx context.Context, payment models.Payment) (int, error) {
	args := m.Called(ctx, payment)
	return args.Int(0), args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentByID(id int) (*models.Payment, error) {
	args := m.Called(id)
	if p, ok := args.Get(0).(*models.Payment); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentByBillAndUser(billID, userID int) (*models.Payment, error) {
	args := m.Called(billID, userID)
	if p, ok := args.Get(0).(*models.Payment); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentsByUser(userID int) ([]models.Payment, error) {
	args := m.Called(userID)
	if p, ok := args.Get(0).([]models.Payment); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) GetPendingPaymentsByUser(userID int) ([]models.Payment, error) {
	args := m.Called(userID)
	if p, ok := args.Get(0).([]models.Payment); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) GetPaymentsByBill(billID int) ([]models.Payment, error) {
	args := m.Called(billID)
	if p, ok := args.Get(0).([]models.Payment); ok {
		return p, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockPaymentRepository) UpdatePaymentStatus(ctx context.Context, payment models.Payment) error {
	args := m.Called(ctx, payment)
	return args.Error(0)
}

func (m *MockPaymentRepository) UpdatePaymentsStatus(ctx context.Context, payments []models.Payment) error {
	args := m.Called(ctx, payments)
	return args.Error(0)
}

func (m *MockPaymentRepository) DeletePayment(id int) error {
	args := m.Called(id)
	return args.Error(0)
}
