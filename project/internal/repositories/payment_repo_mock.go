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
