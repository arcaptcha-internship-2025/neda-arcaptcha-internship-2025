package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(ctx context.Context, req dto.CreateUserRequest, botAddress string) (*dto.SignUpResponse, error) {
	args := m.Called(ctx, req, botAddress)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.SignUpResponse), args.Error(1)
}

func (m *MockUserService) AuthenticateUser(ctx context.Context, req dto.LoginRequest) (*dto.LoginResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.LoginResponse), args.Error(1)
}

func (m *MockUserService) GetUserProfile(ctx context.Context, userID int) (*dto.ProfileResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileResponse), args.Error(1)
}

func (m *MockUserService) UpdateUserProfile(ctx context.Context, userID int, req dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	args := m.Called(ctx, userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.ProfileResponse), args.Error(1)
}

func (m *MockUserService) GetPublicUser(ctx context.Context, userID int) (*dto.PublicUserResponse, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dto.PublicUserResponse), args.Error(1)
}

func (m *MockUserService) GetAllPublicUsers(ctx context.Context) ([]dto.PublicUserResponse, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]dto.PublicUserResponse), args.Error(1)
}

func (m *MockUserService) DeleteUser(ctx context.Context, userID int) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func TestUserHandler_SignUp(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful signup",
			requestBody: dto.CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				Phone:        "1234567890",
				FullName:     "Test User",
				UserType:     models.Resident,
				TelegramUser: "testuser",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.CreateUserRequest"), "https://t.me/testbot").Return(&dto.SignUpResponse{
					User: dto.UserInfo{
						ID:           1,
						Username:     "testuser",
						Email:        "test@example.com",
						Phone:        "1234567890",
						FullName:     "Test User",
						UserType:     models.Resident,
						TelegramUser: "testuser",
					},
					TelegramSetupRequired: true,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Invalid request body",
		},
		{
			name: "missing required fields",
			requestBody: dto.CreateUserRequest{
				Username: "",
				Password: "password123",
				Email:    "test@example.com",
			},
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "username already exists",
			requestBody: dto.CreateUserRequest{
				Username: "existinguser",
				Password: "password123",
				Email:    "test@example.com",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.CreateUserRequest"), "https://t.me/testbot").Return(nil, fmt.Errorf("username already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectedError:  "username already exists",
		},
		{
			name: "internal server error",
			requestBody: dto.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
			},
			mockSetup: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.CreateUserRequest"), "https://t.me/testbot").Return(nil, fmt.Errorf("failed to create user"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body = *bytes.NewBufferString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/signup", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SignUp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err == nil && response["error"] != nil {
					if errorStr, ok := response["error"].(string); ok {
						assert.Contains(t, errorStr, tt.expectedError)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful login",
			requestBody: dto.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func(m *MockUserService) {
				m.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("dto.LoginRequest")).Return(&dto.LoginResponse{
					Token:    "Bearer token123",
					UserID:   "1",
					UserType: "resident",
					Username: "testuser",
					Email:    "test@example.com",
					FullName: "Test User",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid credentials",
			requestBody: dto.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func(m *MockUserService) {
				m.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("dto.LoginRequest")).Return(nil, fmt.Errorf("invalid username or password"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedError:  "invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body = *bytes.NewBufferString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPost, "/login", &body)
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedError != "" {
				var response map[string]interface{}
				err := json.Unmarshal(w.Body.Bytes(), &response)
				if err == nil && response["error"] != nil {
					if errorStr, ok := response["error"].(string); ok {
						assert.Contains(t, errorStr, tt.expectedError)
					}
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful get profile",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, 1).Return(&dto.ProfileResponse{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					FullName: "Test User",
					UserType: models.Resident,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, 999).Return(nil, fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    interface{}
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful update",
			userID: "1",
			requestBody: dto.UpdateProfileRequest{
				Username: "updateduser",
				Email:    "updated@example.com",
			},
			mockSetup: func(m *MockUserService) {
				m.On("UpdateUserProfile", mock.Anything, 1, mock.AnythingOfType("dto.UpdateProfileRequest")).Return(&dto.ProfileResponse{
					ID:       1,
					Username: "updateduser",
					Email:    "updated@example.com",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid request body",
			userID:         "1",
			requestBody:    "invalid json",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			var body bytes.Buffer
			if str, ok := tt.requestBody.(string); ok {
				body = *bytes.NewBufferString(str)
			} else {
				json.NewEncoder(&body).Encode(tt.requestBody)
			}

			req := httptest.NewRequest(http.MethodPut, "/profile", &body)
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.UpdateProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful get user",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("GetPublicUser", mock.Anything, 1).Return(&dto.PublicUserResponse{
					ID:       1,
					Username: "testuser",
					FullName: "Test User",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("GetPublicUser", mock.Anything, 999).Return(nil, fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			req := httptest.NewRequest(http.MethodGet, "/users/"+tt.userID, nil)
			req.SetPathValue("user_id", tt.userID)
			w := httptest.NewRecorder()

			handler.GetUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	tests := []struct {
		name           string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name: "successful get all users",
			mockSetup: func(m *MockUserService) {
				users := []dto.PublicUserResponse{
					{ID: 1, Username: "user1", FullName: "User 1"},
					{ID: 2, Username: "user2", FullName: "User 2"},
				}
				m.On("GetAllPublicUsers", mock.Anything).Return(users, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			mockSetup: func(m *MockUserService) {
				m.On("GetAllPublicUsers", mock.Anything).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()

			handler.GetAllUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			userID: "1",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userID:         "invalid",
			mockSetup:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, 999).Return(fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := &MockUserService{}
			tt.mockSetup(mockService)

			handler := NewUserHandler(mockService, "https://t.me/testbot")

			req := httptest.NewRequest(http.MethodDelete, "/users/"+tt.userID, nil)
			req.SetPathValue("user_id", tt.userID)
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_getCurrentUserID(t *testing.T) {
	tests := []struct {
		name        string
		userIDValue interface{}
		expectError bool
		expectedID  int
	}{
		{
			name:        "valid user ID",
			userIDValue: "123",
			expectError: false,
			expectedID:  123,
		},
		{
			name:        "missing user ID",
			userIDValue: nil,
			expectError: true,
		},
		{
			name:        "invalid format",
			userIDValue: 123,
			expectError: true,
		},
		{
			name:        "invalid number format",
			userIDValue: "invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := &UserHandler{}

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.userIDValue != nil {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userIDValue)
				req = req.WithContext(ctx)
			}

			userID, err := handler.getCurrentUserID(req)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}
