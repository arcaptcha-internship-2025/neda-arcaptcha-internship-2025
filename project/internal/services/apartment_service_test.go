package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateApartment(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		apartmentName string
		address       string
		unitsCount    int
		mockSetup     func(*repositories.MockApartmentRepo, *repositories.MockUserApartmentRepository)
		expectedID    int
		expectedError string
	}{
		{
			name:          "successful apartment creation",
			userID:        1,
			apartmentName: "Sunny Apartments",
			address:       "123 Main St",
			unitsCount:    10,
			mockSetup: func(aptRepo *repositories.MockApartmentRepo, userAptRepo *repositories.MockUserApartmentRepository) {
				aptRepo.On("CreateApartment", mock.Anything, mock.MatchedBy(func(apt models.Apartment) bool {
					return apt.ApartmentName == "Sunny Apartments" &&
						apt.Address == "123 Main St" &&
						apt.UnitsCount == 10 &&
						apt.ManagerID == 1
				})).Return(1, nil)
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.MatchedBy(func(ua models.User_apartment) bool {
					return ua.UserID == 1 && ua.ApartmentID == 1 && ua.IsManager
				})).Return(nil)
			},
			expectedID: 1,
		},
		{
			name:          "failed to create apartment",
			userID:        1,
			apartmentName: "Sunny Apartments",
			address:       "123 Main St",
			unitsCount:    10,
			mockSetup: func(aptRepo *repositories.MockApartmentRepo, userAptRepo *repositories.MockUserApartmentRepository) {
				aptRepo.On("CreateApartment", mock.Anything, mock.Anything).Return(0, errors.New("database error"))
			},
			expectedError: "failed to create apartment",
		},
		{
			name:          "failed to assign manager",
			userID:        1,
			apartmentName: "Sunny Apartments",
			address:       "123 Main St",
			unitsCount:    10,
			mockSetup: func(aptRepo *repositories.MockApartmentRepo, userAptRepo *repositories.MockUserApartmentRepository) {
				aptRepo.On("CreateApartment", mock.Anything, mock.Anything).Return(1, nil)
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.Anything).Return(errors.New("assignment error"))
			},
			expectedError: "failed to assign manager to apartment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockAptRepo, mockUserAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			id, err := service.CreateApartment(context.Background(), tt.userID, tt.apartmentName, tt.address, tt.unitsCount)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Equal(t, 0, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}

			mockAptRepo.AssertExpectations(t)
			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestGetApartmentByID(t *testing.T) {
	tests := []struct {
		name           string
		id             int
		managerID      int
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo)
		expectedResult *models.Apartment
		expectedError  string
	}{
		{
			name:      "successful get apartment",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("GetApartmentByID", 1).Return(&models.Apartment{
					BaseModel:     models.BaseModel{ID: 1},
					ApartmentName: "Sunny Apartments",
					Address:       "123 Main St",
					UnitsCount:    10,
					ManagerID:     1,
				}, nil)
			},
			expectedResult: &models.Apartment{
				BaseModel:     models.BaseModel{ID: 1},
				ApartmentName: "Sunny Apartments",
				Address:       "123 Main St",
				UnitsCount:    10,
				ManagerID:     1,
			},
		},
		{
			name:      "not manager of apartment",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedError: "",
		},
		{
			name:      "error checking manager status",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, errors.New("database error"))
			},
			expectedError: "",
		},
		{
			name:      "apartment not found",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("GetApartmentByID", 1).Return((*models.Apartment)(nil), errors.New("not found"))
			},
			expectedError: "failed to get apartment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo, mockAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			apartment, err := service.GetApartmentByID(context.Background(), tt.id, tt.managerID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, apartment)
			} else if tt.expectedError == "" && tt.name != "successful get apartment" {

				assert.Error(t, err)
				assert.Equal(t, "", err.Error())
				assert.Nil(t, apartment)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, apartment)
			}

			mockUserAptRepo.AssertExpectations(t)
			mockAptRepo.AssertExpectations(t)
		})
	}
}

func TestGetResidentsInApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    int
		managerID      int
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedResult []models.User
		expectedError  string
	}{
		{
			name:        "successful get residents",
			apartmentID: 1,
			managerID:   1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userAptRepo.On("GetResidentsInApartment", 1).Return([]models.User{
					{BaseModel: models.BaseModel{ID: 1}, Username: "user1"},
					{BaseModel: models.BaseModel{ID: 2}, Username: "user2"},
				}, nil)
			},
			expectedResult: []models.User{
				{BaseModel: models.BaseModel{ID: 1}, Username: "user1"},
				{BaseModel: models.BaseModel{ID: 2}, Username: "user2"},
			},
		},
		{
			name:        "not manager of apartment",
			apartmentID: 1,
			managerID:   1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedError: "",
		},
		{
			name:        "error checking manager status",
			apartmentID: 1,
			managerID:   1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, errors.New("database error"))
			},
			expectedError: "",
		},
		{
			name:        "failed to get residents",
			apartmentID: 1,
			managerID:   1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userAptRepo.On("GetResidentsInApartment", 1).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get residents",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			residents, err := service.GetResidentsInApartment(context.Background(), tt.apartmentID, tt.managerID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, residents)
			} else if tt.expectedError == "" && tt.name != "successful get residents" {
				// For empty error messages (authorization failures)
				assert.Error(t, err)
				assert.Equal(t, "", err.Error())
				assert.Nil(t, residents)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, residents)
			}

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateApartment(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		apartmentName string
		address       string
		unitsCount    int
		managerID     int
		mockSetup     func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo)
		expectedError string
	}{
		{
			name:          "successful update",
			id:            1,
			apartmentName: "Updated Name",
			address:       "Updated Address",
			unitsCount:    20,
			managerID:     1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("UpdateApartment", mock.Anything, mock.MatchedBy(func(apt models.Apartment) bool {
					return apt.ID == 1 &&
						apt.ApartmentName == "Updated Name" &&
						apt.Address == "Updated Address" &&
						apt.UnitsCount == 20 &&
						apt.ManagerID == 1
				})).Return(nil)
			},
		},
		{
			name:          "not manager of apartment",
			id:            1,
			apartmentName: "Updated Name",
			address:       "Updated Address",
			unitsCount:    20,
			managerID:     1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedError: "", // Empty error message as per the service implementation
		},
		{
			name:          "error checking manager status",
			id:            1,
			apartmentName: "Updated Name",
			address:       "Updated Address",
			unitsCount:    20,
			managerID:     1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, errors.New("database error"))
			},
			expectedError: "", // Empty error message as per the service implementation
		},
		{
			name:          "failed to update",
			id:            1,
			apartmentName: "Updated Name",
			address:       "Updated Address",
			unitsCount:    20,
			managerID:     1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("UpdateApartment", mock.Anything, mock.Anything).Return(errors.New("database error"))
			},
			expectedError: "failed to update apartment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo, mockAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			err := service.UpdateApartment(context.Background(), tt.id, tt.apartmentName, tt.address, tt.unitsCount, tt.managerID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else if tt.expectedError == "" && tt.name != "successful update" {
				// For empty error messages (authorization failures)
				assert.Error(t, err)
				assert.Equal(t, "", err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserAptRepo.AssertExpectations(t)
			mockAptRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteApartment(t *testing.T) {
	tests := []struct {
		name          string
		id            int
		managerID     int
		mockSetup     func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo)
		expectedError string
	}{
		{
			name:      "successful delete",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("DeleteApartment", 1).Return(nil)
				userAptRepo.On("DeleteApartmentFromUserApartments", 1).Return(nil)
			},
		},
		{
			name:      "not manager of apartment",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedError: "",
		},
		{
			name:      "error checking manager status",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, errors.New("database error"))
			},
			expectedError: "",
		},
		{
			name:      "failed to delete apartment",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("DeleteApartment", 1).Return(errors.New("database error"))
			},
			expectedError: "failed to delete apartment",
		},
		{
			name:      "failed to delete from user apartments",
			id:        1,
			managerID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("DeleteApartment", 1).Return(nil)
				userAptRepo.On("DeleteApartmentFromUserApartments", 1).Return(errors.New("database error"))
			},
			expectedError: "failed to remove apartment from user apartments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo, mockAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			err := service.DeleteApartment(context.Background(), tt.id, tt.managerID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else if tt.expectedError == "" && tt.name != "successful delete" {
				assert.Error(t, err)
				assert.Equal(t, "", err.Error())
			} else {
				assert.NoError(t, err)
			}

			mockUserAptRepo.AssertExpectations(t)
			mockAptRepo.AssertExpectations(t)
		})
	}
}

