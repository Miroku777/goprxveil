package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/Miroku77/goprxveil/internal/middleware"
	"github.com/Miroku77/goprxveil/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

func TestLoginMiddleware(t *testing.T) {
	os.Setenv("SECRET_KEY", "test_secret")
	defer os.Unsetenv("SECRET_KEY")

	validToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, models.JWTClaims{
		UserId: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(1 * time.Hour)),
		},
	}).SignedString([]byte("test_secret"))

	expiredToken, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, models.JWTClaims{
		UserId: 42,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}).SignedString([]byte("test_secret"))

	nextHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value("userId").(int)
		if userId != 42 {
			t.Errorf("Expected userId 42, got %v", userId)
		}
		w.WriteHeader(http.StatusOK)
	})

	tests := []struct {
		name           string
		authHeader     string
		wantStatusCode int
	}{
		{"No Header", "", http.StatusUnauthorized},
		{"Invalid Token", "Bearer invalid.token.string", http.StatusUnauthorized},
		{"Expired Token", "Bearer " + expiredToken, http.StatusUnauthorized},
		{"Valid Token", "Bearer " + validToken, http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			middleware.LoginMiddleware(nextHandler).ServeHTTP(w, req)

			if w.Code != tt.wantStatusCode {
				t.Errorf("Expected status %d, got %d", tt.wantStatusCode, w.Code)
			}
		})
	}
}
