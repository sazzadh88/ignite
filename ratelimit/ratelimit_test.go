package ratelimit

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestAttemptSucceedsWithinLimit tests that attempts succeed when within the limit.
func TestAttemptSucceedsWithinLimit(t *testing.T) {
	limiter := NewLimiter()
	key := "test-key"

	// First 3 attempts should succeed
	for i := 0; i < 3; i++ {
		if !limiter.Attempt(key, 3, 1) {
			t.Errorf("Attempt %d should succeed", i+1)
		}
	}
}

// TestAttemptFailsWhenExceeded tests that attempts fail when the limit is exceeded.
func TestAttemptFailsWhenExceeded(t *testing.T) {
	limiter := NewLimiter()
	key := "test-key"

	// Use up all attempts
	for i := 0; i < 3; i++ {
		limiter.Attempt(key, 3, 1)
	}

	// Next attempt should fail
	if limiter.Attempt(key, 3, 1) {
		t.Error("Attempt should fail after limit exceeded")
	}
}

// TestTooManyAttemptsDetection tests the TooManyAttempts method.
func TestTooManyAttemptsDetection(t *testing.T) {
	limiter := NewLimiter()
	key := "test-key"

	// Should not have too many attempts initially
	if limiter.TooManyAttempts(key, 3) {
		t.Error("Should not have too many attempts initially")
	}

	// Use up all attempts
	for i := 0; i < 3; i++ {
		limiter.Hit(key, 1)
	}

	// Now should have too many attempts
	if !limiter.TooManyAttempts(key, 3) {
		t.Error("Should have too many attempts after hitting limit")
	}
}

// TestClearResetsCounter tests that Clear resets the attempt counter.
func TestClearResetsCounter(t *testing.T) {
	limiter := NewLimiter()
	key := "test-key"

	// Use up all attempts
	for i := 0; i < 3; i++ {
		limiter.Hit(key, 1)
	}

	// Clear the counter
	limiter.Clear(key)

	// Should be able to attempt again
	if !limiter.Attempt(key, 3, 1) {
		t.Error("Attempt should succeed after clear")
	}
}

// TestRemainingCountAccurate tests that Remaining returns the correct count.
func TestRemainingCountAccurate(t *testing.T) {
	limiter := NewLimiter()
	key := "test-key"
	maxAttempts := 5

	// Initially should have all attempts remaining
	if remaining := limiter.Remaining(key, maxAttempts); remaining != maxAttempts {
		t.Errorf("Expected %d remaining, got %d", maxAttempts, remaining)
	}

	// Use 2 attempts
	limiter.Hit(key, 1)
	limiter.Hit(key, 1)

	// Should have 3 remaining
	if remaining := limiter.Remaining(key, maxAttempts); remaining != 3 {
		t.Errorf("Expected 3 remaining, got %d", remaining)
	}

	// Use all remaining attempts
	for i := 0; i < 3; i++ {
		limiter.Hit(key, 1)
	}

	// Should have 0 remaining
	if remaining := limiter.Remaining(key, maxAttempts); remaining != 0 {
		t.Errorf("Expected 0 remaining, got %d", remaining)
	}
}

// TestDecayTimeRespected tests that the decay time is respected.
func TestDecayTimeRespected(t *testing.T) {
	limiter := NewLimiter()
	key := "test-key"

	// Use up all attempts with 100ms decay
	decayMinutes := 0.001666 // ~100ms in minutes
	for i := 0; i < 3; i++ {
		limiter.Hit(key, int(decayMinutes*60*1000)) // Convert to minutes
	}

	// Should have too many attempts
	if !limiter.TooManyAttempts(key, 3) {
		t.Error("Should have too many attempts")
	}

	// Wait for decay (use actual minutes-based decay for testing)
	// Instead, let's use a very short decay time
	limiter2 := NewLimiter()
	key2 := "test-key-2"

	// Hit with 1 minute decay
	for i := 0; i < 3; i++ {
		limiter2.Hit(key2, 0) // 0 minutes = immediate expiry
	}

	// After immediate expiry, should be able to attempt again
	time.Sleep(10 * time.Millisecond)

	// With 0 minute decay, the entry expires immediately
	// So a new attempt should succeed
	if !limiter2.Attempt(key2, 3, 1) {
		t.Error("Attempt should succeed after decay")
	}
}

