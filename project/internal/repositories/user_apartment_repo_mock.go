package repositories

import (
	"context"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/mock"
)

type MockUserApartmentRepository struct {
	mock.Mock
}

func (m *MockUserApartmentRepository) CreateUserApartment(ctx context.Context, userApartment models.User_apartment) error {
	args := m.Called(ctx, userApartment)
	return args.Error(0)
}

func (m *MockUserApartmentRepository) GetResidentsInApartment(apartmentID int) ([]models.User, error) {
	args := m.Called(apartmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserApartmentRepository) GetUserApartmentByID(userID, apartmentID int) (*models.User_apartment, error) {
	args := m.Called(userID, apartmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User_apartment), args.Error(1)
}

func (m *MockUserApartmentRepository) UpdateUserApartment(ctx context.Context, userApartment models.User_apartment) error {
	args := m.Called(ctx, userApartment)
	return args.Error(0)
}

func (m *MockUserApartmentRepository) DeleteUserApartment(userID, apartmentID int) error {
	args := m.Called(userID, apartmentID)
	return args.Error(0)
}

func (m *MockUserApartmentRepository) DeleteUserFromApartments(userID int) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserApartmentRepository) GetAllApartmentsForAResident(residentID int) ([]models.Apartment, error) {
	args := m.Called(residentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Apartment), args.Error(1)
}

func (m *MockUserApartmentRepository) IsUserManagerOfApartment(ctx context.Context, userID, apartmentID int) (bool, error) {
	args := m.Called(ctx, userID, apartmentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserApartmentRepository) IsUserInApartment(ctx context.Context, userID, apartmentID int) (bool, error) {
	args := m.Called(ctx, userID, apartmentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserApartmentRepository) DeleteApartmentFromUserApartments(apartmentID int) error {
	args := m.Called(apartmentID)
	return args.Error(0)
}
