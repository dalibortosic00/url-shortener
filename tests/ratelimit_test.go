//go:build e2e

package tests

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/middleware"
	h "github.com/dalibortosic00/url-shortener/tests/helpers"
	"github.com/stretchr/testify/assert"
)

// realRateLimitConfig has realistic limits for testing rate limiting behavior.
var realRateLimitConfig = middleware.RateLimitConfig{
	AnonLimit:  5,
	AnonWindow: 24 * time.Hour,
	AuthLimit:  5,
	AuthWindow: 1 * time.Hour,
}

// TestRateLimitHitsLimit verifies that both anonymous and authenticated users hit their rate limits.
// Tests that limits are correctly enforced and Retry-After headers are set appropriately.
func TestRateLimitHitsLimit(t *testing.T) {
	tests := []struct {
		name   string
		setup  func(t *testing.T, srv *httptest.Server) func() *http.Response
		limit  int
		window time.Duration
	}{
		{
			name: "anonymous",
			setup: func(t *testing.T, srv *httptest.Server) func() *http.Response {
				params := h.ShortenParams{URL: "http://example.com"}
				return func() *http.Response {
					return h.ShortenRaw(t, srv, params)
				}
			},
			limit:  realRateLimitConfig.AnonLimit,
			window: realRateLimitConfig.AnonWindow,
		},
		{
			name: "authenticated",
			setup: func(t *testing.T, srv *httptest.Server) func() *http.Response {
				apiKey := h.Register(t, srv, "testuser")
				params := h.ShortenParams{URL: "http://example.com"}
				return func() *http.Response {
					return h.ShortenRaw(t, srv, params, h.WithAPIKey(apiKey))
				}
			},
			limit:  realRateLimitConfig.AuthLimit,
			window: realRateLimitConfig.AuthWindow,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := h.NewTestServer(t,
				h.WithRateLimitConfig(realRateLimitConfig),
				h.WithPolicyResolver(h.UniqueIPResolver(t.Name(), realRateLimitConfig)),
			)
			makeRequest := tt.setup(t, srv)

			for i := range tt.limit {
				res := makeRequest()
				res.Body.Close()
				assert.Equal(t, http.StatusOK, res.StatusCode, "request %d should succeed", i+1)
			}

			res := makeRequest()
			defer res.Body.Close()
			assert.Equal(t, http.StatusTooManyRequests, res.StatusCode)

			retryAfter, err := strconv.Atoi(res.Header.Get("Retry-After"))
			assert.NoError(t, err)
			assert.Equal(t, int(tt.window.Seconds()), retryAfter)
		})
	}
}

func TestRateLimitResponseFormat(t *testing.T) {
	srv := h.NewTestServer(t,
		h.WithRateLimitConfig(realRateLimitConfig),
		h.WithPolicyResolver(h.UniqueIPResolver(t.Name(), realRateLimitConfig)),
	)
	params := h.ShortenParams{URL: "http://example.com"}

	// Make 4 successful requests (leave room for another without hitting limit)
	for range 4 {
		res := h.ShortenRaw(t, srv, params)
		res.Body.Close()
	}

	// Make 1 successful request to set up for format test
	res := h.ShortenRaw(t, srv, params)
	res.Body.Close()

	// Now we're at the limit. Next request will be rate limited.
	res = h.ShortenRaw(t, srv, params)
	defer res.Body.Close()

	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode)
	assert.Equal(t, "application/json", res.Header.Get("Content-Type"))

	var errResp struct {
		Error string `json:"error"`
	}
	err := json.NewDecoder(res.Body).Decode(&errResp)
	assert.NoError(t, err)
	assert.NotEmpty(t, errResp.Error)
	assert.Contains(t, strings.ToLower(errResp.Error), "rate limit")
}

func TestRateLimitAuthenticatedDifferentUsers(t *testing.T) {
	srv := h.NewTestServer(t,
		h.WithRateLimitConfig(realRateLimitConfig),
		h.WithPolicyResolver(h.UniqueIPResolver(t.Name(), realRateLimitConfig)),
	)
	user1Key := h.Register(t, srv, "user1")
	user2Key := h.Register(t, srv, "user2")
	params := h.ShortenParams{URL: "http://example.com"}

	// User 1 makes 5 requests (hits limit)
	for i := range 5 {
		res := h.ShortenRaw(t, srv, params, h.WithAPIKey(user1Key))
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode, "user1 request %d should succeed", i+1)
	}

	// User 1 is now rate limited
	res := h.ShortenRaw(t, srv, params, h.WithAPIKey(user1Key))
	defer res.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode, "user1 should be rate limited")

	// User 2 can still make requests (independent limit)
	for i := range 5 {
		res := h.ShortenRaw(t, srv, params, h.WithAPIKey(user2Key))
		defer res.Body.Close()
		assert.Equal(t, http.StatusOK, res.StatusCode, "user2 request %d should succeed", i+1)
	}

	// User 2 is also rate limited after 5 requests
	res = h.ShortenRaw(t, srv, params, h.WithAPIKey(user2Key))
	defer res.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode, "user2 should be rate limited")
}

func TestRateLimitMixedEndpointsShareLimit(t *testing.T) {
	srv := h.NewTestServer(t,
		h.WithRateLimitConfig(realRateLimitConfig),
		h.WithPolicyResolver(h.UniqueIPResolver(t.Name(), realRateLimitConfig)),
	)
	apiKey := h.Register(t, srv, "mixeduser")
	params := h.ShortenParams{URL: "http://example.com"}

	// Use up all 5 requests across different endpoints
	// Request 1: shorten
	res := h.ShortenRaw(t, srv, params, h.WithAPIKey(apiKey))
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Request 2: /links
	res = h.LinksRaw(t, srv, apiKey)
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Request 3: shorten
	res = h.ShortenRaw(t, srv, params, h.WithAPIKey(apiKey))
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Request 4: /links
	res = h.LinksRaw(t, srv, apiKey)
	res.Body.Close()
	assert.Equal(t, http.StatusOK, res.StatusCode)

	// Request 5: shorten and get code
	code := h.Shorten(t, srv, params, h.WithAPIKey(apiKey))

	// Request 6: any endpoint should be rate limited
	res = h.ShortenRaw(t, srv, params, h.WithAPIKey(apiKey))
	defer res.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode, "shorten should be rate limited")

	// Verify /links is also rate limited
	res = h.LinksRaw(t, srv, apiKey)
	defer res.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode, "/links should be rate limited")

	// Verify DELETE is also rate limited
	res = h.DeleteRaw(t, srv, code, apiKey)
	defer res.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, res.StatusCode, "DELETE should be rate limited")
}
