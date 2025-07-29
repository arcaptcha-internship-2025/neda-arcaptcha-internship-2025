package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserHandler_SignUp(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    CreateUserRequest
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful signup",
			requestBody: CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				Phone:        "1234567890",
				FullName:     "Test User",
				UserType:     models.Resident,
				TelegramUser: "testuser_tg",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(nil, sql.ErrNoRows)
				m.On("GetUserByTelegramUser", "testuser_tg").Return(nil, sql.ErrNoRows)
				m.On("CreateUser", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("models.User")).Return(1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "missing required fields",
			requestBody: CreateUserRequest{
				Username: "testuser",
				// missing password and email
			},
			setupMocks:     func(m *repositories.MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid user type",
			requestBody: CreateUserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: "invalid",
			},
			setupMocks:     func(m *repositories.MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid telegram username",
			requestBody: CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				UserType:     models.Resident,
				TelegramUser: "abc", // too short
			},
			setupMocks:     func(m *repositories.MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "username already exists",
			requestBody: CreateUserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: models.Resident,
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{Username: "testuser"}
				m.On("GetUserByUsername", "testuser").Return(existingUser, nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name: "telegram username already exists",
			requestBody: CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				UserType:     models.Resident,
				TelegramUser: "existing_tg",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(nil, sql.ErrNoRows)
				existingTgUser := &models.User{TelegramUser: "existing_tg"}
				m.On("GetUserByTelegramUser", "existing_tg").Return(existingTgUser, nil)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SignUp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name           string
		requestBody    LoginRequest
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful login",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				user := &models.User{
					Username:       "testuser",
					Password:       string(hashedPassword),
					Email:          "test@example.com",
					FullName:       "Test User",
					UserType:       models.Resident,
					TelegramUser:   "testuser_tg",
					TelegramChatID: 123456,
				}
				user.BaseModel.ID = 1
				m.On("GetUserByUsername", "testuser").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "user not found",
			requestBody: LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("GetUserByUsername", "nonexistent").Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "wrong password",
			requestBody: LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				user := &models.User{
					Username: "testuser",
					Password: string(hashedPassword),
				}
				m.On("GetUserByUsername", "testuser").Return(user, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "missing fields",
			requestBody: LoginRequest{
				Username: "testuser",
				// missing password
			},
			setupMocks:     func(m *repositories.MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name:   "successful profile retrieval",
			userID: "1",
			setupMocks: func(m *repositories.MockUserRepository) {
				user := &models.User{
					Username:       "testuser",
					Email:          "test@example.com",
					Phone:          "1234567890",
					FullName:       "Test User",
					UserType:       models.Resident,
					TelegramUser:   "testuser_tg",
					TelegramChatID: 123456,
				}
				user.BaseModel.ID = 1
				m.On("GetUserByID", 1).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "999",
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("GetUserByID", 999).Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		requestBody    UpdateProfileRequest
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name:   "successful profile update",
			userID: "1",
			requestBody: UpdateProfileRequest{
				Username:     "updateduser",
				Email:        "updated@example.com",
				Phone:        "0987654321",
				FullName:     "Updated User",
				TelegramUser: "updated_tg",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					Username:       "testuser",
					Email:          "test@example.com",
					Phone:          "1234567890",
					FullName:       "Test User",
					UserType:       models.Resident,
					TelegramUser:   "testuser_tg",
					TelegramChatID: 123456,
				}
				existingUser.BaseModel.ID = 1
				m.On("GetUserByID", 1).Return(existingUser, nil)
				m.On("GetUserByTelegramUser", "updated_tg").Return(nil, sql.ErrNoRows)
				m.On("UpdateUser", mock.AnythingOfType("*context.timerCtx"), mock.AnythingOfType("models.User")).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "telegram username already taken",
			userID: "1",
			requestBody: UpdateProfileRequest{
				TelegramUser: "taken_tg",
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					TelegramUser: "old_tg",
				}
				existingUser.BaseModel.ID = 1
				otherUser := &models.User{
					TelegramUser: "taken_tg",
				}
				otherUser.BaseModel.ID = 2
				m.On("GetUserByID", 1).Return(existingUser, nil)
				m.On("GetUserByTelegramUser", "taken_tg").Return(otherUser, nil)
			},
			expectedStatus: http.StatusConflict,
		},
		{
			name:   "invalid telegram username format",
			userID: "1",
			requestBody: UpdateProfileRequest{
				TelegramUser: "abc", // too short
			},
			setupMocks: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					TelegramUser: "old_tg",
				}
				existingUser.BaseModel.ID = 1
				m.On("GetUserByID", 1).Return(existingUser, nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			jsonBody, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPut, "/profile", bytes.NewBuffer(jsonBody))
			req.Header.Set("Content-Type", "application/json")
			ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
			req = req.WithContext(ctx)
			w := httptest.NewRecorder()

			handler.UpdateProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful user retrieval",
			path: "/users/1",
			setupMocks: func(m *repositories.MockUserRepository) {
				user := &models.User{
					Username: "testuser",
					FullName: "Test User",
				}
				user.BaseModel.ID = 1
				m.On("GetUserByID", 1).Return(user, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			path:           "/users/invalid",
			setupMocks:     func(m *repositories.MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			path: "/users/999",
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("GetUserByID", 999).Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			w := httptest.NewRecorder()

			handler.GetUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetAllUsers(t *testing.T) {
	tests := []struct {
		name           string
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful users retrieval",
			setupMocks: func(m *repositories.MockUserRepository) {
				user1 := models.User{
					Username: "user1",
					FullName: "User One",
					UserType: models.Resident,
				}
				user1.BaseModel.ID = 1

				user2 := models.User{
					Username: "user2",
					FullName: "User Two",
					UserType: models.Manager,
				}
				user2.BaseModel.ID = 2

				users := []models.User{user1, user2}
				m.On("GetAllUsers", mock.Anything).Return(users, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "database error",
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("GetAllUsers", mock.Anything).Return(nil, fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest(http.MethodGet, "/users", nil)
			w := httptest.NewRecorder()

			handler.GetAllUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		setupMocks     func(*repositories.MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful user deletion",
			path: "/users/1",
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("DeleteUser", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid user ID",
			path:           "/users/invalid",
			setupMocks:     func(m *repositories.MockUserRepository) {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "user not found",
			path: "/users/999",
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("DeleteUser", 999).Return(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name: "database error",
			path: "/users/1",
			setupMocks: func(m *repositories.MockUserRepository) {
				m.On("DeleteUser", 1).Return(fmt.Errorf("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(repositories.MockUserRepository)
			tt.setupMocks(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest(http.MethodDelete, tt.path, nil)
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_getCurrentUserID(t *testing.T) {
	handler := &UserHandler{}

	tests := []struct {
		name        string
		contextFunc func() context.Context
		expectedID  int
		expectError bool
	}{
		{
			name: "valid user ID",
			contextFunc: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDKey, "123")
			},
			expectedID:  123,
			expectError: false,
		},
		{
			name: "missing user ID in context",
			contextFunc: func() context.Context {
				return context.Background()
			},
			expectedID:  0,
			expectError: true,
		},
		{
			name: "invalid user ID format",
			contextFunc: func() context.Context {
				return context.WithValue(context.Background(), middleware.UserIDKey, 123) // int instead of string
			},
			expectedID:  0,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req = req.WithContext(tt.contextFunc())

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

func TestIsValidTelegramUsername(t *testing.T) {
	tests := []struct {
		username string
		expected bool
	}{
		{"valid_username", true},
		{"user123", true},
		{"test_user_123", true},
		{"abc", false}, // too short
		{"a", false},   // too short
		{"this_is_a_very_long_username_that_exceeds_limit", false}, // too long
		{"user@name", false}, // invalid character
		{"user-name", false}, // invalid character
		{"User_Name", false}, // uppercase not allowed
		{"user.name", false}, // invalid character
		{"", false},          // empty
	}

	for _, tt := range tests {
		t.Run(tt.username, func(t *testing.T) {
			result := isValidTelegramUsername(tt.username)
			assert.Equal(t, tt.expected, result)
		})
	}
}
