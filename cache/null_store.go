package cache

import "time"

// NullStore is a cache implementation that always misses.
// Useful for testing and disabling caching.
type NullStore struct{}

// NewNullStore creates a new null cache store.
func NewNullStore() *NullStore {
	return &NullStore{}
}

// Get always returns nil and false.
func (s *NullStore) Get(key string) (any, bool) {
	return nil, false
}

// Put does nothing.
func (s *NullStore) Put(key string, value any, ttl time.Duration) error {
	return nil
}

// Forever does nothing.
func (s *NullStore) Forever(key string, value any) error {
	return nil
}

// Forget always returns false.
func (s *NullStore) Forget(key string) bool {
	return false
}

// Flush does nothing.
func (s *NullStore) Flush() error {
	return nil
}

// Has always returns false.
func (s *NullStore) Has(key string) bool {
	return false
}
