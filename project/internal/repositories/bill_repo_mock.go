package repositories

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockBillRepository struct {
	mock.Mock
}

func (m *MockBillRepository) CreateBill(ctx context.Context, bill models.Bill) (int, error) {
	args := m.Called(ctx, bill)
	return args.Int(0), args.Error(1)
}

func (m *MockBillRepository) GetBillByID(id int) (*models.Bill, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Bill), args.Error(1)
}

func (m *MockBillRepository) GetBillsByApartmentID(apartmentID int) ([]models.Bill, error) {
	args := m.Called(apartmentID)
	return args.Get(0).([]models.Bill), args.Error(1)
}

func (m *MockBillRepository) UpdateBill(ctx context.Context, bill models.Bill) error {
	args := m.Called(ctx, bill)
	return args.Error(0)
}

func (m *MockBillRepository) DeleteBill(id int) {
	m.Called(id)
}

func (m *MockBillRepository) GetPaymentByBillAndUser(billID, userID int) (*models.Payment, error) {
	args := m.Called(billID, userID)
	return args.Get(0).(*models.Payment), args.Error(1)
}
