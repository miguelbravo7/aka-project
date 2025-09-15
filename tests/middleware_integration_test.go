package tests

import (
	"aka-project/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func TestAPIKeyAndRateLimit(t *testing.T) {
	// Start in-memory Redis
	mr, err := miniredis.Run()
	assert.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	// Create middleware instance
	apiKey := "secret-key"
	mw, _ := middleware.NewMiddleware(rdb, "2-S", apiKey) // limit 2 requests

	// Protected handler
	handler := mw.RequireAPIKey(mw.RateLimit(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})))

	// Helper to perform request with API key
	doReq := func() *httptest.ResponseRecorder {
		req := httptest.NewRequest("GET", "/", nil)
		req.Header.Set("X-API-Key", apiKey)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, req)
		return w
	}

	// --- First request should pass ---
	resp := doReq()
	assert.Equal(t, http.StatusOK, resp.Code, "Expected status OK")
	assert.Empty(t, resp.Body.String(), "Expected empty response body")
	assert.Equal(t, "2", resp.Header().Get("X-RateLimit-Limit"), "Expected rate limit limit to be 2")
	assert.Equal(t, "1", resp.Header().Get("X-RateLimit-Remaining"), "Expected rate limit remaining to be 1")
	assert.NotEmpty(t, resp.Header().Get("X-RateLimit-Reset"), "Expected rate limit reset header to be present")

	// --- Second request should pass ---
	resp = doReq()
	assert.Equal(t, http.StatusOK, resp.Code, "Expected status OK")
	assert.Empty(t, resp.Body.String(), "Expected empty response body")
	assert.Equal(t, "2", resp.Header().Get("X-RateLimit-Limit"), "Expected rate limit limit to be 2")
	assert.Equal(t, "0", resp.Header().Get("X-RateLimit-Remaining"), "Expected rate limit remaining to be 0")
	assert.NotEmpty(t, resp.Header().Get("X-RateLimit-Reset"), "Expected rate limit reset header to be present")

	// --- Third request should be rate limited ---
	resp = doReq()
	assert.Equal(t, http.StatusTooManyRequests, resp.Code, "Expected status Too Many Requests")
	assert.Contains(t, resp.Body.String(), "rate limit exceeded", "Expected response body to contain 'rate limit exceeded'")
	assert.Equal(t, "2", resp.Header().Get("X-RateLimit-Limit"), "Expected rate limit limit to be 2")
	assert.Equal(t, "0", resp.Header().Get("X-RateLimit-Remaining"), "Expected rate limit remaining to be 0")
	assert.NotEmpty(t, resp.Header().Get("X-RateLimit-Reset"), "Expected rate limit reset header to be present")

	// --- Test missing API key ---
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusUnauthorized, w.Code, "Expected status Unauthorized")
	assert.Contains(t, w.Body.String(), "unauthorized", "Expected response body to contain 'unauthorized'")
}
