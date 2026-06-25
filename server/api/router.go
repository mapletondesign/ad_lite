package api

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mapletondesign/ad_lite/api/middleware"
	"github.com/mapletondesign/ad_lite/internal/advertisers"
	"github.com/mapletondesign/ad_lite/internal/analytics"
	"github.com/mapletondesign/ad_lite/internal/assets"
	"github.com/mapletondesign/ad_lite/internal/audit"
	"github.com/mapletondesign/ad_lite/internal/auth"
	"github.com/mapletondesign/ad_lite/internal/bookings"
	"github.com/mapletondesign/ad_lite/internal/devices"
	"github.com/mapletondesign/ad_lite/internal/scheduler"
	"github.com/mapletondesign/ad_lite/internal/slots"
	"github.com/mapletondesign/ad_lite/internal/venues"
	"github.com/redis/go-redis/v9"
)

func NewRouter(db *pgxpool.Pool, rdb *redis.Client, privateKey *rsa.PrivateKey, publicKey *rsa.PublicKey, assetsSvc *assets.Service) http.Handler {
	r := chi.NewRouter()

	r.Use(chimw.Recoverer)
	r.Use(chimw.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.CORS(envOr("PORTAL_ORIGIN", "")))
	r.Use(middleware.SecurityHeaders)

	authSvc := auth.NewService(db, privateKey, publicKey)
	authHandler := auth.NewHandler(authSvc)

	auditSvc := audit.NewService(db)

	venueSvc := venues.NewService(db)
	venueHandler := venues.NewHandler(venueSvc, auditSvc)

	advertiserSvc := advertisers.NewService(db)
	advertiserHandler := advertisers.NewHandler(advertiserSvc, auditSvc)

	schedSvc := scheduler.NewService(db, rdb)
	deviceSvc := devices.NewService(db, schedSvc, privateKey, publicKey)
	deviceHandler := devices.NewHandler(deviceSvc)

	slotSvc := slots.NewService(db)
	slotHandler := slots.NewHandler(slotSvc, auditSvc)

	bookingSvc := bookings.NewService(db, envOr("CDN_BASE_URL", ""))
	bookingHandler := bookings.NewHandler(bookingSvc, auditSvc)

	analyticsSvc := analytics.NewService(db)
	analyticsHandler := analytics.NewHandler(analyticsSvc)

	r.Get("/health", healthHandler(db, rdb))

	r.Route("/api/v1", func(r chi.Router) {
		// Asset upload — 100 MB body limit, advertiser or admin JWT
		if assetsSvc != nil {
			assetsHandler := assets.NewHandler(assetsSvc)
			r.Group(func(r chi.Router) {
				r.Use(middleware.MaxBodySize(100 << 20))
				r.Use(middleware.UserAuth(publicKey, "admin", "advertiser"))
				r.Post("/assets/upload", assetsHandler.Upload)
			})
		}

		// All other API routes — 1 MB body limit
		r.Group(func(r chi.Router) {
			r.Use(middleware.MaxBodySize(1 << 20))

			// Auth — open, but rate-limited: 10 req/min per IP
			r.Group(func(r chi.Router) {
				r.Use(middleware.IPRateLimit(rdb, "auth", 10, time.Minute))
				r.Post("/auth/register", authHandler.Register)
				r.Post("/auth/login", authHandler.Login)
				r.Post("/auth/refresh", authHandler.Refresh)
				r.Post("/auth/device/login", deviceHandler.DeviceLogin)
			})

			// Management routes — require admin JWT
			r.Group(func(r chi.Router) {
				r.Use(middleware.UserAuth(publicKey, "admin"))

				r.Route("/venues", func(r chi.Router) {
					r.Post("/", venueHandler.Create)
					r.Get("/", venueHandler.List)
					r.With(middleware.RequireValidUUID("id")).Get("/{id}", venueHandler.Get)
				})

				r.Route("/advertisers", func(r chi.Router) {
					r.Post("/", advertiserHandler.Create)
					r.Get("/", advertiserHandler.List)
					r.With(middleware.RequireValidUUID("id")).Get("/{id}", advertiserHandler.Get)
				})

				r.Route("/slots", func(r chi.Router) {
					r.Get("/", slotHandler.List)
					r.Post("/", slotHandler.Create)
					r.With(middleware.RequireValidUUID("id")).Patch("/{id}", slotHandler.Update)
					r.With(middleware.RequireValidUUID("id")).Delete("/{id}", slotHandler.Delete)
				})

				r.Route("/bookings", func(r chi.Router) {
					r.Post("/", bookingHandler.Create)
					r.Get("/", bookingHandler.List)
					r.With(middleware.RequireValidUUID("id")).Patch("/{id}", bookingHandler.AdminUpdate)
				})

				r.Get("/devices", deviceHandler.List)

				r.Route("/analytics", func(r chi.Router) {
					r.Get("/impressions", analyticsHandler.AdminImpressions)
				})
			})

			// Advertiser-scoped routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.UserAuth(publicKey, "advertiser"))
				r.Get("/advertiser/slots", slotHandler.List)
				r.Get("/advertiser/bookings", bookingHandler.ListForAdvertiser)
				r.Post("/advertiser/bookings", bookingHandler.CreateForAdvertiser)
				r.With(middleware.RequireValidUUID("id")).Patch("/advertiser/bookings/{id}", bookingHandler.AdvertiserUpdate)
				r.Get("/advertiser/analytics/impressions", analyticsHandler.AdvertiserImpressions)
			})

			// Venue-scoped routes
			r.Group(func(r chi.Router) {
				r.Use(middleware.UserAuth(publicKey, "venue"))
				r.Get("/venue/devices", deviceHandler.ListForVenue)
			})

			// Device registration — open (device obtains its token here)
			r.Post("/devices/register", deviceHandler.Register)

			// Device routes — require device JWT
			r.Group(func(r chi.Router) {
				r.Use(middleware.DeviceAuth(publicKey))

				// Heartbeat and impression are rate-limited: 30 req/min per device
				r.Group(func(r chi.Router) {
					r.Use(middleware.DeviceRateLimit(rdb, 30, time.Minute))
					r.Post("/devices/{id}/heartbeat", deviceHandler.Heartbeat)
					r.Post("/devices/{id}/impression", deviceHandler.Impression)
				})
			})
		})
	})

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
