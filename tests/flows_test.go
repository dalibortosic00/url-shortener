//go:build e2e

package tests

import (
	"testing"

	h "github.com/dalibortosic00/url-shortener/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGuestFlow(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL: "http://example.com",
	}

	code := h.Shorten(t, srv, params)

	location := h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)
}

func TestUserFlow(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL: "http://example.com",
	}

	apiKey := h.Register(t, srv, "testuser")
	code := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	location := h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)
}

func TestUserCustomFlow(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL:        "http://example.com",
		CustomCode: "mylink",
	}

	apiKey := h.Register(t, srv, "testuser")
	code := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	location := h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)
}
