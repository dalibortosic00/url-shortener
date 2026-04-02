//go:build e2e

package tests

import (
	"net/http"
	"testing"

	h "github.com/dalibortosic00/url-shortener/tests/helpers"
	"github.com/stretchr/testify/assert"
)

func TestDeleteFlow_SuccessfulDelete(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL: "http://example.com",
	}

	apiKey := h.Register(t, srv, "testuser")
	code := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	location := h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)

	h.Delete(t, srv, code, apiKey)

	res := h.ResolveRaw(t, srv, code)
	defer res.Body.Close()
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestDeleteFlow_DeleteCustomCode(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL:        "http://example.com",
		CustomCode: "mylink",
	}

	apiKey := h.Register(t, srv, "testuser")
	code := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	location := h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)

	h.Delete(t, srv, code, apiKey)

	res := h.ResolveRaw(t, srv, code)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestDeleteFlow_DeleteNonexistentCode(t *testing.T) {
	srv := h.NewTestServer(t)

	apiKey := h.Register(t, srv, "testuser")

	res := h.DeleteRaw(t, srv, "nonexistent", apiKey)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}

func TestDeleteFlow_CannotDeleteOthersLink(t *testing.T) {
	srv := h.NewTestServer(t)
	params := h.ShortenParams{
		URL: "http://example.com",
	}

	apiKey1 := h.Register(t, srv, "user1")
	apiKey2 := h.Register(t, srv, "user2")

	code := h.Shorten(t, srv, params, h.WithAPIKey(apiKey1))

	location := h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)

	res := h.DeleteRaw(t, srv, code, apiKey2)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)

	location = h.Resolve(t, srv, code)
	assert.Equal(t, params.URL, location)
}
