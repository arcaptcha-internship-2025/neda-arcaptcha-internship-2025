package handlers

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type contextKey string

const (
	userIDContextKey contextKey = "user_id"
)

type mockApartmentRepo struct {
	mock.Mock
}

func (m *mockApartmentRepo) CreateApartment(ctx context.Context, apartment models.Apartment) (int, error) {
	args := m.Called(ctx, apartment)
	return args.Int(0), args.Error(1)
}

func (m *mockApartmentRepo) GetApartmentByID(id int) (*models.Apartment, error) {
	args := m.Called(id)
	return args.Get(0).(*models.Apartment), args.Error(1)
}

func (m *mockApartmentRepo) UpdateApartment(ctx context.Context, apartment models.Apartment) error {
	args := m.Called(ctx, apartment)
	return args.Error(0)
}

func (m *mockApartmentRepo) DeleteApartment(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) CreateUser(ctx context.Context, user models.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

func (m *mockUserRepo) GetUserByID(id int) (*models.User, error) {
	args := m.Called(id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *mockUserRepo) GetAllUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByPhone(phone string) (*models.User, error) {
	args := m.Called(phone)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByTelegramUser(telegramUser string) (*models.User, error) {
	args := m.Called(telegramUser)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserRepo) UpdateTelegramChatID(ctx context.Context, telegramUsername string, chatID int64) error {
	args := m.Called(ctx, telegramUsername, chatID)
	return args.Error(0)
}

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

func TestCreateApartment(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		mockSetup      func(*mockApartmentRepo, *mockUserApartmentRepo)
		expectedStatus int
		userID         int
	}{
		{
			name:        "successful creation",
			requestBody: `{"apartment_name": "Sunny Apartments", "address": "123 Main St", "units_count": 10}`,
			mockSetup: func(ar *mockApartmentRepo, uar *mockUserApartmentRepo) {
				ar.On("CreateApartment", mock.Anything, mock.AnythingOfType("models.Apartment")).
					Return(1, nil)
				uar.On("CreateUserApartment", mock.Anything, mock.AnythingOfType("models.User_apartment")).
					Return(nil)
			},
			expectedStatus: http.StatusOK,
			userID:         1,
		},
		{
			name:           "invalid request body",
			requestBody:    `invalid json`,
			mockSetup:      func(ar *mockApartmentRepo, uar *mockUserApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
			userID:         1,
		},
		{
			name:        "creation error",
			requestBody: `{"apartment_name": "Sunny Apartments", "address": "123 Main St", "units_count": 10}`,
			mockSetup: func(ar *mockApartmentRepo, uar *mockUserApartmentRepo) {
				ar.On("CreateApartment", mock.Anything, mock.AnythingOfType("models.Apartment")).
					Return(0, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			userID:         1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := new(mockApartmentRepo)
			uar := new(mockUserApartmentRepo)
			tt.mockSetup(ar, uar)

			handler := &ApartmentHandler{
				apartmentRepo:     ar,
				userApartmentRepo: uar,
			}

			req := httptest.NewRequest("POST", "/apartments", strings.NewReader(tt.requestBody))
			req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, tt.userID))
			w := httptest.NewRecorder()

			handler.CreateApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			ar.AssertExpectations(t)
			uar.AssertExpectations(t)
		})
	}
}

func TestGetApartmentByID(t *testing.T) {
	tests := []struct {
		name           string
		queryParams    string
		mockSetup      func(*mockApartmentRepo)
		expectedStatus int
	}{
		{
			name:        "successful retrieval",
			queryParams: "id=1",
			mockSetup: func(ar *mockApartmentRepo) {
				ar.On("GetApartmentByID", 1).
					Return(&models.Apartment{BaseModel: models.BaseModel{ID: 1}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id parameter",
			queryParams:    "",
			mockSetup:      func(ar *mockApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid id parameter",
			queryParams:    "id=invalid",
			mockSetup:      func(ar *mockApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "not found",
			queryParams: "id=999",
			mockSetup: func(ar *mockApartmentRepo) {
				ar.On("GetApartmentByID", 999).
					Return(&models.Apartment{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := new(mockApartmentRepo)
			tt.mockSetup(ar)

			handler := &ApartmentHandler{
				apartmentRepo: ar,
			}

			req := httptest.NewRequest("GET", "/apartments?"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			handler.GetApartmentByID(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			ar.AssertExpectations(t)
		})
	}
}

func TestGetResidentsInApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		mockSetup      func(*mockUserApartmentRepo)
		expectedStatus int
	}{
		{
			name:        "successful retrieval",
			apartmentID: "1",
			mockSetup: func(uar *mockUserApartmentRepo) {
				uar.On("GetResidentsInApartment", 1).
					Return([]models.User{{BaseModel: models.BaseModel{ID: 1}}}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid apartment id",
			apartmentID:    "invalid",
			mockSetup:      func(uar *mockUserApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "database error",
			apartmentID: "1",
			mockSetup: func(uar *mockUserApartmentRepo) {
				uar.On("GetResidentsInApartment", 1).
					Return([]models.User{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uar := new(mockUserApartmentRepo)
			tt.mockSetup(uar)

			handler := &ApartmentHandler{
				userApartmentRepo: uar,
			}

			req := httptest.NewRequest("GET", "/apartments/"+tt.apartmentID+"/residents", nil)
			req = mux.SetURLVars(req, map[string]string{"apartment-id": tt.apartmentID})
			w := httptest.NewRecorder()

			handler.GetResidentsInApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			uar.AssertExpectations(t)
		})
	}
}

func TestInviteUserToApartment(t *testing.T) {
	tests := []struct {
		name           string
		telegramUser   string
		managerID      int
		apartmentID    string
		mockSetup      func(*mockApartmentRepo, *mockUserRepo, *mockUserApartmentRepo, *mockInviteLinkRepo, *mockNotificationService)
		expectedStatus int
	}{
		{
			name:         "successful invitation",
			telegramUser: "testuser",
			managerID:    1,
			apartmentID:  "1",
			mockSetup: func(ar *mockApartmentRepo, ur *mockUserRepo, uar *mockUserApartmentRepo, ilr *mockInviteLinkRepo, ns *mockNotificationService) {
				uar.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				ar.On("GetApartmentByID", 1).Return(&models.Apartment{
					BaseModel:     models.BaseModel{ID: 1},
					ApartmentName: "Test Apartments",
				}, nil)
				ur.On("GetUserByID", 1).Return(&models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "manager",
				}, nil)
				ur.On("GetUserByTelegramUser", "testuser").Return(&models.User{
					BaseModel:      models.BaseModel{ID: 2},
					Username:       "testuser",
					TelegramChatID: 12345,
				}, nil)
				uar.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, nil)
				ilr.On("CreateInvitation", mock.Anything, mock.MatchedBy(func(inv models.InvitationLink) bool {
					return inv.ApartmentID == 1 && inv.ReceiverUsername == "testuser"
				})).Return(nil)
				ns.On("SendInvitation", mock.Anything, mock.MatchedBy(func(inv models.InvitationLink) bool {
					return inv.ApartmentID == 1 && inv.ReceiverUsername == "testuser"
				})).Return(nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:         "user not found",
			telegramUser: "nonexistent",
			managerID:    1,
			apartmentID:  "1",
			mockSetup: func(ar *mockApartmentRepo, ur *mockUserRepo, uar *mockUserApartmentRepo, ilr *mockInviteLinkRepo, ns *mockNotificationService) {
				uar.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				ar.On("GetApartmentByID", 1).Return(&models.Apartment{BaseModel: models.BaseModel{ID: 1}}, nil)
				ur.On("GetUserByID", 1).Return(&models.User{BaseModel: models.BaseModel{ID: 1}}, nil)
				ur.On("GetUserByTelegramUser", "nonexistent").Return(&models.User{}, errors.New("not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:         "user already in apartment",
			telegramUser: "testuser",
			managerID:    1,
			apartmentID:  "1",
			mockSetup: func(ar *mockApartmentRepo, ur *mockUserRepo, uar *mockUserApartmentRepo, ilr *mockInviteLinkRepo, ns *mockNotificationService) {
				uar.On("IsUserManagerOfApartment", mock.Anything, 1, 1).Return(true, nil)
				ar.On("GetApartmentByID", 1).Return(&models.Apartment{BaseModel: models.BaseModel{ID: 1}}, nil)
				ur.On("GetUserByID", 1).Return(&models.User{BaseModel: models.BaseModel{ID: 1}}, nil)
				ur.On("GetUserByTelegramUser", "testuser").Return(&models.User{BaseModel: models.BaseModel{ID: 2}}, nil)
				uar.On("IsUserInApartment", mock.Anything, 2, 1).Return(true, nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:         "not manager",
			telegramUser: "testuser",
			managerID:    2,
			apartmentID:  "1",
			mockSetup: func(ar *mockApartmentRepo, ur *mockUserRepo, uar *mockUserApartmentRepo, ilr *mockInviteLinkRepo, ns *mockNotificationService) {
				uar.On("IsUserManagerOfApartment", mock.Anything, 2, 1).Return(false, errors.New("not manager"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ar := new(mockApartmentRepo)
			ur := new(mockUserRepo)
			uar := new(mockUserApartmentRepo)
			ilr := new(mockInviteLinkRepo)
			ns := new(mockNotificationService)
			tt.mockSetup(ar, ur, uar, ilr, ns)

			handler := &ApartmentHandler{
				apartmentRepo:       ar,
				userRepo:            ur,
				userApartmentRepo:   uar,
				inviteLinkRepo:      ilr,
				notificationService: ns,
				appBaseURL:          "http://localhost",
			}

			req := httptest.NewRequest("POST", "/apartments/"+tt.apartmentID+"/invite/"+tt.telegramUser, nil)
			req = mux.SetURLVars(req, map[string]string{
				"apartment-id":      tt.apartmentID,
				"telegram-username": tt.telegramUser,
			})
			req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, tt.managerID))
			w := httptest.NewRecorder()

			handler.InviteUserToApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			ar.AssertExpectations(t)
			ur.AssertExpectations(t)
			uar.AssertExpectations(t)

			if tt.name == "successful invitation" {
				ilr.AssertExpectations(t)
				ns.AssertExpectations(t)
			}
		})
	}
}

func TestJoinApartment(t *testing.T) {
	tests := []struct {
		name           string
		token          string
		userID         int
		mockSetup      func(*mockInviteLinkRepo, *mockUserApartmentRepo)
		expectedStatus int
	}{
		{
			name:   "successful join",
			token:  "valid-token",
			userID: 2,
			mockSetup: func(ilr *mockInviteLinkRepo, uar *mockUserApartmentRepo) {
				inv := &models.InvitationLink{
					ApartmentID: 1,
					Status:      models.InvitationStatusPending,
					ExpiresAt:   time.Now().Add(24 * time.Hour),
				}
				ilr.On("GetInvitationByToken", mock.Anything, "valid-token").Return(inv, nil)
				uar.On("IsUserInApartment", mock.Anything, 2, 1).Return(false, nil)
				uar.On("CreateUserApartment", mock.Anything, mock.AnythingOfType("models.User_apartment")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "invalid token",
			token:  "invalid-token",
			userID: 2,
			mockSetup: func(ilr *mockInviteLinkRepo, uar *mockUserApartmentRepo) {
				ilr.On("GetInvitationByToken", mock.Anything, "invalid-token").Return(&models.InvitationLink{}, errors.New("invalid token"))
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "already a resident",
			token:  "valid-token",
			userID: 2,
			mockSetup: func(ilr *mockInviteLinkRepo, uar *mockUserApartmentRepo) {
				inv := &models.InvitationLink{
					ApartmentID: 1,
					Status:      models.InvitationStatusPending,
					ExpiresAt:   time.Now().Add(24 * time.Hour),
				}
				ilr.On("GetInvitationByToken", mock.Anything, "valid-token").Return(inv, nil)
				uar.On("IsUserInApartment", mock.Anything, 2, 1).Return(true, nil)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ilr := new(mockInviteLinkRepo)
			uar := new(mockUserApartmentRepo)
			tt.mockSetup(ilr, uar)

			handler := &ApartmentHandler{
				inviteLinkRepo:    ilr,
				userApartmentRepo: uar,
			}

			//creating req with query parameter
			req := httptest.NewRequest("POST", "/apartment/join?token="+tt.token, nil)
			req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, tt.userID))
			w := httptest.NewRecorder()

			handler.JoinApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			ilr.AssertExpectations(t)
			uar.AssertExpectations(t)
		})
	}
}

func TestLeaveApartment(t *testing.T) {
	tests := []struct {
		name           string
		apartmentID    string
		userID         int
		mockSetup      func(*mockUserApartmentRepo)
		expectedStatus int
	}{
		{
			name:        "successful leave",
			apartmentID: "1",
			userID:      1,
			mockSetup: func(uar *mockUserApartmentRepo) {
				uar.On("DeleteUserApartment", 1, 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid apartment id",
			apartmentID:    "invalid",
			userID:         1,
			mockSetup:      func(uar *mockUserApartmentRepo) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "database error",
			apartmentID: "1",
			userID:      1,
			mockSetup: func(uar *mockUserApartmentRepo) {
				uar.On("DeleteUserApartment", 1, 1).Return(errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uar := new(mockUserApartmentRepo)
			tt.mockSetup(uar)

			handler := &ApartmentHandler{
				userApartmentRepo: uar,
			}

			req := httptest.NewRequest("POST", "/apartments/leave?apartment_id="+tt.apartmentID, nil)
			req = req.WithContext(context.WithValue(req.Context(), userIDContextKey, tt.userID))
			w := httptest.NewRecorder()

			handler.LeaveApartment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			uar.AssertExpectations(t)
		})
	}
}
