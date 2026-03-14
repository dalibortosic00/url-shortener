package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"unicode"

	"github.com/dalibortosic00/url-shortener/internal/models"
	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/dalibortosic00/url-shortener/internal/util"
)

const (
	errEmptyURL        = "URL cannot be empty"
	errInvalidFormat   = "The URL format is invalid"
	errInvalidDomain   = "A valid domain name is required"
	errInvalidTLD      = "Domain must have a valid TLD (e.g. .com, .net)"
	errForbiddenDomain = "Cannot shorten URLs from this domain"
)

func isAlphanumeric(s string) bool {
	for _, r := range s {
		if !unicode.IsLetter(r) && !unicode.IsNumber(r) {
			return false
		}
	}
	return true
}

type LinkCreator interface {
	Create(ctx context.Context, url string, ownerID string) (string, error)
	CreateCustom(ctx context.Context, url string, customCode string, ownerID string) (string, error)
}

type ShortenHandler struct {
	service       LinkCreator
	baseURL       string
	forbiddenHost string
}

func NewShortenHandler(service LinkCreator, baseURL string) *ShortenHandler {
	parsed, _ := url.Parse(baseURL)
	return &ShortenHandler{
		service:       service,
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		forbiddenHost: parsed.Host,
	}
}

type shortenRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
}

func (r *shortenRequest) validate(forbiddenHost string) (string, bool) {
	input := strings.TrimSpace(r.URL)
	if input == "" {
		return errEmptyURL, false
	}

	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	u, err := url.ParseRequestURI(input)
	if err != nil {
		return errInvalidFormat, false
	}

	host := u.Hostname()
	if host == "" || host == "." {
		return errInvalidDomain, false
	}

	if host != "localhost" {
		parts := strings.Split(host, ".")
		if len(parts) < 2 || parts[0] == "" || parts[len(parts)-1] == "" {
			return errInvalidTLD, false
		}
	}

	if strings.EqualFold(u.Host, forbiddenHost) {
		return errForbiddenDomain, false
	}

	r.URL = input

	if r.CustomCode != "" {
		length := len(r.CustomCode)
		if length < 3 {
			return "Custom code must be at least 3 characters", false
		}
		if length > 12 {
			return "Custom code cannot exceed 12 characters", false
		}
		if !isAlphanumeric(r.CustomCode) {
			return "Custom code must only contain letters and numbers", false
		}
	}

	return "", true
}

func (h *ShortenHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req shortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		util.RespondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if msg, ok := req.validate(h.forbiddenHost); !ok {
		util.RespondWithError(w, http.StatusBadRequest, msg)
		return
	}

	userID := request.UserID(r.Context())
	var code string
	var err error

	if req.CustomCode != "" {
		if userID == "" {
			util.RespondWithError(w, http.StatusForbidden, "Custom codes are only available for registered users")
			return
		}
		code, err = h.service.CreateCustom(r.Context(), req.URL, req.CustomCode, userID)
	} else {
		code, err = h.service.Create(r.Context(), req.URL, userID)
	}

	if err != nil {
		if errors.Is(err, models.ErrCollision) {
			util.RespondWithError(w, http.StatusConflict, "This custom code is already taken")
			return
		}
		util.RespondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	util.RespondWithJSON(w, http.StatusOK, map[string]string{
		"short_url": h.baseURL + "/" + code,
	})
}
