package ratelimit_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/sazzad/ignite/ratelimit"
)

// Example_basicUsage demonstrates basic rate limiting with Attempt.
func Example_basicUsage() {
	limiter := ratelimit.NewLimiter()
	key := "user:123"

	// Try 3 attempts with a limit of 2
	for i := 1; i <= 3; i++ {
		allowed := limiter.Attempt(key, 2, 1)
		fmt.Printf("Attempt %d: %v\n", i, allowed)
	}

	// Output:
	// Attempt 1: true
	// Attempt 2: true
	// Attempt 3: false
}

// Example_middleware demonstrates using the middleware.
func Example_middleware() {
	// Create a simple handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Success"))
	})

	// Apply throttle middleware (5 requests per minute)
	throttled := ratelimit.ThrottleWithLimit(5, 1)(handler)

	// Make a request
	req := httptest.NewRequest("GET", "/api/data", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()

	throttled.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Limit:", w.Header().Get("X-RateLimit-Limit"))
	fmt.Println("Remaining:", w.Header().Get("X-RateLimit-Remaining"))

	// Output:
	// Status: 200
	// Limit: 5
	// Remaining: 4
}

// Example_namedLimiter demonstrates using named limiters.
func Example_namedLimiter() {
	// Define a named limiter
	ratelimit.RateLimiter.For("api", func(r *http.Request) *ratelimit.Limit {
		// Different limits based on authentication
		if r.Header.Get("Authorization") != "" {
			return ratelimit.PerHour(1000) // Authenticated users get more
		}
		return ratelimit.PerHour(100) // Anonymous users
	})

	// Create handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("API Response"))
	})

	// Apply named limiter
	throttled := ratelimit.Throttle("api")(handler)

	// Make authenticated request
	req := httptest.NewRequest("GET", "/api/endpoint", nil)
	req.Header.Set("Authorization", "Bearer token123")
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()

	throttled.ServeHTTP(w, req)

	fmt.Println("Status:", w.Code)
	fmt.Println("Has Rate Limit Headers:", w.Header().Get("X-RateLimit-Limit") != "")

	// Output:
	// Status: 200
	// Has Rate Limit Headers: true
}

// Example_perMinute demonstrates PerMinute limit helper.
func Example_perMinute() {
	limit := ratelimit.PerMinute(60)
	fmt.Printf("MaxAttempts: %d, DecayMinutes: %d\n", limit.MaxAttempts, limit.DecayMinutes)

	// Output:
	// MaxAttempts: 60, DecayMinutes: 1
}

// Example_perHour demonstrates PerHour limit helper.
func Example_perHour() {
	limit := ratelimit.PerHour(1000)
	fmt.Printf("MaxAttempts: %d, DecayMinutes: %d\n", limit.MaxAttempts, limit.DecayMinutes)

	// Output:
	// MaxAttempts: 1000, DecayMinutes: 60
}

// Example_customKey demonstrates using custom keys with By().
func Example_customKey() {
	limit := ratelimit.PerMinute(100).By("api:key123")
	fmt.Printf("Key: %s, MaxAttempts: %d\n", limit.Key, limit.MaxAttempts)

	// Output:
	// Key: api:key123, MaxAttempts: 100
}

// Example_unlimited demonstrates unlimited rate limiting.
func Example_unlimited() {
	limit := ratelimit.None()
	limiter := ratelimit.NewLimiter()

	// Should allow unlimited attempts
	for i := 1; i <= 5; i++ {
		allowed := limiter.Attempt("key", limit.MaxAttempts, limit.DecayMinutes)
		fmt.Printf("Attempt %d: %v\n", i, allowed)
	}

	// Output:
	// Attempt 1: true
	// Attempt 2: true
	// Attempt 3: true
	// Attempt 4: true
	// Attempt 5: true
}
