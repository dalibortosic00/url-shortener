package handlers

import (
	"context"
	"errors"
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/dalibortosic00/url-shortener/internal/util"
	"github.com/go-chi/chi/v5"
)

type LinkDeleter interface {
	Delete(ctx context.Context, code string, ownerID string) error
}

type DeleteHandler struct {
	service LinkDeleter
}

func NewDeleteHandler(service LinkDeleter) *DeleteHandler {
	return &DeleteHandler{service: service}
}

func (h *DeleteHandler) Delete(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	userID := request.UserID(r.Context())

	if userID == nil {
		util.RespondWithError(w, http.StatusUnauthorized, "Authentication required")
		return
	}

	err := h.service.Delete(r.Context(), code, *userID)
	if err != nil {
		if errors.Is(err, models.ErrRecordNotFound) {
			util.RespondWithError(w, http.StatusNotFound, "Short code not found")
			return
		}

		util.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
