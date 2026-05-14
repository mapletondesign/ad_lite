package assets

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	chimw "github.com/go-chi/chi/v5/middleware"
)

const maxMemory = 50 << 20 // 50 MB in memory; remainder spills to disk

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) Upload(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseMultipartForm(maxMemory); err != nil {
		writeError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "field 'file' is required")
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	url, err := h.svc.Upload(r.Context(), contentType, file, header.Size)
	if err != nil {
		if isUnsupportedType(err) {
			writeError(w, http.StatusUnsupportedMediaType, err.Error())
			return
		}
		log.Printf("[%s] asset upload: %v", chimw.GetReqID(r.Context()), err)
		writeError(w, http.StatusInternalServerError, "upload failed")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"url": url})
}

func isUnsupportedType(err error) bool {
	return err != nil && strings.HasPrefix(err.Error(), "unsupported content type:")
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}
