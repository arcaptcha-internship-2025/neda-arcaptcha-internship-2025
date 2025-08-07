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
	assert.Contains(t, err.Error(), "not authorized for user type")

	validatedID, err = ValidateToken(token)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedID)

	validatedID, err = ValidateToken(token, models.Manager, models.Resident)
	assert.NoError(t, err)
	assert.Equal(t, userID, validatedID)
}

func TestGenerateAndValidateToken_InvalidToken(t *testing.T) {
	//malformed token
	_, err := ValidateToken("invalid.token.blahblah")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "token parsing failed")

	//empty token
	_, err = ValidateToken("")
	assert.Error(t, err)
}

func TestIdempotentKeyMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
		expectedKey    string
	}{
		{
			name: "valid idempotent key",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/", nil)
				req.Header.Set("X-Idempotent-Key", "test-key-123")
				return req
			},
			expectedStatus: http.StatusOK,
			expectedKey:    "test-key-123",
		},
		{
			name: "missing idempotent key",
			setupRequest: func() *http.Request {
				return httptest.NewRequest("POST", "/", nil)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty idempotent key",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest("POST", "/", nil)
				req.Header.Set("X-Idempotent-Key", "")
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context
			handler := IdempotentKeyMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				ctx = r.Context()
				if tt.expectedKey != "" {
					key := ctx.Value(IdempotentKey)
					assert.Equal(t, tt.expectedKey, key)
				}
				w.WriteHeader(http.StatusOK)
			}))

			req := tt.setupRequest()
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.name == "valid idempotent key" && w.Code == http.StatusOK {
				key := ctx.Value(IdempotentKey)
				assert.Equal(t, tt.expectedKey, key)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	handler := LoggingMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())
}

func TestCorsMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		checkHeaders   bool
	}{
		{
			name:           "OPTIONS request",
			method:         "OPTIONS",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "GET request",
			method:         "GET",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
		{
			name:           "POST request",
			method:         "POST",
			expectedStatus: http.StatusOK,
			checkHeaders:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := CorsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "OPTIONS" {
					w.WriteHeader(http.StatusOK)
				}
			}))

			req := httptest.NewRequest(tt.method, "/", nil)
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkHeaders {
				assert.Equal(t, "*", w.Header().Get("Access-Control-Allow-Origin"))
				assert.Equal(t, "GET, POST, PUT, DELETE, OPTIONS", w.Header().Get("Access-Control-Allow-Methods"))
				assert.Equal(t, "Content-Type, Authorization", w.Header().Get("Access-Control-Allow-Headers"))
			}
		})
	}
}

func TestResponseWrapper(t *testing.T) {
	w := httptest.NewRecorder()
	wrapper := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

	wrapper.WriteHeader(http.StatusCreated)
	assert.Equal(t, http.StatusCreated, wrapper.statusCode)

	data := []byte("test data")
	n, err := wrapper.Write(data)
	assert.NoError(t, err)
	assert.Equal(t, len(data), n)
	assert.Equal(t, string(data), w.Body.String())
}
