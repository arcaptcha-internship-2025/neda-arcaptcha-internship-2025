package payment

import "github.com/stretchr/testify/mock"

type MockPaymentService struct {
	mock.Mock
}

func (m *MockPaymentService) PayBills(billIDs []int) error {
	args := m.Called(billIDs)
	return args.Error(0)
}
