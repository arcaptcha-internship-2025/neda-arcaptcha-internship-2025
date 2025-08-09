package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/handlers"
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
			mockSetup: func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockApartmentRepo) {
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
			handler := handlers.NewApartmentHandler(service)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/apartments", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("GET", "/apartments?"+tt.queryParams, nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
		pathParam      string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedStatus int
	}{
		{
			name:      "successful get residents",
			pathParam: "1",
			userID:    "1",
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
			pathParam:      "invalid",
			userID:         "1",
			mockSetup:      func(*repositories.MockUserApartmentRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:      "not manager of apartment",
			pathParam: "1",
			userID:    "1",
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
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("GET", "/apartments/"+tt.pathParam+"/residents", nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
		name           string
		pathParams     map[string]string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockUserRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification)
		expectedStatus int
	}{
		{
			name: "successful invitation",
			pathParams: map[string]string{
				"apartment_id":      "1",
				"telegram_username": "testuser",
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, nil)
				inviteRepo.On("CreateInvitation", mock.Anything, 2, 1, 1).Return("invite123", nil)
				notif.On("SendInvitation", mock.Anything, mock.Anything, 1, "testuser").Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "user already resident",
			pathParams: map[string]string{
				"apartment_id":      "1",
				"telegram_username": "testuser",
			},
			userID: "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, userRepo *repositories.MockUserRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				userAptRepo.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				userRepo.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 2, 1).Return(true, nil)
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

			tt.mockSetup(mockUserAptRepo, mockUserRepo, mockInviteRepo, mockNotif)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("POST", "/apartments/"+tt.pathParams["apartment_id"]+"/invite/"+tt.pathParams["telegram_username"], nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
		pathParam      string
		userID         string
		mockSetup      func(*repositories.MockUserApartmentRepository, *repositories.MockInviteLinkRepository, *notification.MockNotification)
		expectedStatus int
	}{
		{
			name:      "successful join",
			pathParam: "validcode",
			userID:    "1",
			mockSetup: func(userAptRepo *repositories.MockUserApartmentRepository, inviteRepo *repositories.MockInviteLinkRepository, notif *notification.MockNotification) {
				inviteRepo.On("ValidateAndConsumeInvitation", mock.Anything, "validcode").Return(1, nil)
				userAptRepo.On("IsUserInApartment", mock.Anything, 1, 1).Return(false, nil)
				userAptRepo.On("CreateUserApartment", mock.Anything, mock.Anything).Return(nil)
				notif.On("SendNotification", mock.Anything, 1, mock.Anything).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:      "invalid invitation code",
			pathParam: "invalidcode",
			userID:    "1",
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
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("POST", "/apartments/join/"+tt.pathParam, nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("POST", "/apartments/leave?"+tt.queryParam, nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
			userID:         "1",
			mockSetup:      func(*repositories.MockUserApartmentRepository, *repositories.MockApartmentRepo) {},
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

			tt.mockSetup(mockUserAptRepo, mockAptRepo)

			service := services.NewApartmentService(
				mockAptRepo,
				mockUserRepo,
				mockUserAptRepo,
				mockInviteRepo,
				mockNotif,
			)
			handler := handlers.NewApartmentHandler(service)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/apartments", bytes.NewReader(body))
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("DELETE", "/apartments?"+tt.queryParam, nil)
			ctx := context.WithValue(req.Context(), "user_id", tt.userID)
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
		pathParam      string
		mockSetup      func(*repositories.MockUserApartmentRepository)
		expectedStatus int
	}{
		{
			name:      "successful get apartments",
			pathParam: "1",
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
			pathParam:      "invalid",
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
			handler := handlers.NewApartmentHandler(service)

			req := httptest.NewRequest("GET", "/users/"+tt.pathParam+"/apartments", nil)
			w := httptest.NewRecorder()

			handler.GetAllApartmentsForResident(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			mockUserAptRepo.AssertExpectations(t)
		})
	}
}
