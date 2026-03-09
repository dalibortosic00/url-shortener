//go:build e2e

package tests

import (
	"testing"

	h "github.com/dalibortosic00/url-shortener/tests/helpers"
)

func TestGuestShortenDeduplication(t *testing.T) {
	srv := h.NewTestServer(t)
	testUrl := "http://example.com"

	code1 := h.Shorten(t, srv, testUrl)
	code2 := h.Shorten(t, srv, testUrl)

	if code1 != code2 {
		t.Errorf("got different codes for same URL: %q and %q", code1, code2)
	}
}

func TestUserShortenNoDeduplication(t *testing.T) {
	// authenticated users get a unique code each time they shorten the same URL
	srv := h.NewTestServer(t)
	testUrl := "http://example.com"

	apiKey := h.Register(t, srv, "testuser")
	code1 := h.Shorten(t, srv, testUrl, h.WithAPIKey(apiKey))
	code2 := h.Shorten(t, srv, testUrl, h.WithAPIKey(apiKey))

	if code1 == code2 {
		t.Errorf("got same codes for same URL: %q", code1)
	}
}
