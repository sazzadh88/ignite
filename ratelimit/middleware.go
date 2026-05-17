package ratelimit

import (
	"fmt"
	"net/http"
	"strconv"
)

// Throttle returns middleware that applies a named rate limiter.
// The limiter must be defined using RateLimiter.For() before use.
func Throttle(limiterName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limiterFn, exists := RateLimiter.GetNamedLimiter(limiterName)
			if !exists {
				// Named limiter not found, allow request
				next.ServeHTTP(w, r)
				return
			}

			limit := limiterFn(r)
			if limit == nil {
				// No limit configured
				next.ServeHTTP(w, r)
				return
			}

			// Determine the key
			key := limit.Key
			if key == "" {
				// Default to IP address
				key = getClientIP(r)
			}

			// Check if unlimited
			if limit.MaxAttempts < 0 {
				next.ServeHTTP(w, r)
				return
			}

			// Check rate limit
			if RateLimiter.TooManyAttempts(key, limit.MaxAttempts) {
				sendRateLimitResponse(w, r, key, limit.MaxAttempts, limit.DecayMinutes)
				return
			}

			// Increment attempt count
			RateLimiter.Hit(key, limit.DecayMinutes)

			// Set rate limit headers
			setRateLimitHeaders(w, key, limit.MaxAttempts)

			next.ServeHTTP(w, r)
		})
	}
}

// ThrottleWithLimit returns middleware that applies a simple rate limit
// with the specified maxAttempts and decayMinutes.
func ThrottleWithLimit(maxAttempts, decayMinutes int) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if unlimited
			if maxAttempts < 0 {
				next.ServeHTTP(w, r)
				return
			}

			key := getClientIP(r)

			// Check rate limit
			if RateLimiter.TooManyAttempts(key, maxAttempts) {
				sendRateLimitResponse(w, r, key, maxAttempts, decayMinutes)
				return
			}

			// Increment attempt count
			RateLimiter.Hit(key, decayMinutes)

			// Set rate limit headers
			setRateLimitHeaders(w, key, maxAttempts)

			next.ServeHTTP(w, r)
		})
	}
}

// getClientIP extracts the client IP address from the request.
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For header first
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}

	// Check X-Real-IP header
	realIP := r.Header.Get("X-Real-IP")
	if realIP != "" {
		return realIP
	}

	// Fall back to RemoteAddr
	return r.RemoteAddr
}

// setRateLimitHeaders sets the rate limit response headers.
func setRateLimitHeaders(w http.ResponseWriter, key string, maxAttempts int) {
	remaining := RateLimiter.Remaining(key, maxAttempts)
	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxAttempts))
	w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
}

// sendRateLimitResponse sends a 429 Too Many Requests response.
func sendRateLimitResponse(w http.ResponseWriter, r *http.Request, key string, maxAttempts, decayMinutes int) {
	retryAfter := RateLimiter.AvailableIn(key)
	retryAfterSeconds := int(retryAfter.Seconds())
	if retryAfterSeconds < 1 {
		retryAfterSeconds = 1
	}

	w.Header().Set("X-RateLimit-Limit", strconv.Itoa(maxAttempts))
	w.Header().Set("X-RateLimit-Remaining", "0")
	w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusTooManyRequests)
	fmt.Fprintf(w, "Too Many Requests. Retry after %d seconds.\n", retryAfterSeconds)
}
