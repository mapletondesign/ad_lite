package slots

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

	slot, err := h.svc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.auditSvc.Log(actorID, "create_slot", "ad_slots", slot.ID, map[string]any{
		"device_id":   slot.DeviceID,
		"price_cents": slot.PriceCents,
	})
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
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
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
	h.auditSvc.Log(actorID, "update_slot", "ad_slots", slot.ID, map[string]any{"status": slot.Status})
	writeJSON(w, http.StatusOK, slot)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	actorID, _ := r.Context().Value(middleware.UserIDKey).(string)
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id); err != nil {
		switch err.Error() {
		case "slot not found":
			writeError(w, http.StatusNotFound, "slot not found")
		case "slot has active or pending bookings and cannot be deleted":
			writeError(w, http.StatusConflict, err.Error())
		default:
			log.Printf("[%s] delete slot %s: %v", chimw.GetReqID(r.Context()), id, err)
			writeError(w, http.StatusInternalServerError, "internal server error")
		}
		return
	}
	h.auditSvc.Log(actorID, "delete_slot", "ad_slots", id, nil)
	w.WriteHeader(http.StatusNoContent)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
