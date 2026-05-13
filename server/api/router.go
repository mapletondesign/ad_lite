package api

import (
	"context"
	"encoding/json"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mapletondesign/ad_pack/api/middleware"
	"github.com/mapletondesign/ad_pack/internal/advertisers"
	"github.com/mapletondesign/ad_pack/internal/bookings"
	"github.com/mapletondesign/ad_pack/internal/devices"
	"github.com/mapletondesign/ad_pack/internal/scheduler"
	"github.com/mapletondesign/ad_pack/internal/slots"
	"github.com/mapletondesign/ad_pack/internal/venues"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *pgxpool.Pool, rdb *redis.Client) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.MaxBodySize(1 << 20)) // 1 MB

	venueSvc := venues.NewService(db)
	venueHandler := venues.NewHandler(venueSvc)

	advertiserSvc := advertisers.NewService(db)
	advertiserHandler := advertisers.NewHandler(advertiserSvc)

	schedSvc := scheduler.NewService(db)
	deviceSvc := devices.NewService(db, schedSvc)
	deviceHandler := devices.NewHandler(deviceSvc)

	slotSvc := slots.NewService(db)
	slotHandler := slots.NewHandler(slotSvc)

	bookingSvc := bookings.NewService(db)
	bookingHandler := bookings.NewHandler(bookingSvc)

	r.Get("/health", healthHandler(db, rdb))

	r.Route("/api/v1", func(r chi.Router) {
		// Management routes — require API key
		r.Group(func(r chi.Router) {
			r.Use(middleware.APIKeyAuth)

			r.Route("/venues", func(r chi.Router) {
				r.Post("/", venueHandler.Create)
				r.Get("/", venueHandler.List)
			})

			r.Route("/advertisers", func(r chi.Router) {
				r.Post("/", advertiserHandler.Create)
				r.Get("/", advertiserHandler.List)
			})

			r.Route("/slots", func(r chi.Router) {
				r.Get("/", slotHandler.List)
				r.Post("/", slotHandler.Create)
				r.Patch("/{id}", slotHandler.Update)
			})

			r.Route("/bookings", func(r chi.Router) {
				r.Post("/", bookingHandler.Create)
				r.Get("/", bookingHandler.List)
			})

			r.Route("/analytics", func(r chi.Router) {
				r.Get("/impressions", stubHandler("impression reporting — coming in Stage 4"))
			})
		})

		// Device registration — open (device needs to obtain its token)
		r.Post("/devices/register", deviceHandler.Register)

		// Device routes — require device JWT
		r.Group(func(r chi.Router) {
			r.Use(middleware.DeviceAuth)

			r.Get("/devices", deviceHandler.List)
			r.Post("/devices/{id}/heartbeat", deviceHandler.Heartbeat)
			r.Post("/devices/{id}/impression", deviceHandler.Impression)
		})
	})

	assetsDir := envOr("ASSETS_DIR", "./assets")
	r.Handle("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir))))

	clientDir := envOr("CLIENT_DIR", "../client")
	r.Handle("/*", http.FileServer(http.Dir(clientDir)))

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

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
