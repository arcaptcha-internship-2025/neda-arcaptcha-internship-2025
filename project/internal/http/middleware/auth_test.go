package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestJWTAuthMiddleware(t *testing.T) {
	//a valid token
	validToken, _ := GenerateToken("123", models.Manager)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := JWTAuthMiddleware(models.Manager)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				userID := r.Context().Value(UserIDKey)
				if tt.expectedUserID != "" {
					assert.Equal(t, tt.expectedUserID, userID)
				}
				w.WriteHeader(http.StatusOK)
			}))

			w := httptest.NewRecorder()
			handler.ServeHTTP(w, tt.setupRequest())

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGenerateAndValidateToken(t *testing.T) {
	//token generation and validation
	userID := "123"
	userType := models.Manager

	token, err := GenerateToken(userID, userType)
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	//valid token
	validatedID, err := ValidateToken(token, userType)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedID)

	//invalid user type
	_, err = ValidateToken(token, models.Resident)
	assert.Error(t, err)

	//expired token
	expiredToken := jwt.NewWithClaims(jwt.SigningMethodHS256, CustomClaims{
		EncryptedUserID: "encrypted123",
		UserType:        userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	})
	signedExpiredToken, _ := expiredToken.SignedString(jwtSecret)
	_, err = ValidateToken(signedExpiredToken, userType)
	assert.Error(t, err)
}
