package middleware

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// RequireValidUUID rejects requests where the named URL parameter is not a valid UUID.
func RequireValidUUID(param string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, err := uuid.Parse(chi.URLParam(r, param)); err != nil {
				http.Error(w, `{"error":"invalid id"}`, http.StatusBadRequest)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
