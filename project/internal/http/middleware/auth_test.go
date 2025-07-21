package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthMiddleware(t *testing.T) {
	jwtSecret = []byte("arcaptcha-project")

	validToken, err := GenerateToken("123", models.Manager)
	if err != nil {
		t.Fatalf("Failed to generate token: %v", err)
	}
	assert.NotEmpty(t, validToken)
	t.Logf("Generated valid token: %s", validToken)

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		expectedUserID string
	}{
		{
			name: "valid token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer "+validToken)
				return req
			},
			expectedStatus: http.StatusOK,
			expectedUserID: "123",
		},
		{
			name: "missing token",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("GET", "/", nil)
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid token",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "Bearer invalidtoken")
				return req
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "malformed authorization header",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("GET", "/", nil)
				req.Header.Set("Authorization", "InvalidBearer "+validToken)
				return req
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			handler := JWTAuthMiddleware(models.Manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx = r.Context()
				userID := ctx.Value(UserIDKey)
				t.Logf("Context userID: %v", userID)
				if tt.expectedUserID != "" {
					assert.Equal(t, tt.expectedUserID, userID)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := tt.setupRequest()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			t.Logf("Test case: %s", tt.name)
			t.Logf("Response status: %d", w.Code)
			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.name == "valid token" && w.Code == http.StatusOK {
				userID := ctx.Value(UserIDKey)
				assert.Equal(t, tt.expectedUserID, userID)
			}
		})
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	userID := "123"
	userType := models.Manager

	token, err := GenerateToken(userID, userType)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	validatedID, err := ValidateToken(token, userType)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedID)

	_, err = ValidateToken(token, models.Resident)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid user type")

	validatedID, err = ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedID)

	validatedID, err = ValidateToken(token, models.Manager, models.Resident)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedID)
}

func TestGenerateAndValidateToken_InvalidToken(t *testing.T) {
	_, err := ValidateToken("invalid.token.here")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token parsing failed")

	_, err = ValidateToken("")
	assert.Error(t, err)
}
