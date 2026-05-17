package cache

import (
	"sync"
	"time"
)

type memoryItem struct {
	value  any
	expiry time.Time
}

func (i *memoryItem) isExpired() bool {
	return !i.expiry.IsZero() && time.Now().After(i.expiry)
}

// MemoryStore is an in-memory cache implementation.
type MemoryStore struct {
	items map[string]*memoryItem
	mu    sync.RWMutex
}

// NewMemoryStore creates a new in-memory cache store.
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		items: make(map[string]*memoryItem),
	}

	// Start background cleanup goroutine
	go s.cleanup()

	return s
}

// Get retrieves a value from the cache.
func (s *MemoryStore) Get(key string) (any, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	if !ok {
		return nil, false
	}

	if item.isExpired() {
		return nil, false
	}

	return item.value, true
}

// Put stores a value in the cache with a TTL.
func (s *MemoryStore) Put(key string, value any, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiry time.Time
	if ttl > 0 {
		expiry = time.Now().Add(ttl)
	}

	s.items[key] = &memoryItem{
		value:  value,
		expiry: expiry,
	}

	return nil
}

// Forever stores a value in the cache permanently.
func (s *MemoryStore) Forever(key string, value any) error {
	return s.Put(key, value, 0)
}

// Forget removes a value from the cache.
func (s *MemoryStore) Forget(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.items[key]
	if ok {
		delete(s.items, key)
	}

	return ok
}

// Flush clears all values from the cache.
func (s *MemoryStore) Flush() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]*memoryItem)
	return nil
}

// Has checks if a key exists in the cache.
func (s *MemoryStore) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	if !ok {
		return false
	}

	return !item.isExpired()
}

// cleanup removes expired items in the background.
func (s *MemoryStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		for key, item := range s.items {
			if item.isExpired() {
				delete(s.items, key)
			}
		}
		s.mu.Unlock()
	}
}
