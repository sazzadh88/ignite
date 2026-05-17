package ratelimit

import (
	"net/http"
	"sync"
	"time"
)

// Limiter manages rate limiting for various named limiters.
type Limiter struct {
	store          Store
	namedLimiters  map[string]func(r *http.Request) *Limit
	limitersMutex  sync.RWMutex
}

// NewLimiter creates a new rate limiter with an in-memory store.
func NewLimiter() *Limiter {
	return &Limiter{
		store:         NewMemoryStore(),
		namedLimiters: make(map[string]func(r *http.Request) *Limit),
	}
}

// NewLimiterWithStore creates a new rate limiter with a custom store.
func NewLimiterWithStore(store Store) *Limiter {
	return &Limiter{
		store:         store,
		namedLimiters: make(map[string]func(r *http.Request) *Limit),
	}
}

// For defines a named rate limiter with a callback function that returns the limit configuration.
// The callback receives the HTTP request and can use it to determine the appropriate limit.
func (l *Limiter) For(name string, fn func(r *http.Request) *Limit) {
	l.limitersMutex.Lock()
	defer l.limitersMutex.Unlock()
	l.namedLimiters[name] = fn
}

// GetNamedLimiter retrieves a named limiter function.
func (l *Limiter) GetNamedLimiter(name string) (func(r *http.Request) *Limit, bool) {
	l.limitersMutex.RLock()
	defer l.limitersMutex.RUnlock()
	fn, exists := l.namedLimiters[name]
	return fn, exists
}

// Attempt tries to perform an attempt for the given key.
// Returns true if the attempt is allowed, false if the rate limit is exceeded.
func (l *Limiter) Attempt(key string, maxAttempts int, decayMinutes int) bool {
	if maxAttempts < 0 {
		// Unlimited
		return true
	}

	if l.TooManyAttempts(key, maxAttempts) {
		return false
	}

	decay := time.Duration(decayMinutes) * time.Minute
	l.store.Increment(key, decay)
	return true
}

// TooManyAttempts checks if the key has exceeded the maximum attempts.
func (l *Limiter) TooManyAttempts(key string, maxAttempts int) bool {
	if maxAttempts < 0 {
		// Unlimited
		return false
	}

	count, _, exists := l.store.Get(key)
	if !exists {
		return false
	}

	return count >= maxAttempts
}

// Hit increments the attempt count for the given key and returns the new count.
func (l *Limiter) Hit(key string, decayMinutes int) int {
	decay := time.Duration(decayMinutes) * time.Minute
	return l.store.Increment(key, decay)
}

// Clear resets all attempts for the given key.
func (l *Limiter) Clear(key string) {
	l.store.Reset(key)
}

// Remaining returns the number of remaining attempts for the given key.
func (l *Limiter) Remaining(key string, maxAttempts int) int {
	if maxAttempts < 0 {
		// Unlimited
		return maxAttempts
	}

	count, _, exists := l.store.Get(key)
	if !exists {
		return maxAttempts
	}

	remaining := maxAttempts - count
	if remaining < 0 {
		return 0
	}
	return remaining
}

// RetriesIn returns the time duration until attempts reset for the given key.
// Returns 0 if the key doesn't exist or has already expired.
func (l *Limiter) RetriesIn(key string) time.Duration {
	_, expiresAt, exists := l.store.Get(key)
	if !exists {
		return 0
	}

	remaining := time.Until(expiresAt)
	if remaining < 0 {
		return 0
	}
	return remaining
}

// AvailableIn is an alias for RetriesIn.
func (l *Limiter) AvailableIn(key string) time.Duration {
	return l.RetriesIn(key)
}

// RateLimiter is the global rate limiter instance.
var RateLimiter = NewLimiter()