func TestInviteUserToApartment(t *testing.T) {
	tests := []struct {
		name             string
		managerID        int
		apartmentID      int
		telegramUsername string
		mockSetup        func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification)
		expectedResult   map[string]interface{}
		expectedError    string
	}{
		{
			name:             "successful invitation",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, errors.New("not in apartment"))
				inviteRepo.On("CreateInvitation", mock.Anything, 2, 1, 1).Return("invite123", nil)
				notif.On("SendInvitation", mock.Anything, mock.Anything, 1, "testuser").Return(nil)
			},
			expectedResult: map[string]interface{}{
				"status":     "invitation sent",
				"expires_at": mock.Anything,
			},
		},
		{
			name:             "not manager of apartment",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedError: "only apartment managers can send invitations",
		},
		{
			name:             "error verifying manager status",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, errors.New("database error"))
			},
			expectedError: "failed to verify apartment manager status",
		},
		{
			name:             "user not found",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(nil, errors.New("not found"))
			},
			expectedError: "user with this Telegram username not found",
		},
		{
			name:             "user already resident",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(true, nil)
			},
			expectedError: "user is already a resident of this apartment",
		},
		{
			name:             "failed to create invitation",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, errors.New("not in apartment"))
				inviteRepo.On("CreateInvitation", mock.Anything, 2, 1, 1).Return("", errors.New("creation failed"))
			},
			expectedError: "failed to created invitation",
		},
		{
			name:             "failed to send notification",
			managerID:        1,
			apartmentID:      1,
			telegramUsername: "testuser",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, errors.New("not in apartment"))
				inviteRepo.On("CreateInvitation", mock.Anything, 2, 1, 1).Return("invite123", nil)
				notif.On("SendInvitation", mock.Anything, mock.Anything, 1, "testuser").Return(errors.New("send failed"))
			},
			expectedError: "invitation created but failed to send notification",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo, mockUserRepo, mockInviteRepo, mockNotif)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			result, err := service.InviteUserToApartment(context.Background(), tt.managerID, tt.apartmentID, tt.telegramUsername)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, "invitation sent", result["status"])
				assert.NotNil(t, result["expires_at"])
			}

			mockUserAptRepo.AssertExpectations(t)
			mockUserRepo.AssertExpectations(t)
			mockInviteRepo.AssertExpectations(t)
			mockNotif.AssertExpectations(t)
		})
	}
}

func TestJoinApartment(t *testing.T) {
	tests := []struct {
		name           string
		userID         int
		invitationCode string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification)
		expectedResult map[string]interface{}
		expectedError  string
	}{
		{
			name:           "successful join",
			userID:         1,
			invitationCode: "validcode",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "validcode").Return(1, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 1, 1).Return(false, errors.New("not in apartment"))
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.MatchedBy(func(ua models.User_apartment) bool {
					return ua.UserID == 1 && ua.ApartmentID == 1 && !ua.IsManager
				})).Return(nil)
				notif.On("SendNotification", mock.Anything, 1, "You joined apartment 1").Return(nil)
			},
			expectedResult: map[string]interface{}{
				"status": "joined apartment",
			},
		},
		{
			name:           "invalid invitation code",
			userID:         1,
			invitationCode: "invalidcode",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "invalidcode").Return(0, errors.New("invalid code"))
			},
			expectedError: "invalid code",
		},
		{
			name:           "already a resident",
			userID:         1,
			invitationCode: "validcode",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "validcode").Return(1, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 1, 1).Return(true, nil)
			},
			expectedError: "you are already a resident of this apartment",
		},
		{
			name:           "failed to join apartment",
			userID:         1,
			invitationCode: "validcode",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "validcode").Return(1, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 1, 1).Return(false, errors.New("not in apartment"))
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.MatchedBy(func(ua models.User_apartment) bool {
					return ua.UserID == 1 && ua.ApartmentID == 1 && !ua.IsManager
				})).Return(errors.New("failed to create"))
			},
			expectedError: "failed to join apartment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo, mockInviteRepo, mockNotif)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			result, err := service.JoinApartment(context.Background(), tt.userID, tt.invitationCode)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}

			mockUserAptRepo.AssertExpectations(t)
			mockInviteRepo.AssertExpectations(t)
			mockNotif.AssertExpectations(t)
		})
	}
}

func TestLeaveApartment(t *testing.T) {
	tests := []struct {
		name          string
		userID        int
		apartmentID   int
		mockSetup     func(*repositories.MockUserApartmentRepository)
		expectedError string
	}{
		{
			name:        "successful leave",
			userID:      1,
			apartmentID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("DeleteUserApartment", 1, 1).Return(nil)
			},
		},
		{
			name:        "failed to leave",
			userID:      1,
			apartmentID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("DeleteUserApartment", 1, 1).Return(errors.New("database error"))
			},
			expectedError: "failed to leave apartment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			err := service.LeaveApartment(context.Background(), tt.userID, tt.apartmentID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestGetAllApartmentsForResident(t *testing.T) {
	tests := []struct {
		name           string
		residentID     int
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedResult []models.Apartment
		expectedError  string
	}{
		{
			name:       "successful get apartments",
			residentID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("GetAllApartmentsForAResident", 1).Return([]models.Apartment{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentName: "Apt 1"},
					{BaseModel: models.BaseModel{ID: 2}, ApartmentName: "Apt 2"},
				}, nil)
			},
			expectedResult: []models.Apartment{
				{BaseModel: models.BaseModel{ID: 1}, ApartmentName: "Apt 1"},
				{BaseModel: models.BaseModel{ID: 2}, ApartmentName: "Apt 2"},
			},
		},
		{
			name:       "failed to get apartments",
			residentID: 1,
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("GetAllApartmentsForAResident", 1).Return(nil, errors.New("database error"))
			},
			expectedError: "failed to get apartments",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)

			apartments, err := service.GetAllApartmentsForResident(context.Background(), tt.residentID)

			if tt.expectedError != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Nil(t, apartments)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedResult, apartments)
			}

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}
