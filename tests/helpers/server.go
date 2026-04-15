package helpers

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/config"
	"github.com/dalibortosic00/url-shortener/internal/generators"
	"github.com/dalibortosic00/url-shortener/internal/middleware"
	"github.com/dalibortosic00/url-shortener/internal/server"
	"github.com/dalibortosic00/url-shortener/internal/services"
	"github.com/dalibortosic00/url-shortener/internal/store"
)

type TestServerOption func(*testServerConfig)

type testServerConfig struct {
	rateLimitConfig middleware.RateLimitConfig
	policyResolver  middleware.PolicyResolver
}

// noopRateLimitConfig has very high limits so rate limiting is effectively disabled.
// Used by default for non-rate-limit tests.
var noopRateLimitConfig = middleware.RateLimitConfig{
	AnonLimit:  99999,
	AnonWindow: time.Second,
	AuthLimit:  99999,
	AuthWindow: time.Second,
}

// WithRateLimitConfig sets the rate limit configuration for the test server.
// Useful for overriding the default high-limit config.
func WithRateLimitConfig(cfg middleware.RateLimitConfig) TestServerOption {
	return func(c *testServerConfig) {
		c.rateLimitConfig = cfg
	}
}

// WithPolicyResolver sets a custom policy resolver for rate limiting.
// Useful for testing rate limiting with test-isolated keys.
func WithPolicyResolver(resolver middleware.PolicyResolver) TestServerOption {
	return func(c *testServerConfig) {
		c.policyResolver = resolver
	}
}

func NewTestServer(t *testing.T, opts ...TestServerOption) *httptest.Server {
	t.Helper()

	cfg := config.Load()
	if cfg.TestDatabaseURL == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping e2e tests")
	}

	testCfg := &testServerConfig{
		rateLimitConfig: noopRateLimitConfig,
		policyResolver:  nil, // Will use default if not overridden
	}
	for _, opt := range opts {
		opt(testCfg)
	}

	db := NewTestDB(t, cfg.TestDatabaseURL)
	gen := generators.NewRandomGenerator()

	store := store.NewDatabaseStore(db)

	userService := services.NewUserService(store, gen)
	linkService := services.NewLinkService(store, gen)
	authMiddleware := middleware.NewAuthMiddleware(userService)
	rateLimiter := middleware.NewRateLimiter()

	rateLimitOpts := &server.RateLimitOptions{
		Limiter:  rateLimiter,
		Config:   testCfg.rateLimitConfig,
		Resolver: testCfg.policyResolver,
	}
	srv := server.New(cfg, userService, linkService, authMiddleware, nil, rateLimitOpts)

	ts := httptest.NewServer(srv.Handler)
	t.Cleanup(func() {
		Reset(t, db)
		ts.Close()
		rateLimiter.Stop()
	})

	return ts
}
