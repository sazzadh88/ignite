package ratelimit

import (
	"sync"
	"time"
)

// Store is the interface for rate limit storage backends.
type Store interface {
	// Get retrieves the current attempt count and expiry time for a key.
	// Returns (count, expiresAt, exists).
	Get(key string) (int, time.Time, bool)

	// Increment increments the attempt count for a key with the given decay duration.
	// Returns the new count.
	Increment(key string, decay time.Duration) int

	// Reset clears all attempts for a key.
	Reset(key string)

	// Clean removes expired entries from the store.
	Clean()
}

// MemoryStore is an in-memory implementation of Store.
type MemoryStore struct {
	mu      sync.RWMutex
	entries map[string]*entry
}

type entry struct {
	count     int
	expiresAt time.Time
}

// NewMemoryStore creates a new in-memory rate limit store.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{
		entries: make(map[string]*entry),
	}
	// Start background cleanup goroutine
	go store.cleanupLoop()
	return store
}

// Get retrieves the current attempt count and expiry time for a key.
func (s *MemoryStore) Get(key string) (int, time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	e, exists := s.entries[key]
	if !exists {
		return 0, time.Time{}, false
	}

	// Check if expired
	if time.Now().After(e.expiresAt) {
		return 0, time.Time{}, false
	}

	return e.count, e.expiresAt, true
}

// Increment increments the attempt count for a key with the given decay duration.
func (s *MemoryStore) Increment(key string, decay time.Duration) int {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	expiresAt := now.Add(decay)

	e, exists := s.entries[key]
	if !exists || now.After(e.expiresAt) {
		// Create new entry or reset expired entry
		s.entries[key] = &entry{
			count:     1,
			expiresAt: expiresAt,
		}
		return 1
	}

	// Increment existing entry
	e.count++
	return e.count
}

// Reset clears all attempts for a key.
func (s *MemoryStore) Reset(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, key)
}

// Clean removes expired entries from the store.
func (s *MemoryStore) Clean() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for key, e := range s.entries {
		if now.After(e.expiresAt) {
			delete(s.entries, key)
		}
	}
}

// cleanupLoop runs periodic cleanup in the background.
func (s *MemoryStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.Clean()
	}
}
