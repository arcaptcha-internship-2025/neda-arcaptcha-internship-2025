package payment

import (
	"github.com/stretchr/testify/mock"
)

type MockPayment struct {
	mock.Mock
}

func (m *MockPayment) PayBills(billIDs []int, idempotentKey string) error {
	args := m.Called(billIDs, idempotentKey)
	return args.Error(0)
}

func NewMockPayment() *MockPayment {
	return &MockPayment{}
}

func (m *MockPayment) ExpectPayBills(billIDs []int, idempotentKey string, returnError error) *mock.Call {
	return m.On("PayBills", billIDs, idempotentKey).Return(returnError)
}

func (m *MockPayment) ExpectPayBillsOnce(billIDs []int, idempotentKey string, returnError error) *mock.Call {
	return m.On("PayBills", billIDs, idempotentKey).Return(returnError).Once()
}

func (m *MockPayment) ExpectPayBillsTimes(times int, billIDs []int, idempotentKey string, returnError error) *mock.Call {
	return m.On("PayBills", billIDs, idempotentKey).Return(returnError).Times(times)
}

func (m *MockPayment) ExpectPayBillsWithAnyArgs(returnError error) *mock.Call {
	return m.On("PayBills", mock.Anything, mock.Anything).Return(returnError)
}

func (m *MockPayment) ExpectPayBillsWithAnyArgsOnce(returnError error) *mock.Call {
	return m.On("PayBills", mock.Anything, mock.Anything).Return(returnError).Once()
}

func (m *MockPayment) ExpectPayBillsWithSpecificBillIDs(billIDs []int, returnError error) *mock.Call {
	return m.On("PayBills", billIDs, mock.AnythingOfType("string")).Return(returnError)
}

func (m *MockPayment) ExpectPayBillsWithSpecificIdempotentKey(idempotentKey string, returnError error) *mock.Call {
	return m.On("PayBills", mock.AnythingOfType("[]int"), idempotentKey).Return(returnError)
}

func (m *MockPayment) ExpectNoPayBillsCalls(t mock.TestingT) {
	m.AssertNotCalled(t, "PayBills")
}

func (m *MockPayment) ExpectPayBillsCalledWith(t mock.TestingT, billIDs []int, idempotentKey string) {
	m.AssertCalled(t, "PayBills", billIDs, idempotentKey)
}

func (m *MockPayment) ExpectPayBillsCalledOnceWith(t mock.TestingT, billIDs []int, idempotentKey string) {
	m.AssertNumberOfCalls(t, "PayBills", 1)
	m.AssertCalled(t, "PayBills", billIDs, idempotentKey)
}
