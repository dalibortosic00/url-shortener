package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strings"

	"github.com/dalibortosic00/url-shortener/internal/services"
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

func (r *shortenRequest) Validate(forbiddenHost string) (string, bool) {
	input := strings.TrimSpace(r.URL)
	if input == "" {
		return "URL cannot be empty", false
	}

	if !strings.HasPrefix(input, "http://") && !strings.HasPrefix(input, "https://") {
		input = "https://" + input
	}

	u, err := url.ParseRequestURI(input)
	if err != nil {
		return "The URL format is invalid", false
	}

	host := u.Hostname()
	if host == "" || host == "." {
		return "A valid domain name is required", false
	}

	if host != "localhost" {
		parts := strings.Split(host, ".")
		if len(parts) < 2 || parts[len(parts)-1] == "" {
			return "Domain must have a valid TLD (e.g. .com, .net)", false
		}
	}

	if strings.EqualFold(u.Host, forbiddenHost) {
		return "Cannot shorten URLs from this domain", false
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

	if msg, ok := req.Validate(h.forbiddenHost); !ok {
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