// TestNamedLimiterWithFor tests defining named limiters with For().
func TestNamedLimiterWithFor(t *testing.T) {
	limiter := NewLimiter()

	// Define a named limiter
	limiter.For("api", func(r *http.Request) *Limit {
		return PerMinute(10)
	})

	// Retrieve the named limiter
	fn, exists := limiter.GetNamedLimiter("api")
	if !exists {
		t.Fatal("Named limiter 'api' should exist")
	}

	// Test the limiter function
	req := httptest.NewRequest("GET", "/api/test", nil)
	limit := fn(req)

	if limit.MaxAttempts != 10 {
		t.Errorf("Expected MaxAttempts=10, got %d", limit.MaxAttempts)
	}
	if limit.DecayMinutes != 1 {
		t.Errorf("Expected DecayMinutes=1, got %d", limit.DecayMinutes)
	}
}

// TestMiddlewareReturns429WhenExceeded tests that middleware returns 429 when limit exceeded.
func TestMiddlewareReturns429WhenExceeded(t *testing.T) {
	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// Apply throttle middleware with very low limit
	throttled := ThrottleWithLimit(2, 1)(handler)

	// Make requests
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.1:1234"
		w := httptest.NewRecorder()
		throttled.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should return 200, got %d", i+1, w.Code)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	w := httptest.NewRecorder()
	throttled.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}
}

