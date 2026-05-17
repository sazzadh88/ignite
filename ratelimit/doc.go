// Package ratelimit provides rate limiting functionality for HTTP applications.
//
// The package offers flexible rate limiting with support for named limiters,
// custom keys, and both in-memory and custom storage backends. It includes
// HTTP middleware for easy integration with web applications.
//
// # Basic Usage
//
// The simplest way to use the rate limiter is with the Attempt method:
//
//	limiter := ratelimit.NewLimiter()
//	key := "user:123"
//
//	if limiter.Attempt(key, 10, 1) {
//	    // Allow the request (within limit)
//	} else {
//	    // Reject the request (limit exceeded)
//	}
//
// # Middleware
//
// The package provides HTTP middleware for automatic rate limiting:
//
//	// Simple rate limit: 60 requests per minute
//	handler := ratelimit.ThrottleWithLimit(60, 1)(yourHandler)
//
//	// Named limiter with custom logic
//	ratelimit.RateLimiter.For("api", func(r *http.Request) *ratelimit.Limit {
//	    if isAuthenticated(r) {
//	        return ratelimit.PerHour(1000)
//	    }
//	    return ratelimit.PerHour(100)
//	})
//	handler := ratelimit.Throttle("api")(yourHandler)
//
// # Limit Helpers
//
// The package provides convenient helpers for common time periods:
//
//	limit := ratelimit.PerMinute(60)   // 60 requests per minute
//	limit := ratelimit.PerHour(1000)   // 1000 requests per hour
//	limit := ratelimit.PerDay(10000)   // 10000 requests per day
//	limit := ratelimit.None()          // Unlimited requests
//
// # Custom Keys
//
// By default, the middleware uses the client IP address as the rate limit key.
// You can specify custom keys using the By method:
//
//	limit := ratelimit.PerMinute(60).By("api:" + apiKey)
//
// # Manual Operations
//
// The Limiter type provides several methods for manual rate limit management:
//
//	limiter.Hit(key, decayMinutes)           // Increment attempts
//	limiter.TooManyAttempts(key, max)        // Check if limit exceeded
//	limiter.Remaining(key, max)              // Get remaining attempts
//	limiter.Clear(key)                       // Reset attempts
//	limiter.RetriesIn(key)                   // Time until reset
//	limiter.AvailableIn(key)                 // Alias for RetriesIn
//
// # Custom Storage
//
// The package uses in-memory storage by default, but you can provide a custom
// storage backend by implementing the Store interface:
//
//	type Store interface {
//	    Get(key string) (int, time.Time, bool)
//	    Increment(key string, decay time.Duration) int
//	    Reset(key string)
//	    Clean()
//	}
//
//	limiter := ratelimit.NewLimiterWithStore(myCustomStore)
//
// # Response Headers
//
// When using the middleware, the following headers are set:
//
//   - X-RateLimit-Limit: Maximum number of requests allowed
//   - X-RateLimit-Remaining: Number of requests remaining
//   - Retry-After: Seconds until the limit resets (only on 429 responses)
//
// # Zero Dependencies
//
// This package has zero external dependencies and only uses the Go standard library.
package ratelimit
