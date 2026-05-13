package advertisers

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
	a, err := h.svc.Create(r.Context(), req)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	advertisers, err := h.svc.List(r.Context())
	if err != nil {
		log.Printf("[%s] list advertisers: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if advertisers == nil {
		advertisers = []Advertiser{}
	}
	writeJSON(w, http.StatusOK, advertisers)
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}
