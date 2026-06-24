package analytics

import (
	"encoding/json"
	"log"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/mapletondesign/ad_lite/api/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// AdminImpressions handles GET /analytics/impressions for admin.
func (h *Handler) AdminImpressions(w http.ResponseWriter, r *http.Request) {
	f := ImpressionFilter{
		BookingID: r.URL.Query().Get("booking_id"),
		DeviceID:  r.URL.Query().Get("device_id"),
		From:      r.URL.Query().Get("from"),
		To:        r.URL.Query().Get("to"),
	}
	h.respond(w, r, f, "admin impressions")
}

// AdvertiserImpressions handles GET /advertiser/analytics/impressions, scoped to the caller's advertiser_id.
func (h *Handler) AdvertiserImpressions(w http.ResponseWriter, r *http.Request) {
	advertiserID, _ := r.Context().Value(middleware.AdvertiserIDKey).(string)
	if advertiserID == "" {
		writeError(w, http.StatusForbidden, "advertiser account not linked to this user")
		return
	}
	f := ImpressionFilter{
		AdvertiserID: advertiserID,
		BookingID:    r.URL.Query().Get("booking_id"),
		DeviceID:     r.URL.Query().Get("device_id"),
		From:         r.URL.Query().Get("from"),
		To:           r.URL.Query().Get("to"),
	}
	h.respond(w, r, f, "advertiser impressions")
}

func (h *Handler) respond(w http.ResponseWriter, r *http.Request, f ImpressionFilter, op string) {
	rows, err := h.svc.Impressions(r.Context(), f)
	if err != nil {
		log.Printf("[%s] %s: %v", chimw.GetReqID(r.Context()), op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, rows)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
