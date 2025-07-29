package repositories

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type mockInviteLinkRepo struct {
	mock.Mock
}

func (m *mockInviteLinkRepo) CreateInvitation(ctx context.Context, inv models.InvitationLink) error {
	args := m.Called(ctx, inv)
	return args.Error(0)
}

func (m *mockInviteLinkRepo) GetInvitationByToken(ctx context.Context, token string) (*models.InvitationLink, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*models.InvitationLink), args.Error(1)
}
