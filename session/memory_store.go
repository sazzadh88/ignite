package session

import (
	"sync"
	"time"
)

// MemoryStore implements an in-memory session store.
// Useful for testing and development.
type MemoryStore struct {
	sessions map[string]*memorySession
	mu       sync.RWMutex
}

type memorySession struct {
	data      map[string]any
	expiresAt time.Time
}

// NewMemoryStore creates a new in-memory session store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*memorySession),
	}
}

// Read retrieves session data from memory.
func (m *MemoryStore) Read(id string) (map[string]any, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	sess, ok := m.sessions[id]
	if !ok {
		return make(map[string]any), nil
	}

	// Check if expired
	if time.Now().After(sess.expiresAt) {
		delete(m.sessions, id)
		return make(map[string]any), nil
	}

	// Return a copy to prevent external modification
	result := make(map[string]any, len(sess.data))
	for k, v := range sess.data {
		result[k] = v
	}

	return result, nil
}

// Write persists session data to memory.
func (m *MemoryStore) Write(id string, data map[string]any, ttl time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Make a copy of the data
	dataCopy := make(map[string]any, len(data))
	for k, v := range data {
		dataCopy[k] = v
	}

	m.sessions[id] = &memorySession{
		data:      dataCopy,
		expiresAt: time.Now().Add(ttl),
	}

	return nil
}

// Destroy removes session data from memory.
func (m *MemoryStore) Destroy(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sessions, id)
	return nil
}

// GC performs garbage collection on expired sessions.
func (m *MemoryStore) GC(maxLifetime time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	cutoff := time.Now().Add(-maxLifetime)

	for id, sess := range m.sessions {
		if sess.expiresAt.Before(cutoff) {
			delete(m.sessions, id)
		}
	}

	return nil
}
