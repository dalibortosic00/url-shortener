//go:build e2e

package tests

import (
	"testing"

	h "github.com/dalibortosic00/url-shortener/tests/helpers"
)

func TestGuestFlow(t *testing.T) {
	srv := h.NewTestServer(t)
	testUrl := "http://example.com"

	code := h.Shorten(t, srv, testUrl)

	location := h.Resolve(t, srv, code)
	if location != testUrl {
		t.Errorf("got %q, want %q", location, testUrl)
	}
}

func TestUserFlow(t *testing.T) {
	srv := h.NewTestServer(t)
	testUrl := "http://example.com"

	apiKey := h.Register(t, srv, "testuser")
	code := h.Shorten(t, srv, testUrl, h.WithAPIKey(apiKey))

	location := h.Resolve(t, srv, code)
	if location != testUrl {
		t.Errorf("got %q, want %q", location, testUrl)
	}
}
