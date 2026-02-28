package handlers

import (
	"context"
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/util"
	"github.com/go-chi/chi/v5"
)

type LinkResolver interface {
	Resolve(ctx context.Context, code string) (string, bool)
}

type ResolveHandler struct {
	service LinkResolver
}

func NewResolveHandler(service LinkResolver) *ResolveHandler {
	return &ResolveHandler{service: service}
}

func (h *ResolveHandler) Resolve(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")

	url, exists := h.service.Resolve(r.Context(), code)
	if !exists {
		util.RespondWithError(w, http.StatusNotFound, "Short code not found")
		return
	}

	http.Redirect(w, r, url, http.StatusFound)
}
