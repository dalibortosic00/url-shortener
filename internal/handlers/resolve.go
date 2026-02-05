package handlers

import (
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/go-chi/chi/v5"
)

type ResolveHandler struct {
	service *services.ShortenerService
}

func NewResolveHandler(service *services.ShortenerService) *ResolveHandler {
	return &ResolveHandler{service: service}
}

func (h *ResolveHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	url, exists := h.service.Resolve(r.Context(), code)
	if !exists {
		respondWithError(w, http.StatusNotFound, "Short code not found")
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
