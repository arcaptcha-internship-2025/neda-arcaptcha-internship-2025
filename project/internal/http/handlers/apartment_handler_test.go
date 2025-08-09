package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/notification"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/services"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateApartment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockApartmentRepo)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful apartment creation",
			requestBody: map[string]interface{}{
				"apartment_name": "Sunny Apartments",
				"address":        "123 Main St",
				"units_count":    10,
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, aptRepo *repositories.MockApartmentRepo) {
				aptRepo.On("CreateApartment", mock.Anything, mock.Anything).Return(1, nil)
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody:   map[string]interface{}{"id": float64(1)},
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid_field": "value",
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, aptRepo *repositories.MockApartmentRepo) {
				//no mocks needed since validation fails before repository calls
			},
			expectedStatus: http.StatusBadRequest,
		},

		{
			name: "failed to create apartment",
			requestBody: map[string]interface{}{
				"apartment_name": "Sunny Apartments",
				"address":        "123 Main St",
				"units_count":    10,
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, aptRepo *repositories.MockApartmentRepo) {
				aptRepo.On("CreateApartment", mock.Anything, mock.Anything).Return(0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAptRepo := new(repositories.MockApartmentRepo)
			mockUserRepo := new(repositories.MockUserRepository)
			mockUserAptRepo := new(repositories.MockUserApartmentRepository)
			mockInviteRepo := new(repositories.MockInviteLinkRepository)
			mockNotif := new(notification.MockNotification)

			tt.mockSetup(mockUserAptRepo, mockUserRepo, mockAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)
			handler := NewApartmentHandler(service)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/apartments", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.CreateApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedBody != nil {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)
			}

			mockAptRepo.AssertExpectations(t)
			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestGetApartmentByID(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo)
		expectedStatus int
	}{
		{
			name:        "successful get apartment",
			queryParams: "id=1",
			userID:      "1",
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
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid apartment id",
			queryParams:    "id=invalid",
			userID:         "1",
			mockSetup:      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "not manager of apartment",
			queryParams: "id=1",
			userID:      "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedStatus: http.StatusNotFound,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("GET", "/apartments?"+tt.queryParams, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.GetApartmentByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockAptRepo.AssertExpectations(t)
			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestGetResidentsInApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedStatus int
	}{
		{
			name:        "successful get residents",
			apartmentID: "1",
			userID:      "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userAptRepo.On("GetResidentsInApartment", 1).Return([]models.User{
					{BaseModel: models.BaseModel{ID: 1}, Username: "user1"},
					{BaseModel: models.BaseModel{ID: 2}, Username: "user2"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid apartment id",
			apartmentID:    "invalid",
			userID:         "1",
			mockSetup:      func(*repositories.MockUserApartmentRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "not manager of apartment",
			apartmentID: "1",
			userID:      "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedStatus: http.StatusInternalServerError,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("GET", "/apartments/"+tt.apartmentID+"/residents", nil)
			req.SetPathValue("apartment_id", tt.apartmentID)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.GetResidentsInApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestInviteUserToApartment(t *testing.T) {
	tests := []struct {
		name             string
		apartmentID      string
		telegramUsername string
		userID           string
		mockSetup        func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification)
		expectedStatus   int
	}{
		{
			name:             "successful invitation",
			apartmentID:      "1",
			telegramUsername: "testuser",
			userID:           "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, errors.New("not in apartment"))
				inviteRepo.On("CreateInvitation", mock.Anything, 2, 1, 1).Return("invite123", nil)
				notif.On("SendInvitation", mock.Anything, mock.Anything, 1, "testuser").Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:             "user already resident",
			apartmentID:      "1",
			telegramUsername: "testuser",
			userID:           "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(true, nil)
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:             "invalid apartment id",
			apartmentID:      "invalid",
			telegramUsername: "testuser",
			userID:           "1",
			mockSetup: func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification) {
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:             "empty telegram username",
			apartmentID:      "1",
			telegramUsername: "",
			userID:           "1",
			mockSetup: func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification) {
			},
			expectedStatus: http.StatusBadRequest,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("POST", "/apartments/"+tt.apartmentID+"/invite/"+tt.telegramUsername, nil)
			req.SetPathValue("apartment_id", tt.apartmentID)
			req.SetPathValue("telegram_username", tt.telegramUsername)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.InviteUserToApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

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
		invitationCode string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification)
		expectedStatus int
	}{
		{
			name:           "successful join",
			invitationCode: "validcode",
			userID:         "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "validcode").Return(1, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 1, 1).Return(false, errors.New("not in apartment"))
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.Anything).Return(nil)
				notif.On("SendNotification", mock.Anything, 1, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid invitation code",
			invitationCode: "invalidcode",
			userID:         "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "invalidcode").Return(0, errors.New("invalid code"))
			},
			expectedStatus: http.StatusInternalServerError,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("POST", "/apartments/join/"+tt.invitationCode, nil)
			req.SetPathValue("invitation_code", tt.invitationCode)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.JoinApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockUserAptRepo.AssertExpectations(t)
			mockInviteRepo.AssertExpectations(t)
			mockNotif.AssertExpectations(t)
		})
	}
}

func TestLeaveApartment(t *testing.T) {
	tests := []struct {
		name           string
		queryParam     string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedStatus int
	}{
		{
			name:       "successful leave",
			queryParam: "apartment_id=1",
			userID:     "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("DeleteUserApartment", 1, 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid apartment id",
			queryParam:     "apartment_id=invalid",
			userID:         "1",
			mockSetup:      func(*repositories.MockUserApartmentRepository) {},
			expectedStatus: http.StatusBadRequest,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("POST", "/apartments/leave?"+tt.queryParam, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.LeaveApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateApartment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo)
		expectedStatus int
	}{
		{
			name: "successful update",
			requestBody: map[string]interface{}{
				"id":             1,
				"apartment_name": "Updated Name",
				"address":        "Updated Address",
				"units_count":    20,
				"manager_id":     1,
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("UpdateApartment", mock.Anything, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid request body",
			requestBody: map[string]interface{}{
				"invalid": "data",
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				//no mocks needed since validation fails before repository calls
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "not authorized to update",
			requestBody: map[string]interface{}{
				"id":             1,
				"apartment_name": "Updated Name",
				"address":        "Updated Address",
				"units_count":    20,
				"manager_id":     1,
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedStatus: http.StatusInternalServerError,
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
			handler := NewApartmentHandler(service)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/apartments", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.UpdateApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockAptRepo.AssertExpectations(t)
			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestDeleteApartment(t *testing.T) {
	tests := []struct {
		name           string
		queryParam     string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo)
		expectedStatus int
	}{
		{
			name:       "successful delete",
			queryParam: "id=1",
			userID:     "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				aptRepo.On("DeleteApartment", 1).Return(nil)
				userAptRepo.On("DeleteApartmentFromUserApartments", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid apartment id",
			queryParam:     "id=invalid",
			userID:         "1",
			mockSetup:      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "not authorized to delete",
			queryParam: "id=1",
			userID:     "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, aptRepo *repositories.MockApartmentRepo) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(false, nil)
			},
			expectedStatus: http.StatusInternalServerError,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("DELETE", "/apartments?"+tt.queryParam, nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.DeleteApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockAptRepo.AssertExpectations(t)
			mockUserAptRepo.AssertExpectations(t)
		})
	}
}

func TestGetAllApartmentsForResident(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedStatus int
	}{
		{
			name:   "successful get apartments",
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("GetAllApartmentsForAResident", 1).Return([]models.Apartment{
					{BaseModel: models.BaseModel{ID: 1}, ApartmentName: "Apt 1"},
					{BaseModel: models.BaseModel{ID: 2}, ApartmentName: "Apt 2"},
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid resident id",
			userID:         "invalid",
			mockSetup:      func(*repositories.MockUserApartmentRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "failed to get apartments",
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository) {
				userAptRepo.On("GetAllApartmentsForAResident", 1).Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
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
			handler := NewApartmentHandler(service)

			req := httptest.NewRequest("GET", "/users/"+tt.userID+"/apartments", nil)
			req.SetPathValue("user_id", tt.userID)
			w := httptest.NewRecorder()

			handler.GetAllApartmentsForResident(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}
