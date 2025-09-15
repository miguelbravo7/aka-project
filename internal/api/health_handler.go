package api

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

type HealthHandler struct {
	DB    *pgxpool.Pool
	Redis *redis.Client
}

type HealthStatus struct {
	Status string            `json:"status"`
	Checks map[string]string `json:"checks"`
}

func (h *HealthHandler) Check(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.DB.Ping(ctx); err != nil {
		dbStatus = "error"
	}

	redisStatus := "ok"
	if _, err := h.Redis.Ping(ctx).Result(); err != nil {
		redisStatus = "error"
	}

	status := "ok"
	if dbStatus == "error" || redisStatus == "error" {
		status = "error"
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(HealthStatus{
		Status: status,
		Checks: map[string]string{
			"database": dbStatus,
			"redis":    redisStatus,
		},
	})
}
