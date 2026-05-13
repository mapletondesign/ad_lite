package middleware

import (
	"context"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const DeviceIDKey contextKey = "device_id"

func DeviceAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			http.Error(w, `{"error":"server misconfigured"}`, http.StatusInternalServerError)
			return
		}

		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			http.Error(w, `{"error":"missing or invalid authorization header"}`, http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || claims["role"] != "device" {
			http.Error(w, `{"error":"invalid token claims"}`, http.StatusForbidden)
			return
		}

		deviceID, _ := claims["sub"].(string)
		ctx := context.WithValue(r.Context(), DeviceIDKey, deviceID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func APIKeyAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := os.Getenv("API_KEY")
		if expected == "" {
			http.Error(w, `{"error":"server misconfigured"}`, http.StatusInternalServerError)
			return
		}
		if r.Header.Get("X-API-Key") != expected {
			http.Error(w, `{"error":"invalid or missing API key"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}
