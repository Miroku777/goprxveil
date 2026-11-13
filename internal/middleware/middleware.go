package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/Miroku77/goprxveil/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

// LoginMiddleware checks the Authorization header for a valid JWT Bearer token.
// It validates the token, extracts the user ID, and stores it in the request context.
// Returns 401 Unauthorized if the token is missing, invalid, or expired.
func LoginMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		token, err := jwt.ParseWithClaims(tokenString, &models.JWTClaims{}, func(token *jwt.Token) (any, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("SECRET_KEY")), nil
		})

		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(*models.JWTClaims); ok && token.Valid {
			ctx := context.WithValue(r.Context(), "userId", claims.UserId)
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
	})
}
