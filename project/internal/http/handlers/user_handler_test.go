package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/middleware"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) CreateUser(ctx context.Context, user models.User) (int, error) {
	args := m.Called(ctx, user)
	return args.Int(0), args.Error(1)
}

func (m *MockUserRepository) GetUserByID(id int) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) UpdateUser(ctx context.Context, user models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) DeleteUser(id int) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) GetAllUsers(ctx context.Context) ([]models.User, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserRepository) GetUserByPhone(phone string) (*models.User, error) {
	args := m.Called(phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func TestUserHandler_SignUp(t *testing.T) {
	tests := []struct {
		name           string
		input          models.User
		mockSetup      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful signup",
			input: models.User{
				BaseModel: models.BaseModel{ID: 1},
				Username:  "testuser",
				Password:  "password123",
				Email:     "test@example.com",
				UserType:  models.Resident,
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(nil, sql.ErrNoRows)
				m.On("CreateUser", mock.Anything, mock.MatchedBy(func(u models.User) bool {
					return u.Username == "testuser" && u.Email == "test@example.com"
				})).Return(1, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "duplicate username",
			input: models.User{
				BaseModel: models.BaseModel{ID: 1},
				Username:  "existinguser",
				Password:  "password123",
				Email:     "test@example.com",
				UserType:  models.Resident,
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByUsername", "existinguser").Return(&models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "existinguser",
				}, nil)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			handler := NewUserHandler(mockRepo)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/user/signup", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.SignUp(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	tests := []struct {
		name           string
		input          map[string]string
		mockSetup      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful login",
			input: map[string]string{
				"username": "testuser",
				"password": "correctpassword",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(&models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "testuser",
					Password:  string(hashedPassword),
					UserType:  models.Resident,
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "invalid password",
			input: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByUsername", "testuser").Return(&models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "testuser",
					Password:  string(hashedPassword),
					UserType:  models.Resident,
				}, nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			handler := NewUserHandler(mockRepo)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("POST", "/user/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			handler.Login(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.NotEmpty(t, response["token"])
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetProfile(t *testing.T) {
	mockRepo := new(MockUserRepository)
	handler := NewUserHandler(mockRepo)

	t.Run("authenticated user", func(t *testing.T) {
		testUser := &models.User{
			BaseModel: models.BaseModel{ID: 1},
			Username:  "testuser",
			Email:     "test@example.com",
		}

		mockRepo.On("GetUserByID", 1).Return(testUser, nil)

		req := httptest.NewRequest("GET", "/profile", nil)
		ctx := context.WithValue(req.Context(), middleware.UserIDKey, "1")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.Equal(t, "testuser", response["username"])
	})

	t.Run("unauthenticated", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestUserHandler_UpdateProfile(t *testing.T) {
	tests := []struct {
		name           string
		input          map[string]string
		userID         string
		mockSetup      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name: "successful update",
			input: map[string]string{
				"username":  "updateduser",
				"email":     "updated@example.com",
				"full_name": "Updated Name",
			},
			userID: "1",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByID", 1).Return(&models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "olduser",
					Email:     "old@example.com",
				}, nil)
				m.On("UpdateUser", mock.Anything, mock.MatchedBy(func(u models.User) bool {
					return u.Username == "updateduser" && u.Email == "updated@example.com"
				})).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized",
			input: map[string]string{
				"username": "updateduser",
			},
			userID:         "",
			mockSetup:      func(m *MockUserRepository) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			handler := NewUserHandler(mockRepo)

			body, _ := json.Marshal(tt.input)
			req := httptest.NewRequest("PUT", "/profile", bytes.NewBuffer(body))
			if tt.userID != "" {
				ctx := context.WithValue(req.Context(), middleware.UserIDKey, tt.userID)
				req = req.WithContext(ctx)
			}
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
		userID         string
		mockSetup      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name:   "user found",
			userID: "1",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByID", 1).Return(&models.User{
					BaseModel: models.BaseModel{ID: 1},
					Username:  "testuser",
				}, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetUserByID", 999).Return(nil, sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest("GET", "/user/"+tt.userID, nil)
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
		mockSetup      func(*MockUserRepository)
		expectedStatus int
		expectedCount  int
	}{
		{
			name: "successful fetch",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetAllUsers", mock.Anything).Return([]models.User{
					{
						BaseModel: models.BaseModel{ID: 1},
						Username:  "user1",
					},
					{
						BaseModel: models.BaseModel{ID: 2},
						Username:  "user2",
					},
				}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  2,
		},
		{
			name: "empty result",
			mockSetup: func(m *MockUserRepository) {
				m.On("GetAllUsers", mock.Anything).Return([]models.User{}, nil)
			},
			expectedStatus: http.StatusOK,
			expectedCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest("GET", "/users", nil)
			w := httptest.NewRecorder()

			handler.GetAllUsers(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var response struct {
					Data []models.User `json:"data"`
				}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedCount, len(response.Data))
			}
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {

	tests := []struct {
		name           string
		userID         string
		mockSetup      func(*MockUserRepository)
		expectedStatus int
	}{
		{
			name:   "successful delete",
			userID: "1",
			mockSetup: func(m *MockUserRepository) {
				m.On("DeleteUser", 1).Return(nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:   "user not found",
			userID: "999",
			mockSetup: func(m *MockUserRepository) {
				m.On("DeleteUser", 999).Return(sql.ErrNoRows)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tt.mockSetup(mockRepo)

			handler := NewUserHandler(mockRepo)

			req := httptest.NewRequest("DELETE", "/user/"+tt.userID, nil)
			w := httptest.NewRecorder()

			handler.DeleteUser(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			mockRepo.AssertExpectations(t)
		})
	}
}
