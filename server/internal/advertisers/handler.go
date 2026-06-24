package advertisers

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
	a, err := h.svc.Create(r.Context(), req, actorID)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	h.auditSvc.Log(actorID, "create_advertiser", "advertisers", a.ID, map[string]any{"name": a.Name, "email": a.Email})
	writeJSON(w, http.StatusCreated, a)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	a, err := h.svc.Get(r.Context(), id)
	if err != nil {
		writeError(w, http.StatusNotFound, "advertiser not found")
		return
	}
	writeJSON(w, http.StatusOK, a)
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
