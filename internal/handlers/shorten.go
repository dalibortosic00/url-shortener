package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dalibortosic00/url-shortener/internal/services"
)

const (
	errEmptyURL        = "URL cannot be empty"
	errInvalidFormat   = "The URL format is invalid"
	errInvalidDomain   = "A valid domain name is required"
	errInvalidTLD      = "Domain must have a valid TLD (e.g. .com, .net)"
	errForbiddenDomain = "Cannot shorten URLs from this domain"
)

type ShortenHandler struct {
	service       *services.ShortenerService
	baseURL       string
	forbiddenHost string
}

func NewShortenHandler(service *services.ShortenerService, baseURL string) *ShortenHandler {
	parsed, _ := url.Parse(baseURL)
	return &ShortenHandler{
		service:       service,
		baseURL:       strings.TrimSuffix(baseURL, "/"),
		forbiddenHost: parsed.Host,
	}
}

type shortenRequest struct {
	URL string `json:"url"`
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
	return "", true
}

func (h *ShortenHandler) Shorten(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 4096)

	var req shortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid JSON payload")
		return
	}

	if msg, ok := req.validate(h.forbiddenHost); !ok {
		respondWithError(w, http.StatusBadRequest, msg)
		return
	}

	code, err := h.service.Create(r.Context(), req.URL)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Internal service error")
		return
	}

	respondWithJSON(w, http.StatusOK, map[string]string{
		"short_url": h.baseURL + "/" + code,
	})
}