// TestMiddlewareSetsRateLimitHeaders tests that middleware sets rate limit headers.
func TestMiddlewareSetsRateLimitHeaders(t *testing.T) {
	// Reset global limiter
	RateLimiter = NewLimiter()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	throttled := ThrottleWithLimit(5, 1)(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.2:1234"
	w := httptest.NewRecorder()
	throttled.ServeHTTP(w, req)

	// Check headers
	if limit := w.Header().Get("X-RateLimit-Limit"); limit != "5" {
		t.Errorf("Expected X-RateLimit-Limit=5, got %s", limit)
	}

	if remaining := w.Header().Get("X-RateLimit-Remaining"); remaining != "4" {
		t.Errorf("Expected X-RateLimit-Remaining=4, got %s", remaining)
	}
}

// TestPerMinuteCreatesCorrectLimit tests that PerMinute creates correct limit.
func TestPerMinuteCreatesCorrectLimit(t *testing.T) {
	limit := PerMinute(60)

	if limit.MaxAttempts != 60 {
		t.Errorf("Expected MaxAttempts=60, got %d", limit.MaxAttempts)
	}
	if limit.DecayMinutes != 1 {
		t.Errorf("Expected DecayMinutes=1, got %d", limit.DecayMinutes)
	}
}

// TestPerHourCreatesCorrectLimit tests that PerHour creates correct limit.
func TestPerHourCreatesCorrectLimit(t *testing.T) {
	limit := PerHour(100)

	if limit.MaxAttempts != 100 {
		t.Errorf("Expected MaxAttempts=100, got %d", limit.MaxAttempts)
	}
	if limit.DecayMinutes != 60 {
		t.Errorf("Expected DecayMinutes=60, got %d", limit.DecayMinutes)
	}
}

// TestPerDayCreatesCorrectLimit tests that PerDay creates correct limit.
func TestPerDayCreatesCorrectLimit(t *testing.T) {
	limit := PerDay(1000)

	if limit.MaxAttempts != 1000 {
		t.Errorf("Expected MaxAttempts=1000, got %d", limit.MaxAttempts)
	}
	if limit.DecayMinutes != 1440 {
		t.Errorf("Expected DecayMinutes=1440, got %d", limit.DecayMinutes)
	}
}

// TestNoneAllowsUnlimited tests that None() allows unlimited requests.
func TestNoneAllowsUnlimited(t *testing.T) {
	limit := None()

	if limit.MaxAttempts != -1 {
		t.Errorf("Expected MaxAttempts=-1, got %d", limit.MaxAttempts)
	}

	// Test with limiter
	limiter := NewLimiter()
	key := "unlimited-key"

	// Should allow unlimited attempts
	for i := 0; i < 1000; i++ {
		if !limiter.Attempt(key, limit.MaxAttempts, limit.DecayMinutes) {
			t.Errorf("Unlimited limiter should allow attempt %d", i+1)
		}
	}
}

// TestByChaining tests that By() allows chaining.
func TestByChaining(t *testing.T) {
	limit := PerMinute(60).By("user:123")

	if limit.Key != "user:123" {
		t.Errorf("Expected Key='user:123', got '%s'", limit.Key)
	}
	if limit.MaxAttempts != 60 {
		t.Errorf("Expected MaxAttempts=60, got %d", limit.MaxAttempts)
	}
}

// TestRetriesIn tests the RetriesIn method.
func TestRetriesIn(t *testing.T) {
	limiter := NewLimiter()
	key := "test-retries"

	// Hit the limiter
	limiter.Hit(key, 1) // 1 minute decay

	// Check retries time
	retriesIn := limiter.RetriesIn(key)
	if retriesIn <= 0 {
		t.Error("RetriesIn should return positive duration")
	}
	if retriesIn > 1*time.Minute {
		t.Error("RetriesIn should be less than 1 minute")
	}
}

// TestAvailableIn tests that AvailableIn is an alias for RetriesIn.
func TestAvailableIn(t *testing.T) {
	limiter := NewLimiter()
	key := "test-available"

	limiter.Hit(key, 1)

	retriesIn := limiter.RetriesIn(key)
	availableIn := limiter.AvailableIn(key)

	// Allow small time difference due to execution time
	diff := retriesIn - availableIn
	if diff < 0 {
		diff = -diff
	}
	if diff > 10*time.Millisecond {
		t.Error("AvailableIn and RetriesIn should return similar values")
	}
}

// TestStoreClean tests that expired entries are cleaned.
func TestStoreClean(t *testing.T) {
	store := NewMemoryStore()

	// Add an entry with very short expiry
	store.Increment("test-key", 10*time.Millisecond)

	// Wait for expiry
	time.Sleep(20 * time.Millisecond)

	// Clean should remove it
	store.Clean()

	// Entry should not exist
	_, _, exists := store.Get("test-key")
	if exists {
		t.Error("Expired entry should be cleaned")
	}
}

// TestHitReturnsCorrectCount tests that Hit returns the correct count.
func TestHitReturnsCorrectCount(t *testing.T) {
	limiter := NewLimiter()
	key := "test-hit"

	// First hit should return 1
	if count := limiter.Hit(key, 1); count != 1 {
		t.Errorf("Expected count=1, got %d", count)
	}

	// Second hit should return 2
	if count := limiter.Hit(key, 1); count != 2 {
		t.Errorf("Expected count=2, got %d", count)
	}

	// Third hit should return 3
	if count := limiter.Hit(key, 1); count != 3 {
		t.Errorf("Expected count=3, got %d", count)
	}
}

// TestThrottleWithNamedLimiter tests the Throttle middleware with named limiters.
func TestThrottleWithNamedLimiter(t *testing.T) {
	// Reset global limiter
	RateLimiter = NewLimiter()

	// Define named limiter
	RateLimiter.For("test-api", func(r *http.Request) *Limit {
		return PerMinute(2)
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	throttled := Throttle("test-api")(handler)

	// First two requests should succeed
	for i := 0; i < 2; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.3:1234"
		w := httptest.NewRecorder()
		throttled.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Request %d should succeed", i+1)
		}
	}

	// Third request should be rate limited
	req := httptest.NewRequest("GET", "/test", nil)
	req.RemoteAddr = "192.168.1.3:1234"
	w := httptest.NewRecorder()
	throttled.ServeHTTP(w, req)

	if w.Code != http.StatusTooManyRequests {
		t.Errorf("Expected status 429, got %d", w.Code)
	}

	// Check Retry-After header
	if retryAfter := w.Header().Get("Retry-After"); retryAfter == "" {
		t.Error("Retry-After header should be set")
	}
}

// TestThrottleWithUnlimitedNamedLimiter tests that unlimited named limiters work.
func TestThrottleWithUnlimitedNamedLimiter(t *testing.T) {
	// Reset global limiter
	RateLimiter = NewLimiter()

	// Define unlimited limiter
	RateLimiter.For("unlimited", func(r *http.Request) *Limit {
		return None()
	})

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	throttled := Throttle("unlimited")(handler)

	// Should allow many requests
	for i := 0; i < 100; i++ {
		req := httptest.NewRequest("GET", "/test", nil)
		req.RemoteAddr = "192.168.1.4:1234"
		w := httptest.NewRecorder()
		throttled.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Unlimited limiter should allow request %d", i+1)
		}
	}
}
