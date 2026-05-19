package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

// IPRateLimit limits requests by client IP using a fixed-window counter in Redis.
// On Redis failure the request is rejected to protect the endpoint.
func IPRateLimit(rdb *redis.Client, prefix string, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := clientIP(r)
			key := "rl:" + prefix + ":" + ip
			count, err := rdb.Incr(r.Context(), key).Result()
			if err != nil {
				http.Error(w, `{"error":"service unavailable"}`, http.StatusServiceUnavailable)
				return
			}
			if count == 1 {
				rdb.Expire(r.Context(), key, window)
			}
			if count > int64(limit) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("X-Forwarded-For"); fwd != "" {
		return strings.SplitN(fwd, ",", 2)[0]
	}
	if ip, _, ok := strings.Cut(r.RemoteAddr, ":"); ok {
		return ip
	}
	return r.RemoteAddr
}

// DeviceRateLimit limits device write endpoints using a fixed-window counter in Redis.
// On Redis failure the request is allowed through (fail open) to avoid taking devices offline.
func DeviceRateLimit(rdb *redis.Client, limit int, window time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			deviceID, _ := r.Context().Value(DeviceIDKey).(string)
			if deviceID == "" {
				next.ServeHTTP(w, r)
				return
			}

			key := "rl:device:" + deviceID
			count, err := rdb.Incr(r.Context(), key).Result()
			if err != nil {
				next.ServeHTTP(w, r)
				return
			}
			if count == 1 {
				rdb.Expire(r.Context(), key, window)
			}
			if count > int64(limit) {
				http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
