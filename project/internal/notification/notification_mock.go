package notification

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockNotification struct {
	mock.Mock
}

func (m *MockNotification) SendNotification(ctx context.Context, userID int, message string) error {
	args := m.Called(ctx, userID, message)
	return args.Error(0)
}

func (m *MockNotification) SendInvitation(ctx context.Context, inviteURL string, apartmentID int, receiverUsername string) error {
	args := m.Called(ctx, inviteURL, apartmentID, receiverUsername)
	return args.Error(0)
}

func (m *MockNotification) SendBillNotification(ctx context.Context, userID int, bill models.Bill, amount float64) error {
	args := m.Called(ctx, userID, bill, amount)
	return args.Error(0)
}

func (m *MockNotification) ListenForUpdates(ctx context.Context) {
	m.Called(ctx)
}

func NewMockNotification() *MockNotification {
	return &MockNotification{}
}

func (m *MockNotification) ExpectSendNotification(ctx context.Context, userID int, message string, returnError error) *mock.Call {
	return m.On("SendNotification", ctx, userID, message).Return(returnError)
}

func (m *MockNotification) ExpectSendInvitation(ctx context.Context, inviteURL string, apartmentID int, receiverUsername string, returnError error) *mock.Call {
	return m.On("SendInvitation", ctx, inviteURL, apartmentID, receiverUsername).Return(returnError)
}

func (m *MockNotification) ExpectSendBillNotification(ctx context.Context, userID int, bill models.Bill, amount float64, returnError error) *mock.Call {
	return m.On("SendBillNotification", ctx, userID, bill, amount).Return(returnError)
}

func (m *MockNotification) ExpectListenForUpdates(ctx context.Context) *mock.Call {
	return m.On("ListenForUpdates", ctx)
}

func (m *MockNotification) ExpectSendNotificationTimes(times int, ctx context.Context, userID int, message string, returnError error) *mock.Call {
	return m.On("SendNotification", ctx, userID, message).Return(returnError).Times(times)
}

func (m *MockNotification) ExpectSendInvitationTimes(times int, ctx context.Context, inviteURL string, apartmentID int, receiverUsername string, returnError error) *mock.Call {
	return m.On("SendInvitation", ctx, inviteURL, apartmentID, receiverUsername).Return(returnError).Times(times)
}

func (m *MockNotification) ExpectSendBillNotificationTimes(times int, ctx context.Context, userID int, bill models.Bill, amount float64, returnError error) *mock.Call {
	return m.On("SendBillNotification", ctx, userID, bill, amount).Return(returnError).Times(times)
}

func (m *MockNotification) ExpectAnyNotificationCall(returnError error) {
	m.On("SendNotification", mock.Anything, mock.Anything, mock.Anything).Maybe().Return(returnError)
	m.On("SendInvitation", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(returnError)
	m.On("SendBillNotification", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe().Return(returnError)
}

func (m *MockNotification) ExpectNoNotificationCalls(t mock.TestingT) {
	m.AssertNotCalled(t, "SendNotification")
	m.AssertNotCalled(t, "SendInvitation")
	m.AssertNotCalled(t, "SendBillNotification")
}
