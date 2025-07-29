package notification

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type mockNotificationService struct {
	mock.Mock
}

func (m *mockNotificationService) SendNotification(ctx context.Context, userID int, message string) error {
	args := m.Called(ctx, userID, message)
	return args.Error(0)
}

func (m *mockNotificationService) SendInvitation(ctx context.Context, inv models.InvitationLink) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *mockNotificationService) SendBillNotification(ctx context.Context, userID int, bill models.Bill, amount float64) error {
	args := m.Called(ctx, userID, bill, amount)
	return args.Error(0)
}

func (m *mockNotificationService) ListenForUpdates(ctx context.Context) {
	m.Called(ctx)
}
