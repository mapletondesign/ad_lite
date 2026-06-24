package bookings

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/mapletondesign/ad_lite/api/middleware"
	"github.com/mapletondesign/ad_lite/internal/audit"
)

type Handler struct {
	svc      *Service
	auditSvc *audit.Service
}

func NewHandler(svc *Service, auditSvc *audit.Service) *Handler {
	return &Handler{svc: svc, auditSvc: auditSvc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)

	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.svc.Create(r.Context(), req, actorID)
	if err != nil {
		writeBookingError(w, r, "create booking", err)
		return
	}
	h.auditSvc.Log(actorID, "create_booking", "bookings", b.ID, map[string]any{
		"slot_id":       b.SlotID,
		"advertiser_id": b.AdvertiserID,
		"starts_on":     b.StartsOn,
		"ends_on":       b.EndsOn,
	})
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f := ListFilter{
		SlotID:       r.URL.Query().Get("slot_id"),
		AdvertiserID: r.URL.Query().Get("advertiser_id"),
		Status:       r.URL.Query().Get("status"),
	}

	bookings, err := h.svc.List(r.Context(), f)
	if err != nil {
		log.Printf("[%s] list bookings: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if bookings == nil {
		bookings = []Booking{}
	}
	writeJSON(w, http.StatusOK, bookings)
}

func (h *Handler) CreateForAdvertiser(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	advertiserID, _ := r.Context().Value(middleware.AdvertiserIDKey).(string)
	if advertiserID == "" {
		writeError(w, http.StatusForbidden, "advertiser account not linked to this user")
		return
	}

	var body struct {
		SlotID      string  `json:"slot_id"`
		StartsOn    string  `json:"starts_on"`
		EndsOn      string  `json:"ends_on"`
		CreativeURL *string `json:"creative_url,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	req := CreateRequest{
		SlotID:       body.SlotID,
		AdvertiserID: advertiserID,
		StartsOn:     body.StartsOn,
		EndsOn:       body.EndsOn,
		CreativeURL:  body.CreativeURL,
	}

	b, err := h.svc.Create(r.Context(), req, actorID)
	if err != nil {
		writeBookingError(w, r, "create advertiser booking", err)
		return
	}
	writeJSON(w, http.StatusCreated, b)
}

func (h *Handler) ListForAdvertiser(w http.ResponseWriter, r *http.Request) {
	advertiserID, _ := r.Context().Value(middleware.AdvertiserIDKey).(string)
	if advertiserID == "" {
		writeError(w, http.StatusForbidden, "advertiser account not linked to this user")
		return
	}
	f := ListFilter{
		AdvertiserID: advertiserID,
		Status:       r.URL.Query().Get("status"),
	}
	bookings, err := h.svc.List(r.Context(), f)
	if err != nil {
		log.Printf("[%s] list advertiser bookings: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if bookings == nil {
		bookings = []Booking{}
	}
	writeJSON(w, http.StatusOK, bookings)
}

func (h *Handler) AdminUpdate(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	id := chi.URLParam(r, "id")

	var req AdminUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	b, err := h.svc.AdminUpdate(r.Context(), id, req, actorID)
	if err != nil {
		switch err.Error() {
		case "booking not found":
			writeError(w, http.StatusNotFound, "booking not found")
		case "invalid status":
			writeError(w, http.StatusBadRequest, "invalid status value")
		default:
			log.Printf("[%s] admin update booking: %v", chimw.GetReqID(r.Context()), err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	details := map[string]any{"status": b.Status}
	if req.CreativeURL != nil {
		details["creative_url_updated"] = true
	}
	h.auditSvc.Log(actorID, "update_booking", "bookings", b.ID, details)
	writeJSON(w, http.StatusOK, b)
}

func (h *Handler) AdvertiserUpdate(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	id := chi.URLParam(r, "id")
	advertiserID, _ := r.Context().Value(middleware.AdvertiserIDKey).(string)
	if advertiserID == "" {
		writeError(w, http.StatusForbidden, "advertiser account not linked to this user")
		return
	}
	var req AdvertiserUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	b, err := h.svc.AdvertiserUpdate(r.Context(), id, advertiserID, req, actorID)
	if err != nil {
		switch err.Error() {
		case "booking not found":
			writeError(w, http.StatusNotFound, "booking not found")
		case "forbidden":
			writeError(w, http.StatusForbidden, "you do not own this booking")
		case "booking cannot be cancelled in its current state":
			writeError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("[%s] advertiser update booking: %v", chimw.GetReqID(r.Context()), err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func writeBookingError(w http.ResponseWriter, r *http.Request, op string, err error) {
	switch err.Error() {
	case "slot not found":
		writeError(w, http.StatusNotFound, "slot not found")
	case "slot is not available for the requested date range",
		"slot is paused and cannot be booked":
		writeError(w, http.StatusConflict, err.Error())
	case "slot_id is required", "advertiser_id is required",
		"starts_on and ends_on are required",
		"starts_on must be on or before ends_on":
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		log.Printf("[%s] %s: %v", chimw.GetReqID(r.Context()), op, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
