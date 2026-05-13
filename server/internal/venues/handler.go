package venues

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
	v, err := h.svc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, v)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	venues, err := h.svc.List(r.Context())
	if err != nil {
		log.Printf("[%s] list venues: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if venues == nil {
		venues = []Venue{}
	}
	writeJSON(w, http.StatusOK, venues)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
