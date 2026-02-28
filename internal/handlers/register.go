package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/util"
)

type UserCreator interface {
	Create(ctx context.Context, name string) (string, error)
}

type RegisterHandler struct {
	service UserCreator
}

func NewRegisterHandler(service UserCreator) *RegisterHandler {
	return &RegisterHandler{service: service}
}

type registerRequest struct {
	Name string `json:"name"`
}

func (h *RegisterHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	apiKey, err := h.service.Create(r.Context(), req.Name)
	if err != nil {
		util.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	util.RespondWithJSON(w, http.StatusCreated, map[string]string{"message": "User registered successfully", "api_key": apiKey})
}
