package services

import (
	"context"
	"database/sql"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/dto"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/repositories"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_CreateUser(t *testing.T) {
	tests := []struct {
		name        string
		request     dto.CreateUserRequest
		botAddress  string
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful user creation with telegram",
			request: dto.CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				Phone:        "1234567890",
				FullName:     "Test User",
				UserType:     models.Resident,
				TelegramUser: "test_user",
			},
			botAddress: "https://t.me/testbot",
			mockSetup: func(m *repositories.MockUserRepository) {
				// Check username doesn't exist
				m.On("GetUserByUsername", "testuser").Return(nil, sql.ErrNoRows)
				// Check telegram username doesn't exist
				m.On("GetUserByTelegramUser", "test_user").Return(nil, sql.ErrNoRows)
				// Create user
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("models.User")).Return(1, nil)
			},
			expectError: false,
		},
		{
			name: "successful user creation without telegram",
			request: dto.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: models.Manager,
			},
			botAddress: "https://t.me/testbot",
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(nil, sql.ErrNoRows)
				m.On("CreateUser", mock.Anything, mock.AnythingOfType("models.User")).Return(1, nil)
			},
			expectError: false,
		},
		{
			name: "invalid user type",
			request: dto.CreateUserRequest{
				Username: "testuser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: "invalid",
			},
			mockSetup:   func(m *repositories.MockUserRepository) {},
			expectError: true,
			errorMsg:    "invalid user type",
		},
		{
			name: "invalid telegram username format",
			request: dto.CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				UserType:     models.Resident,
				TelegramUser: "ab", // too short
			},
			mockSetup:   func(m *repositories.MockUserRepository) {},
			expectError: true,
			errorMsg:    "invalid Telegram username format",
		},
		{
			name: "username already exists",
			request: dto.CreateUserRequest{
				Username: "existinguser",
				Password: "password123",
				Email:    "test@example.com",
				UserType: models.Resident,
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "existinguser",
				}
				m.On("GetUserByUsername", "existinguser").Return(existingUser, nil)
			},
			expectError: true,
			errorMsg:    "username already exists",
		},
		{
			name: "telegram username already exists",
			request: dto.CreateUserRequest{
				Username:     "testuser",
				Password:     "password123",
				Email:        "test@example.com",
				UserType:     models.Resident,
				TelegramUser: "existing_telegram",
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(nil, sql.ErrNoRows)
				existingTelegramUser := &models.User{
					BaseModel:    models.BaseModel{ID: 2},
					TelegramUser: "existing_telegram",
				}
				m.On("GetUserByTelegramUser", "existing_telegram").Return(existingTelegramUser, nil)
			},
			expectError: true,
			errorMsg:    "telegram username already in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil) // Assuming userApartmentRepo is not needed for this test

			response, err := service.CreateUser(context.Background(), tt.request, tt.botAddress)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.request.Username, response.User.Username)
				assert.Equal(t, tt.request.Email, response.User.Email)
				assert.Equal(t, tt.request.UserType, response.User.UserType)

				if tt.request.TelegramUser != "" {
					assert.True(t, response.TelegramSetupRequired)
					assert.Contains(t, response.TelegramSetupInstructions, tt.botAddress)
				} else {
					assert.False(t, response.TelegramSetupRequired)
				}
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_AuthenticateUser(t *testing.T) {
	// Create a hashed password for testing
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	tests := []struct {
		name        string
		request     dto.LoginRequest
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name: "successful authentication",
			request: dto.LoginRequest{
				Username: "testuser",
				Password: "password123",
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				user := &models.User{
					BaseModel:      models.BaseModel{ID: 1},
					Username:       "testuser",
					Password:       string(hashedPassword),
					Email:          "test@example.com",
					FullName:       "Test User",
					UserType:       models.Resident,
					TelegramUser:   "test_user",
					TelegramChatID: 12345,
				}
				m.On("GetUserByUsername", "testuser").Return(user, nil)
			},
			expectError: false,
		},
		{
			name: "user not found",
			request: dto.LoginRequest{
				Username: "nonexistent",
				Password: "password123",
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetUserByUsername", "nonexistent").Return(nil, sql.ErrNoRows)
			},
			expectError: true,
			errorMsg:    "invalid username or password",
		},
		{
			name: "invalid password",
			request: dto.LoginRequest{
				Username: "testuser",
				Password: "wrongpassword",
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				user := &models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "testuser",
					Password:  string(hashedPassword),
					UserType:  models.Resident,
				}
				m.On("GetUserByUsername", "testuser").Return(user, nil)
			},
			expectError: true,
			errorMsg:    "invalid username or password",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil)

			response, err := service.AuthenticateUser(context.Background(), tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.NotEmpty(t, response.Token)
				assert.NotEmpty(t, response.UserID)
				assert.Equal(t, tt.request.Username, response.Username)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetUserProfile(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
	}{
		{
			name:   "successful get profile",
			userID: 1,
			mockSetup: func(m *repositories.MockUserRepository) {
				user := &models.User{
					BaseModel:      models.BaseModel{ID: 1},
					Username:       "testuser",
					Email:          "test@example.com",
					Phone:          "1234567890",
					FullName:       "Test User",
					UserType:       models.Resident,
					TelegramUser:   "test_user",
					TelegramChatID: 12345,
				}
				m.On("GetUserByID", 1).Return(user, nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetUserByID", 999).Return(nil, sql.ErrNoRows)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil)

			response, err := service.GetUserProfile(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.userID, response.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_UpdateUserProfile(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		request     dto.UpdateProfileRequest
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful update",
			userID: 1,
			request: dto.UpdateProfileRequest{
				Username: "updateduser",
				Email:    "updated@example.com",
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "testuser",
					Email:     "test@example.com",
					UserType:  models.Resident,
				}
				m.On("GetUserByID", 1).Return(existingUser, nil)
				m.On("UpdateUser", mock.Anything, mock.AnythingOfType("models.User")).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetUserByID", 999).Return(nil, sql.ErrNoRows)
			},
			expectError: true,
			errorMsg:    "user not found",
		},
		{
			name:   "invalid telegram username",
			userID: 1,
			request: dto.UpdateProfileRequest{
				TelegramUser: "ab", // too short
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					BaseModel:    models.BaseModel{ID: 1},
					Username:     "testuser",
					TelegramUser: "old_telegram",
				}
				m.On("GetUserByID", 1).Return(existingUser, nil)
			},
			expectError: true,
			errorMsg:    "invalid Telegram username format",
		},
		{
			name:   "telegram username already in use",
			userID: 1,
			request: dto.UpdateProfileRequest{
				TelegramUser: "existing_telegram",
			},
			mockSetup: func(m *repositories.MockUserRepository) {
				existingUser := &models.User{
					BaseModel:    models.BaseModel{ID: 1},
					Username:     "testuser",
					TelegramUser: "old_telegram",
				}
				m.On("GetUserByID", 1).Return(existingUser, nil)

				conflictingUser := &models.User{
					BaseModel:    models.BaseModel{ID: 2},
					TelegramUser: "existing_telegram",
				}
				m.On("GetUserByTelegramUser", "existing_telegram").Return(conflictingUser, nil)
			},
			expectError: true,
			errorMsg:    "telegram username already in use",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil)

			response, err := service.UpdateUserProfile(context.Background(), tt.userID, tt.request)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.userID, response.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetPublicUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
	}{
		{
			name:   "successful get public user",
			userID: 1,
			mockSetup: func(m *repositories.MockUserRepository) {
				user := &models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "testuser",
					FullName:  "Test User",
					UserType:  models.Resident,
				}
				m.On("GetUserByID", 1).Return(user, nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetUserByID", 999).Return(nil, sql.ErrNoRows)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil)

			response, err := service.GetPublicUser(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Equal(t, tt.userID, response.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_GetAllPublicUsers(t *testing.T) {
	tests := []struct {
		name        string
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
	}{
		{
			name: "successful get all users",
			mockSetup: func(m *repositories.MockUserRepository) {
				users := []models.User{
					{BaseModel: models.BaseModel{ID: 1}, Username: "user1", FullName: "User 1", UserType: models.Resident},
					{BaseModel: models.BaseModel{ID: 2}, Username: "user2", FullName: "User 2", UserType: models.Manager},
				}
				m.On("GetAllUsers", mock.Anything).Return(users, nil)
			},
			expectError: false,
		},
		{
			name: "repository error",
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("GetAllUsers", mock.Anything).Return(nil, assert.AnError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil)

			response, err := service.GetAllPublicUsers(context.Background())

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, response)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, response)
				assert.Len(t, response, 2)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_DeleteUser(t *testing.T) {
	tests := []struct {
		name        string
		userID      int
		mockSetup   func(*repositories.MockUserRepository)
		expectError bool
		errorMsg    string
	}{
		{
			name:   "successful delete",
			userID: 1,
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("DeleteUser", 1).Return(nil)
			},
			expectError: false,
		},
		{
			name:   "user not found",
			userID: 999,
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("DeleteUser", 999).Return(sql.ErrNoRows)
			},
			expectError: true,
			errorMsg:    "user not found",
		},
		{
			name:   "repository error",
			userID: 1,
			mockSetup: func(m *repositories.MockUserRepository) {
				m.On("DeleteUser", 1).Return(assert.AnError)
			},
			expectError: true,
			errorMsg:    "failed to delete user",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &repositories.MockUserRepository{}
			tt.mockSetup(mockRepo)

			service := NewUserService(mockRepo, nil)

			err := service.DeleteUser(context.Background(), tt.userID)

			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func TestIsValidTelegramUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		expected bool
	}{
		{"valid username", "valid_user123", true},
		{"minimum length", "user1", true},
		{"maximum length", "user_with_exactly_32_characters", true},
		{"too short", "usr", false},
		{"too long", "this_username_is_way_too_long_for_telegram_username_validation", false},
		{"invalid characters", "user@name", false},
		{"uppercase letters", "UserName", false},
		{"starts with underscore", "_username", true},
		{"ends with underscore", "username_", true},
		{"multiple underscores", "user__name", true},
		{"only numbers", "123456", true},
		{"only letters", "username", true},
		{"mixed valid chars", "user123_test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidTelegramUsername(tt.username)
			assert.Equal(t, tt.expected, result)
		})
	}
}
