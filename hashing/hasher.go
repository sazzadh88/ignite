// Package hashing provides password hashing and verification utilities.
package hashing

import (
	"fmt"
	"sync"
)

// Hasher defines the interface for password hashing implementations.
type Hasher interface {
	// Make generates a hash for the given value.
	Make(value string) (string, error)

	// Check validates a plain-text value against its hash.
	Check(value, hash string) bool

	// NeedsRehash checks if the hash needs to be regenerated (e.g., due to updated parameters).
	NeedsRehash(hash string) bool
}

// Manager manages multiple hasher drivers.
type Manager struct {
	drivers map[string]Hasher
	default_ string
	mu      sync.RWMutex
}

// NewManager creates a new hasher manager with the default SHA256 driver.
func NewManager() *Manager {
	m := &Manager{
		drivers:  make(map[string]Hasher),
		default_: "sha256",
	}

	// Register default driver
	m.Register("sha256", NewSHA256Hasher(10000))

	return m
}

// Register adds a hasher driver with the given name.
func (m *Manager) Register(name string, hasher Hasher) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[name] = hasher
}

// Driver returns the hasher for the given driver name.
// Returns the default driver if name is empty or not found.
func (m *Manager) Driver(name string) Hasher {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name == "" {
		name = m.default_
	}

	if driver, ok := m.drivers[name]; ok {
		return driver
	}

	return m.drivers[m.default_]
}

// Make generates a hash using the default driver.
func (m *Manager) Make(value string) (string, error) {
	return m.Driver("").Make(value)
}

// Check validates a value against its hash using the default driver.
func (m *Manager) Check(value, hash string) bool {
	return m.Driver("").Check(value, hash)
}

// NeedsRehash checks if the hash needs regeneration using the default driver.
func (m *Manager) NeedsRehash(hash string) bool {
	return m.Driver("").NeedsRehash(hash)
}

// SetDefault sets the default hasher driver.
func (m *Manager) SetDefault(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.drivers[name]; !ok {
		return fmt.Errorf("hasher driver %q not found", name)
	}

	m.default_ = name
	return nil
}

// Hash is the package-level facade for the default manager.
var Hash = NewManager()
