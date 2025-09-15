package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"aka-project/internal/api"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func newTestDB(t *testing.T) *pgxpool.Pool {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://postgres:password@localhost:5432/myapp?sslmode=disable"
	} else {
		dbURL = strings.Replace(dbURL, "db:", "localhost:", 1)
	}

	db, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("could not connect to database: %v", err)
	}
	return db
}

func newTestRedis(t *testing.T) *redis.Client {
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	} else {
		redisAddr = strings.Replace(redisAddr, "redis:", "localhost:", 1)
	}

	client := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})
	return client
}

func TestHealthCheck(t *testing.T) {
	// Mock DB and Redis
	db, redis := newTestDB(t), newTestRedis(t)

	// Handler
	hh := &api.HealthHandler{DB: db, Redis: redis}

	// Request
	req := httptest.NewRequest(http.MethodGet, "/healthcheck", nil)
	w := httptest.NewRecorder()

	hh.Check(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)

	var resp api.HealthStatus
	err := json.NewDecoder(w.Body).Decode(&resp)
	assert.NoError(t, err)

	assert.Equal(t, "error", resp.Status)
	assert.Equal(t, "error", resp.Checks["database"])
	assert.Equal(t, "error", resp.Checks["redis"])
}
