package slots

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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slot, err := h.svc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, slot)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	f := ListFilter{
		DeviceID: r.URL.Query().Get("device_id"),
		Status:   r.URL.Query().Get("status"),
	}

	slots, err := h.svc.List(r.Context(), f)
	if err != nil {
		log.Printf("[%s] list slots: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if slots == nil {
		slots = []Slot{}
	}
	writeJSON(w, http.StatusOK, slots)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var req UpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	slot, err := h.svc.Update(r.Context(), id, req)
	if err != nil {
		if err.Error() == "slot not found" {
			writeError(w, http.StatusNotFound, "slot not found")
			return
		}
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, slot)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
