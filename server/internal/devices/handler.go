package devices

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.VenueID == "" || req.Name == "" {
		writeError(w, http.StatusBadRequest, "venue_id and name are required")
		return
	}

	resp, err := h.svc.Register(r.Context(), req)
	if err != nil {
		log.Printf("[%s] register device: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	resp, err := h.svc.Heartbeat(r.Context(), deviceID, req)
	if err != nil {
		if err.Error() == "device not found" {
			writeError(w, http.StatusNotFound, "device not found")
			return
		}
		log.Printf("[%s] heartbeat device %s: %v", chimw.GetReqID(r.Context()), deviceID, err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) Impression(w http.ResponseWriter, r *http.Request) {
	deviceID := chi.URLParam(r, "id")
	var req ImpressionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := h.svc.RecordImpression(r.Context(), deviceID, req); err != nil {
		log.Printf("[%s] impression device %s: %v", chimw.GetReqID(r.Context()), deviceID, err)
		writeError(w, http.StatusBadRequest, "invalid impression request")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"status": "recorded"})
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	devices, err := h.svc.List(r.Context())
	if err != nil {
		log.Printf("[%s] list devices: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if devices == nil {
		devices = []Device{}
	}
	writeJSON(w, http.StatusOK, devices)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
