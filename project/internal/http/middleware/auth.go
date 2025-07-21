package middleware

import (
	"context"
	"fmt"
	"log"

	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/http/utils"
	"github.com/nedaZarei/arcaptcha-internship-2025/neda-arcaptcha-internship-2025/internal/models"
)

type contextKey string

const UserIDKey contextKey = "userID"

type CustomClaims struct {
	EncryptedUserID string          `json:"uid"`
	UserType        models.UserType `json:"user_type"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("arcaptcha-project")

func JWTAuthMiddleware(userMode ...models.UserType) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "Unauthorized - no token", http.StatusUnauthorized)
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			log.Printf("Validating token: %s", tokenStr)

			userID, err := ValidateToken(tokenStr, userMode...)
			if err != nil {
				log.Printf("Token validation failed: %v", err)
				http.Error(w, "Invalid or expired token: "+err.Error(), http.StatusUnauthorized)
				return
			}

			log.Printf("Token validated successfully for user: %s", userID)
			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GenerateToken(userID string, userType models.UserType) (string, error) {
	encryptedID, err := utils.Encrypt(userID)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt user ID: %w", err)
	}

	expirationTime := time.Now().Add(24 * time.Hour)

	claims := &CustomClaims{
		EncryptedUserID: encryptedID,
		UserType:        userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateToken(tokenStr string, userType ...models.UserType) (string, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return "", fmt.Errorf("token parsing failed: %w", err) // Wrap the error
	}

	if !token.Valid {
		return "", errors.New("invalid token")
	}

	claims, ok := token.Claims.(*CustomClaims)
	if !ok {
		return "", errors.New("invalid token claims")
	}

	//skipping user type validation if no types are specified
	if len(userType) > 0 {
		validType := false
		for _, t := range userType {
			if t == claims.UserType {
				validType = true
				break
			}
		}
		if !validType {
			return "", fmt.Errorf("invalid user type, got %s want one of %v", claims.UserType, userType)
		}
	}

	decryptedID, err := utils.Decrypt(claims.EncryptedUserID)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt user ID: %w", err)
	}

	return decryptedID, nil
}

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		log.Printf("%s %s %s - Started", r.RemoteAddr, r.Method, r.URL.Path)

		ww := &responseWrapper{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(ww, r)

		duration := time.Since(start)
		log.Printf("%s %s %s - Completed in %v with status %d",
			r.RemoteAddr, r.Method, r.URL.Path, duration, ww.statusCode)
	})
}

func CorsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// captures the status code for logging
type responseWrapper struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWrapper) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	return rw.ResponseWriter.Write(b)
}
