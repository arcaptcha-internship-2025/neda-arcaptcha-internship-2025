package repositories

import (
	"context"

	"github.com/stretchr/testify/mock"
)

type MockInviteLinkRepository struct {
	mock.Mock
}

func (m *MockInviteLinkRepository) CreateInvitation(ctx context.Context, userID, apartmentID, managerID int) (string, error) {
	args := m.Called(ctx, userID, apartmentID, managerID)
	return args.String(0), args.Error(1)
}

func (m *MockInviteLinkRepository) ValidateAndConsumeInvitation(ctx context.Context, code string) (int, error) {
	args := m.Called(ctx, code)
	return args.Int(0), args.Error(1)
}

func NewMockInviteLinkRepository() *MockInviteLinkRepository {
	return &MockInviteLinkRepository{}
}

func (m *MockInviteLinkRepository) ExpectCreateInvitation(ctx context.Context, userID, apartmentID, managerID int, returnCode string, returnError error) *mock.Call {
	return m.On("CreateInvitation", ctx, userID, apartmentID, managerID).Return(returnCode, returnError)
}

func (m *MockInviteLinkRepository) ExpectValidateAndConsumeInvitation(ctx context.Context, code string, returnApartmentID int, returnError error) *mock.Call {
	return m.On("ValidateAndConsumeInvitation", ctx, code).Return(returnApartmentID, returnError)
}
