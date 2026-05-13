package bookings

import (
	"encoding/json"
	"log"
	"net/http"

	chimw "github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	b, err := h.svc.Create(r.Context(), req)
	if err != nil {
		switch err.Error() {
		case "slot not found":
			writeError(w, http.StatusNotFound, "slot not found")
		case "slot is not available for the requested date range",
			"slot is paused and cannot be booked":
			writeError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("[%s] create booking: %v", chimw.GetReqID(r.Context()), err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
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

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
