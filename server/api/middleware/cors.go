package middleware

import (
	"net/http"
	"strings"
)

// CORS enforces an explicit origin allowlist. allowedOrigins is a comma-separated
// list of permitted origins (e.g. "http://localhost:3000,https://portal.adlite.io").
// Empty string disables CORS headers entirely (same-origin only).
func CORS(allowedOrigins string) func(http.Handler) http.Handler {
	allowed := make(map[string]bool)
	for _, o := range strings.Split(allowedOrigins, ",") {
		if trimmed := strings.TrimSpace(o); trimmed != "" {
			allowed[trimmed] = true
		}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if len(allowed) > 0 && allowed[origin] {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Accept, Authorization, Content-Type")
				w.Header().Set("Access-Control-Max-Age", "300")
				w.Header().Add("Vary", "Origin")
			}
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
