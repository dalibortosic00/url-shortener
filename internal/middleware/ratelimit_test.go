package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/stretchr/testify/assert"
)

// defaultTestConfig returns a standard test configuration.
func defaultTestConfig() RateLimitConfig {
	return RateLimitConfig{
		AnonLimit:  5,
		AnonWindow: 24 * time.Hour,
		AuthLimit:  5,
		AuthWindow: 1 * time.Hour,
	}
}

// newTestHandler returns a basic HTTP handler for testing.
func newTestHandler() http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
}

// setupTestRequest creates a request with optional authentication.
func setupTestRequest(method, path, ip string, userID *string) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	req.RemoteAddr = ip
	if userID != nil {
		ctx := request.WithUserID(req.Context(), *userID)
		req = req.WithContext(ctx)
	}
	return req
}

func TestRateLimiter_IsAllowed(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		limit    int
		expected []bool
	}{
		{
			name:     "allows requests within limit",
			key:      "test-key",
			limit:    3,
			expected: []bool{true, true, true},
		},
		{
			name:     "blocks requests exceeding limit",
			key:      "test-key-2",
			limit:    2,
			expected: []bool{true, true, false},
		},
		{
			name:     "handles zero limit",
			key:      "zero-limit",
			limit:    0,
			expected: []bool{false},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rl := NewRateLimiter()
			defer rl.Stop()

			window := 1 * time.Hour
			results := make([]bool, len(tt.expected))

			for i := range tt.expected {
				results[i] = rl.IsAllowed(tt.key, tt.limit, window)
			}

			assert.Equal(t, tt.expected, results, "%s: expected exact sequence of results", tt.name)
		})
	}
}

func TestRateLimiter_DifferentKeysTrackedSeparately(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	const limit = 2
	window := 1 * time.Hour

	// Fill key1 limit
	assert.True(t, rl.IsAllowed("key1", limit, window))
	assert.True(t, rl.IsAllowed("key1", limit, window))
	assert.False(t, rl.IsAllowed("key1", limit, window))

	// key2 should have its own limit
	assert.True(t, rl.IsAllowed("key2", limit, window))
	assert.True(t, rl.IsAllowed("key2", limit, window))
	assert.False(t, rl.IsAllowed("key2", limit, window))
}

func TestRateLimiter_ResetsAfterWindowExpires(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()

	const key = "reset-test"
	const limit = 1
	window := 100 * time.Millisecond

	// Use the single request in the window
	assert.True(t, rl.IsAllowed(key, limit, window))

	// Should be blocked before window expires
	assert.False(t, rl.IsAllowed(key, limit, window))

	// Wait for window to expire
	time.Sleep(110 * time.Millisecond)

	// Should be allowed again after window expires
	assert.True(t, rl.IsAllowed(key, limit, window))
}

func TestDefaultPolicyResolver(t *testing.T) {
	config := defaultTestConfig()
	resolver := DefaultPolicyResolver(config)

	t.Run("anonymous user policy uses IP key and anon limits", func(t *testing.T) {
		req := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)
		policy := resolver(req)

		assert.Equal(t, ipKeyPrefix+"192.0.2.1", policy.Key)
		assert.Equal(t, config.AnonLimit, policy.Limit)
		assert.Equal(t, config.AnonWindow, policy.Window)
	})

	t.Run("authenticated user policy uses user key and auth limits", func(t *testing.T) {
		userID := "user-123"
		req := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID)
		policy := resolver(req)

		assert.Equal(t, userKeyPrefix+userID, policy.Key)
		assert.Equal(t, config.AuthLimit, policy.Limit)
		assert.Equal(t, config.AuthWindow, policy.Window)
	})

	t.Run("different IPs get different keys", func(t *testing.T) {
		req1 := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)
		req2 := setupTestRequest("POST", "/shorten", "192.0.2.2:8080", nil)

		policy1 := resolver(req1)
		policy2 := resolver(req2)

		assert.NotEqual(t, policy1.Key, policy2.Key)
		assert.Equal(t, ipKeyPrefix+"192.0.2.1", policy1.Key)
		assert.Equal(t, ipKeyPrefix+"192.0.2.2", policy2.Key)
	})

	t.Run("different users get different keys", func(t *testing.T) {
		userID1 := "user-1"
		userID2 := "user-2"
		req1 := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID1)
		req2 := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID2)

		policy1 := resolver(req1)
		policy2 := resolver(req2)

		assert.NotEqual(t, policy1.Key, policy2.Key)
		assert.Equal(t, userKeyPrefix+userID1, policy1.Key)
		assert.Equal(t, userKeyPrefix+userID2, policy2.Key)
	})
}

