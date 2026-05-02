package api

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mapletondesign/ad_pack/api/middleware"
	"github.com/mapletondesign/ad_pack/internal/devices"
	"github.com/mapletondesign/ad_pack/internal/slots"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(middleware.Logger)
	r.Use(chimw.RequestID)

	deviceSvc := devices.NewService(db)
	deviceHandler := devices.NewHandler(deviceSvc)

	slotSvc := slots.NewService(db)
	slotHandler := slots.NewHandler(slotSvc)

	r.Get("/health", healthHandler(db, rdb))

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/devices", func(r chi.Router) {
			r.Post("/register", deviceHandler.Register)
			r.Get("/", deviceHandler.List)
			r.Post("/{id}/heartbeat", deviceHandler.Heartbeat)
		})

		r.Route("/slots", func(r chi.Router) {
			r.Get("/", slotHandler.List)
			r.Post("/", slotHandler.Create)
			r.Patch("/{id}", slotHandler.Update)
		})
		r.Route("/bookings", func(r chi.Router) {
			r.Post("/", stubHandler("booking creation — coming in Stage 6"))
		})
		r.Route("/analytics", func(r chi.Router) {
			r.Get("/impressions", stubHandler("impression reporting — coming in Stage 2"))
		})
	})

	return r
}

func healthHandler(db *pgxpool.Pool, rdb *redis.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		dbOk := db.Ping(context.Background()) == nil
		redisOk := rdb.Ping(context.Background()).Err() == nil

		status := "ok"
		code := http.StatusOK
		if !dbOk || !redisOk {
			status = "degraded"
			code = http.StatusServiceUnavailable
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		json.NewEncoder(w).Encode(map[string]any{
			"status": status,
			"db":     dbOk,
			"redis":  redisOk,
		})
	}
}

func stubHandler(msg string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotImplemented)
		json.NewEncoder(w).Encode(map[string]string{"message": msg})
	}
}
