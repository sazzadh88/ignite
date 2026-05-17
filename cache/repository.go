package cache

import (
	"fmt"
	"time"
)

// Repository wraps a Store with convenience methods.
type Repository struct {
	store Store
}

// NewRepository creates a new Repository wrapping the given Store.
func NewRepository(store Store) *Repository {
	return &Repository{store: store}
}

// Get retrieves a value from the cache, returning the default if not found.
func (r *Repository) Get(key string, defaultVal ...any) any {
	if val, ok := r.store.Get(key); ok {
		return val
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return nil
}

// GetString retrieves a string value from the cache.
func (r *Repository) GetString(key string, defaultVal ...string) string {
	val := r.Get(key)
	if s, ok := val.(string); ok {
		return s
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// GetInt retrieves an int value from the cache.
func (r *Repository) GetInt(key string, defaultVal ...int) int {
	val := r.Get(key)
	if i, ok := val.(int); ok {
		return i
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0
}

// Put stores a value in the cache with a TTL.
func (r *Repository) Put(key string, value any, ttl time.Duration) error {
	return r.store.Put(key, value, ttl)
}

// Add stores a value only if it doesn't already exist.
func (r *Repository) Add(key string, value any, ttl time.Duration) bool {
	if r.store.Has(key) {
		return false
	}
	err := r.store.Put(key, value, ttl)
	return err == nil
}

// Forever stores a value in the cache permanently.
func (r *Repository) Forever(key string, value any) error {
	return r.store.Forever(key, value)
}

// Pull retrieves a value and removes it from the cache.
func (r *Repository) Pull(key string) any {
	val, ok := r.store.Get(key)
	if ok {
		r.store.Forget(key)
		return val
	}
	return nil
}

// Has checks if a key exists in the cache.
func (r *Repository) Has(key string) bool {
	return r.store.Has(key)
}

// Missing checks if a key does not exist in the cache.
func (r *Repository) Missing(key string) bool {
	return !r.store.Has(key)
}

// Forget removes a value from the cache.
func (r *Repository) Forget(key string) bool {
	return r.store.Forget(key)
}

// Flush clears all values from the cache.
func (r *Repository) Flush() error {
	return r.store.Flush()
}

// Remember retrieves a value or stores the result of the callback if missing.
func (r *Repository) Remember(key string, ttl time.Duration, fn func() any) any {
	if val, ok := r.store.Get(key); ok {
		return val
	}

	val := fn()
	r.store.Put(key, val, ttl)
	return val
}

// RememberForever retrieves a value or stores the result of the callback permanently.
func (r *Repository) RememberForever(key string, fn func() any) any {
	if val, ok := r.store.Get(key); ok {
		return val
	}

	val := fn()
	r.store.Forever(key, val)
	return val
}

// Increment increments an integer value in the cache.
func (r *Repository) Increment(key string, amount ...int) (int, error) {
	delta := 1
	if len(amount) > 0 {
		delta = amount[0]
	}

	val := r.GetInt(key, 0)
	newVal := val + delta

	if err := r.store.Put(key, newVal, 0); err != nil {
		return 0, err
	}

	return newVal, nil
}

// Decrement decrements an integer value in the cache.
func (r *Repository) Decrement(key string, amount ...int) (int, error) {
	delta := 1
	if len(amount) > 0 {
		delta = amount[0]
	}

	return r.Increment(key, -delta)
}

// Many retrieves multiple values from the cache.
func (r *Repository) Many(keys []string) map[string]any {
	result := make(map[string]any, len(keys))
	for _, key := range keys {
		if val, ok := r.store.Get(key); ok {
			result[key] = val
		}
	}
	return result
}

// PutMany stores multiple values in the cache with a TTL.
func (r *Repository) PutMany(values map[string]any, ttl time.Duration) error {
	for key, val := range values {
		if err := r.store.Put(key, val, ttl); err != nil {
			return fmt.Errorf("failed to put key %q: %w", key, err)
		}
	}
	return nil
}

// Lock returns a new Lock for the given key.
func (r *Repository) Lock(key string, ttl time.Duration) *Lock {
	return NewLock(r, key, ttl)
}

// Tags returns a new TaggedCache for the given tags.
func (r *Repository) Tags(tags []string) *TaggedCache {
	return NewTaggedCache(r, tags)
}
