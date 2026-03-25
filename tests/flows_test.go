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

func TestUserListFlow(t *testing.T) {
	srv := h.NewTestServer(t)

	apiKey := h.Register(t, srv, "testuser")
	urls := make([]string, 3)
	for i := range 3 {
		urls[i] = "http://example.com/" + string(rune('a'+i))
		h.Shorten(t, srv, h.ShortenParams{URL: urls[i]}, h.WithAPIKey(apiKey))
	}

	links := h.Links(t, srv, apiKey)
	assert.Len(t, links, 3)
	res := make([]string, 0, len(links))
	for _, url := range links {
		res = append(res, url)
	}
	assert.ElementsMatch(t, urls, res)
}
