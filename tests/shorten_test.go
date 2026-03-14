//go:build e2e

package tests

import (
	"net/http"
	"testing"

	h "github.com/dalibortosic00/url-shortener/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGuestShortenDeduplication(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL: "http://example.com",
	}

	code1 := h.Shorten(t, srv, params)
	code2 := h.Shorten(t, srv, params)

	assert.Equal(t, code1, code2)
}

func TestUserShortenNoDeduplication(t *testing.T) {
	// authenticated users get a unique code each time they shorten the same URL
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL: "http://example.com",
	}

	apiKey := h.Register(t, srv, "testuser")
	code1 := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))
	code2 := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	assert.NotEqual(t, code1, code2)
}

func TestUserShortenCustomDeduplication(t *testing.T) {
	srv := h.NewTestServer(t)

	apiKey := h.Register(t, srv, "testuser")
	params := h.ShortenParams{
		URL:        "http://example.com",
		CustomCode: "mylink",
	}

	h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	res := h.ShortenRaw(t, srv, params, h.WithAPIKey(apiKey))
	assert.Equal(t, http.StatusConflict, res.StatusCode)
}
