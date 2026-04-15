package middleware

import (
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dalibortosic00/url-shortener/internal/request"
	"github.com/dalibortosic00/url-shortener/internal/util"
)

// RateLimitMiddleware returns a chi middleware that enforces rate limits.
// It uses a PolicyResolver to determine limits based on request characteristics.
// This separates identity/policy resolution from rate limit enforcement,
// following Single Responsibility Principle (SRP) and Open/Closed Principle (OCP).
// New policies can be added by creating new PolicyResolver implementations
// without modifying the middleware.
func RateLimitMiddleware(rl *RateLimiter, resolver PolicyResolver) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			policy := resolver(r)

			if !rl.IsAllowed(policy.Key, policy.Limit, policy.Window) {
				w.Header().Set("Retry-After", strconv.Itoa(int(policy.Window.Seconds())))
				util.RespondWithError(w, http.StatusTooManyRequests, "rate limit exceeded")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// DefaultPolicyResolver creates a policy resolver with standard auth/anonymous limits.
// Separates identity resolution from rate limit enforcement (SRP).
func DefaultPolicyResolver(config RateLimitConfig) PolicyResolver {
	return func(r *http.Request) LimitPolicy {
		if userID := request.UserID(r.Context()); userID != nil {
			// Authenticated: limit per user ID
			return LimitPolicy{
				Key:    userKeyPrefix + *userID,
				Limit:  config.AuthLimit,
				Window: config.AuthWindow,
			}
		}

		// Anonymous: limit per IP
		ip := extractIP(r.RemoteAddr)
		return LimitPolicy{
			Key:    ipKeyPrefix + ip,
			Limit:  config.AnonLimit,
			Window: config.AnonWindow,
		}
	}
}

func NewRateLimiter() *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		stopCh:   make(chan struct{}),
	}
	go rl.cleanupLoop()
	return rl
}

type RateLimitConfig struct {
	AnonLimit  int           // Max requests for anonymous users
	AnonWindow time.Duration // Time window for anonymous users
	AuthLimit  int           // Max requests for authenticated users
	AuthWindow time.Duration // Time window for authenticated users
}

// PolicyResolver determines rate limit policy for a request.
// It encapsulates the logic of extracting identity and selecting appropriate limits.
// This separation allows new policies to be added without modifying the middleware.
type PolicyResolver func(r *http.Request) LimitPolicy

// LimitPolicy defines the rate limit parameters for a request.
type LimitPolicy struct {
	Key    string        // Unique identifier for rate limiting (user ID or IP)
	Limit  int           // Maximum requests allowed
	Window time.Duration // Time window for the limit
}

// RateLimiter implements sliding window rate limiting in memory.
// It is safe for concurrent use. Use Stop() to clean up resources.
type RateLimiter struct {
	mu        sync.Mutex
	requests  map[string][]time.Time
	stopCh    chan struct{}
	maxWindow time.Duration
}

// IsAllowed checks if a request is allowed for the given key within the window.
// If allowed, it records the request and returns true. Otherwise, returns false.
// This method is safe for concurrent use.
func (rl *RateLimiter) IsAllowed(key string, limit int, window time.Duration) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-window)

	// Clean up old requests outside the window
	requests := rl.requests[key]
	validRequests := make([]time.Time, 0, len(requests))
	for _, req := range requests {
		if req.After(cutoff) {
			validRequests = append(validRequests, req)
		}
	}

	if len(validRequests) >= limit {
		rl.requests[key] = validRequests
		return false
	}

	// Record this request
	validRequests = append(validRequests, now)
	rl.requests[key] = validRequests
	return true
}

// Stop stops the rate limiter's cleanup goroutine and releases resources.
func (rl *RateLimiter) Stop() {
	close(rl.stopCh)
}

const (
	userKeyPrefix = "user:"
	ipKeyPrefix   = "ip:"
)

// cleanupLoop periodically removes empty key entries from the requests map.
// Runs in a background goroutine until Stop() is called.
func (rl *RateLimiter) cleanupLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-rl.stopCh:
			return
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for key, reqs := range rl.requests {
				valid := reqs[:0]
				for _, req := range reqs {
					if req.After(now.Add(-7 * 24 * time.Hour)) {
						valid = append(valid, req)
					}
				}
				if len(valid) == 0 {
					delete(rl.requests, key)
				} else {
					rl.requests[key] = valid
				}
			}
			rl.mu.Unlock()
		}
	}
}

func extractIP(remoteAddr string) string {
	ip, _, err := net.SplitHostPort(remoteAddr)
	if err != nil {
		return remoteAddr
	}
	return ip
}
