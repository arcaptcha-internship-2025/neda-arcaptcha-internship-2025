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
		setupMock      func(*MockUserService)
		expectedStatus int
		expectError    bool
		errorMsg       string
		expectSuccess  bool
		successMsg     string
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
				TelegramUser: "testuser_tg",
			},
			setupMock: func(m *MockUserService) {
				response := &dto.SignUpResponse{
					User: dto.UserInfo{
						ID:           1,
						Username:     "testuser",
						Email:        "test@example.com",
						Phone:        "1234567890",
						FullName:     "Test User",
						UserType:     models.Resident,
						TelegramUser: "testuser_tg",
					},
					TelegramSetupRequired:     true,
					TelegramSetupInstructions: "Please start a chat with our bot in Telegram to complete setup : @testbot",
				}
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.CreateUserRequest"), "@testbot").Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectSuccess:  true,
			successMsg:     "user created successfully",
		},
		{
			name:           "invalid JSON body",
			requestBody:    `{"invalid": "json",`, // malformed JSON
			setupMock:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMsg:       "Invalid request body",
		},
		{
			name: "missing required fields",
			requestBody: dto.CreateUserRequest{
				Username: "testuser",
				// missing password and email
			},
			setupMock:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name: "username already exists",
			requestBody: dto.CreateUserRequest{
				Username: "existinguser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: models.Resident,
			},
			setupMock: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.CreateUserRequest"), "@testbot").Return(nil, fmt.Errorf("username already exists"))
			},
			expectedStatus: http.StatusConflict,
			expectError:    true,
			errorMsg:       "username already exists",
		},
		{
			name: "invalid user type",
			requestBody: dto.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: "invalid_type",
			},
			setupMock: func(m *MockUserService) {
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("dto.CreateUserRequest"), "@testbot").Return(nil, fmt.Errorf("invalid user type"))
			},
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			errorMsg:       "invalid user type",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			var reqBody []byte
			switch body := tt.requestBody.(type) {
			case string:
				reqBody = []byte(body)
			default:
				var err error
				reqBody, err = json.Marshal(body)
				assert.NoError(t, err)
			}

			req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(reqBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SignUp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.name == "invalid JSON body" {
				return
			}

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			if tt.expectError {
				if tt.errorMsg != "" {
					assert.Equal(t, tt.errorMsg, response["error"])
				}
				if success, exists := response["success"]; exists {
					assert.False(t, success.(bool))
				}
			}

			if tt.expectSuccess {
				assert.Equal(t, tt.successMsg, response["message"])
				assert.True(t, response["success"].(bool))
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "successful login",
			requestBody: dto.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMock: func(m *MockUserService) {
				response := &dto.LoginResponse{
					Token:    "jwt-token",
					UserID:   "1",
					UserType: string(models.Resident),
					Username: "testuser",
					Email:    "test@example.com",
					FullName: "Test User",
					Telegram: dto.TelegramInfo{
						Username:  "testuser_tg",
						Connected: true,
					},
				}
				m.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("dto.LoginRequest")).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "login successful",
				"success": true,
			},
		},
		{
			name: "invalid credentials",
			requestBody: dto.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMock: func(m *MockUserService) {
				m.On("AuthenticateUser", mock.Anything, mock.AnythingOfType("dto.LoginRequest")).Return(nil, fmt.Errorf("invalid username or password"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error":   "invalid username or password",
				"success": false,
			},
		},
		{
			name: "missing required fields",
			requestBody: dto.LoginRequest{
				Username: "testuser",
				//missing password
			},
			setupMock:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				if key == "message" || key == "error" || key == "success" {
					assert.Equal(t, expectedValue, response[key])
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
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful get profile",
			userID: "1",
			setupMock: func(m *MockUserService) {
				response := &dto.ProfileResponse{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Phone:    "1234567890",
					FullName: "Test User",
					UserType: models.Resident,
					Telegram: dto.TelegramInfo{
						Username:  "testuser_tg",
						Connected: true,
					},
				}
				m.On("GetUserProfile", mock.Anything, 1).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "profile retrieved successfully",
				"success": true,
			},
		},
		{
			name:   "user not found",
			userID: "999",
			setupMock: func(m *MockUserService) {
				m.On("GetUserProfile", mock.Anything, 999).Return(nil, fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":   "user not found",
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			req := httptest.NewRequest("GET", "/profile", nil)
			//adding user ID to context
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				if key == "message" || key == "error" || key == "success" {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    dto.UpdateProfileRequest
		setupMock      func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:   "successful update profile",
			userID: "1",
			requestBody: dto.UpdateProfileRequest{
				FullName: "Updated Name",
				Phone:    "9876543210",
			},
			setupMock: func(m *MockUserService) {
				response := &dto.ProfileResponse{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Phone:    "9876543210",
					FullName: "Updated Name",
					UserType: models.Resident,
					Telegram: dto.TelegramInfo{
						Username:  "testuser_tg",
						Connected: true,
					},
				}
				m.On("UpdateUserProfile", mock.Anything, 1, mock.AnythingOfType("dto.UpdateProfileRequest")).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"message": "profile updated successfully",
				"success": true,
			},
		},
		{
			name:   "user not found",
			userID: "999",
			requestBody: dto.UpdateProfileRequest{
				FullName: "Updated Name",
			},
			setupMock: func(m *MockUserService) {
				m.On("UpdateUserProfile", mock.Anything, 999, mock.AnythingOfType("dto.UpdateProfileRequest")).Return(nil, fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
			expectedBody: map[string]interface{}{
				"error":   "user not found",
				"success": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			body, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			//adding user id to context
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.UpdateProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedBody {
				if key == "message" || key == "error" || key == "success" {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		userIDParam    string
		setupMock      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful get user",
			userIDParam: "1",
			setupMock: func(m *MockUserService) {
				response := &dto.PublicUserResponse{
					ID:       1,
					Username: "testuser",
					FullName: "Test User",
				}
				m.On("GetPublicUser", mock.Anything, 1).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userIDParam:    "invalid",
			setupMock:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "user not found",
			userIDParam: "999",
			setupMock: func(m *MockUserService) {
				m.On("GetPublicUser", mock.Anything, 999).Return(nil, fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			req := httptest.NewRequest("GET", "/users/"+tt.userIDParam, nil)
			req.SetPathValue("user_id", tt.userIDParam)
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
		setupMock      func(*MockUserService)
		expectedStatus int
	}{
		{
			name: "successful get all users",
			setupMock: func(m *MockUserService) {
				response := []dto.PublicUserResponse{
					{
						ID:       1,
						Username: "user1",
						FullName: "User One",
					},
					{
						ID:       2,
						Username: "user2",
						FullName: "User Two",
					},
				}
				m.On("GetAllPublicUsers", mock.Anything).Return(response, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "service error",
			setupMock: func(m *MockUserService) {
				m.On("GetAllPublicUsers", mock.Anything).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			req := httptest.NewRequest("GET", "/users", nil)
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
		userIDParam    string
		setupMock      func(*MockUserService)
		expectedStatus int
	}{
		{
			name:        "successful delete user",
			userIDParam: "1",
			setupMock: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			userIDParam:    "invalid",
			setupMock:      func(m *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "user not found",
			userIDParam: "999",
			setupMock: func(m *MockUserService) {
				m.On("DeleteUser", mock.Anything, 999).Return(fmt.Errorf("user not found"))
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setupMock(mockService)

			handler := NewUserHandler(mockService, "@testbot")

			req := httptest.NewRequest("DELETE", "/users/"+tt.userIDParam, nil)
			req.SetPathValue("user_id", tt.userIDParam)
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
		contextKey  interface{}
		contextVal  interface{}
		expectedID  int
		expectError bool
	}{
		{
			name:        "valid user ID",
			contextKey:  middleware.UserIDKey,
			contextVal:  "123",
			expectedID:  123,
			expectError: false,
		},
		{
			name:        "missing user ID",
			contextKey:  middleware.UserIDKey,
			contextVal:  nil,
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "invalid user ID format",
			contextKey:  middleware.UserIDKey,
			contextVal:  123, // not a string
			expectedID:  0,
			expectError: true,
		},
		{
			name:        "non-numeric user ID",
			contextKey:  middleware.UserIDKey,
			contextVal:  "abc",
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewUserHandler(nil, "@testbot")

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.contextVal != nil {
				ctx := context.WithValue(req.Context(), tt.contextKey, tt.contextVal)
				req = req.WithContext(ctx)
			}

			userID, err := handler.getCurrentUserID(req)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, 0, userID)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, userID)
			}
		})
	}
}
