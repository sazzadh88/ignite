// Package cache provides caching functionality with multiple storage backends.
package cache

import (
	"fmt"
	"sync"
	"time"
)

// Store defines the interface for cache storage implementations.
type Store interface {
	// Get retrieves a value from the cache.
	Get(key string) (any, bool)

	// Put stores a value in the cache with a TTL.
	Put(key string, value any, ttl time.Duration) error

	// Forever stores a value in the cache permanently.
	Forever(key string, value any) error

	// Forget removes a value from the cache.
	Forget(key string) bool

	// Flush clears all values from the cache.
	Flush() error

	// Has checks if a key exists in the cache.
	Has(key string) bool
}

// Manager manages multiple cache stores.
type Manager struct {
	stores   map[string]Store
	default_ string
	mu       sync.RWMutex
}

// NewManager creates a new cache manager with the default memory store.
func NewManager() *Manager {
	m := &Manager{
		stores:   make(map[string]Store),
		default_: "memory",
	}

	// Register default driver
	m.Register("memory", NewMemoryStore())

	return m
}

// Register adds a cache store with the given name.
func (m *Manager) Register(name string, store Store) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.stores[name] = store
}

// Store returns a Repository for the given store name.
// Returns the default store if name is empty or not found.
func (m *Manager) Store(name string) *Repository {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name == "" {
		name = m.default_
	}

	store, ok := m.stores[name]
	if !ok {
		store = m.stores[m.default_]
	}

	return NewRepository(store)
}

// SetDefault sets the default cache store.
func (m *Manager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.stores[name]; !ok {
		return fmt.Errorf("cache store %q not found", name)
	}

	m.default_ = name
	return nil
}

// Cache is the package-level facade for the default manager.
var Cache = NewManager()