func TestRateLimitMiddleware_Anonymous(t *testing.T) {
	testHandler := newTestHandler()

	t.Run("allows anonymous users within limit", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := defaultTestConfig()
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		req := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)

		// First 5 requests should succeed
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// 6th request should be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
		assert.Contains(t, rec.Body.String(), "rate limit exceeded")
	})

	t.Run("different IPs are tracked separately", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := defaultTestConfig()
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		req1 := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)
		req2 := setupTestRequest("POST", "/shorten", "192.0.2.2:8080", nil)

		// First IP can make 5 requests
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req1)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Second IP should also be able to make 5 requests
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req2)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// First IP should now be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req1)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)

		// Second IP should also be rate limited
		rec = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req2)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})
}

func TestRateLimitMiddleware_Authenticated(t *testing.T) {
	testHandler := newTestHandler()

	t.Run("allows authenticated users within limit", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := defaultTestConfig()
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		userID := "user-123"
		req := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID)

		// First 5 requests should succeed
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// 6th request should be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})

	t.Run("different users are tracked separately", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := defaultTestConfig()
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		userID1 := "user-1"
		userID2 := "user-2"
		req1 := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID1)
		req2 := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID2)

		// First user can make 5 requests
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req1)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Second user should also be able to make 5 requests (different user key)
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req2)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// First user should now be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req1)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)

		// Second user should also be rate limited
		rec = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req2)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})
}

func TestRateLimitMiddleware_AuthenticatedVsAnonymous(t *testing.T) {
	rl := NewRateLimiter()
	defer rl.Stop()
	config := defaultTestConfig()
	middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
	testHandler := newTestHandler()
	wrappedHandler := middleware(testHandler)

	// Anonymous user (no UserID in context)
	anonReq := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)

	// Authenticated user
	userID := "user-123"
	authReq := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID)

	// Make 5 requests as anonymous
	for range 5 {
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, anonReq)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// Anonymous should be blocked
	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, anonReq)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// But authenticated user should still be able to make requests
	for range 5 {
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, authReq)
		assert.Equal(t, http.StatusOK, rec.Code)
	}
}

func TestRateLimitMiddleware_CustomLimits(t *testing.T) {
	testHandler := newTestHandler()

	t.Run("respects custom anonymous limits", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := RateLimitConfig{
			AnonLimit:  2,
			AnonWindow: 24 * time.Hour,
			AuthLimit:  5,
			AuthWindow: 1 * time.Hour,
		}
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		req := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)

		// First 2 should succeed
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		rec = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)

		// 3rd should be rate limited
		rec = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})

	t.Run("respects custom authenticated limits", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := RateLimitConfig{
			AnonLimit:  5,
			AnonWindow: 24 * time.Hour,
			AuthLimit:  3,
			AuthWindow: 1 * time.Hour,
		}
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		userID := "user-123"
		req := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID)

		// First 3 should succeed
		for range 3 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// 4th should be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	})
}

func TestRateLimitMiddleware_RetryAfterHeader(t *testing.T) {
	testHandler := newTestHandler()

	t.Run("anonymous user rate limit includes Retry-After header", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := defaultTestConfig()
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		req := setupTestRequest("POST", "/shorten", "192.0.2.1:8080", nil)

		// Exhaust anonymous limit (5 per 24h)
		for range 5 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// 6th request should be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)

		// Check Retry-After header is set (should be 86400 seconds for 24h)
		retryAfter := rec.Header().Get("Retry-After")
		assert.Equal(t, "86400", retryAfter)
	})

	t.Run("authenticated user rate limit includes Retry-After header", func(t *testing.T) {
		rl := NewRateLimiter()
		defer rl.Stop()
		config := RateLimitConfig{
			AnonLimit:  5,
			AnonWindow: 24 * time.Hour,
			AuthLimit:  2,
			AuthWindow: 1 * time.Hour,
		}
		middleware := RateLimitMiddleware(rl, DefaultPolicyResolver(config))
		wrappedHandler := middleware(testHandler)

		userID := "user-123"
		req := setupTestRequest("GET", "/links", "192.0.2.1:8080", &userID)

		// Exhaust authenticated limit (2 per 1h)
		for range 2 {
			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// 3rd request should be rate limited
		rec := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusTooManyRequests, rec.Code)

		// Check Retry-After header is set (should be 3600 seconds for 1h)
		retryAfter := rec.Header().Get("Retry-After")
		assert.Equal(t, "3600", retryAfter)
	})
}
