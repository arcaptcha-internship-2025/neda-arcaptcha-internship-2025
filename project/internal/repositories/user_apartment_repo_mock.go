package repositories

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type mockUserApartmentRepo struct {
	mock.Mock
}

func (m *mockUserApartmentRepo) CreateUserApartment(ctx context.Context, userApartment models.User_apartment) error {
	args := m.Called(ctx, userApartment)
	return args.Error(0)
}

func (m *mockUserApartmentRepo) GetResidentsInApartment(apartmentID int) ([]models.User, error) {
	args := m.Called(apartmentID)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserApartmentRepo) GetUserApartmentByID(userID, apartmentID int) (*models.User_apartment, error) {
	args := m.Called(userID, apartmentID)
	return args.Get(0).(*models.User_apartment), args.Error(1)
}

func (m *mockUserApartmentRepo) UpdateUserApartment(ctx context.Context, userApartment models.User_apartment) error {
	args := m.Called(ctx, userApartment)
	return args.Error(0)
}

func (m *mockUserApartmentRepo) DeleteUserApartment(userID, apartmentID int) error {
	args := m.Called(userID, apartmentID)
	return args.Error(0)
}

func (m *mockUserApartmentRepo) GetAllApartmentsForAResident(residentID int) ([]models.Apartment, error) {
	args := m.Called(residentID)
	return args.Get(0).([]models.Apartment), args.Error(1)
}

func (m *mockUserApartmentRepo) IsUserManagerOfApartment(ctx context.Context, userID, apartmentID int) (bool, error) {
	args := m.Called(ctx, userID, apartmentID)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserApartmentRepo) IsUserInApartment(ctx context.Context, userID, apartmentID int) (bool, error) {
	args := m.Called(ctx, userID, apartmentID)
	return args.Bool(0), args.Error(1)
}
