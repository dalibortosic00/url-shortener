package helpers

import (
	"net/http"

	"github.com/dalibortosic00/url-shortener/internal/middleware"
	"github.com/dalibortosic00/url-shortener/internal/request"
)

// UniqueIPResolver creates a PolicyResolver that appends a unique test ID to rate limit keys.
// This allows multiple tests to run in parallel without interfering with each other's rate limits, while still enforcing limits within each test.
func UniqueIPResolver(testID string, config middleware.RateLimitConfig) middleware.PolicyResolver {
	return func(r *http.Request) middleware.LimitPolicy {
		// Authenticated users get a unique key per test and user
		if userID := request.UserID(r.Context()); userID != nil {
			return middleware.LimitPolicy{
				Key:    "user:" + *userID + ":" + testID,
				Limit:  config.AuthLimit,
				Window: config.AuthWindow,
			}
		}

		// Anonymous users get a unique key per test (not shared across tests)
		return middleware.LimitPolicy{
			Key:    "ip:" + testID,
			Limit:  config.AnonLimit,
			Window: config.AnonWindow,
		}
	}
}
