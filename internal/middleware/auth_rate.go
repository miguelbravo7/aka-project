package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// Rate limit middleware

func (m *Middleware) RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		key := clientKey(r, m.apiKey)
		lv, err := m.limiter.Get(ctx, key)
		if err != nil {
			http.Error(w, "rate limit error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("X-RateLimit-Limit", fmt.Sprintf("%d", lv.Limit))
		w.Header().Set("X-RateLimit-Remaining", fmt.Sprintf("%d", lv.Remaining))
		w.Header().Set("X-RateLimit-Reset", fmt.Sprintf("%d", lv.Reset))

		if lv.Reached {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// API key middleware
func (m *Middleware) RequireAPIKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("X-API-Key")
		if auth != m.apiKey {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func clientKey(r *http.Request, apiKey string) string {
	key := r.Header.Get("X-API-Key")
	if key != "" {
		return "apikey:" + key
	}
	ip := r.Header.Get("X-Real-IP")
	if ip == "" {
		ip = r.RemoteAddr
		if idx := strings.LastIndex(ip, ":"); idx != -1 {
			ip = ip[:idx]
		}
	}
	return "ip:" + ip
}
