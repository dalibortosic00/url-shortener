package handlers

import (
	"context"
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/dalibortosic00/url-shortener/internal/util"
)

type LinkLister interface {
	List(ctx context.Context, ownerID string) ([]models.LinkRecord, error)
}

type LinksHandler struct {
	service LinkLister
	baseURL string
}

func NewLinksHandler(service LinkLister, baseURL string) *LinksHandler {
	return &LinksHandler{
		service: service,
		baseURL: baseURL,
	}
}

func (h *LinksHandler) List(w http.ResponseWriter, r *http.Request) {
	userID := request.UserID(r.Context())

	links, err := h.service.List(r.Context(), *userID)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	output := make(map[string]string)
	for _, link := range links {
		output[h.baseURL+"/"+link.Code] = link.URL
	}

	util.RespondWithJSON(w, http.StatusOK, output)
}
