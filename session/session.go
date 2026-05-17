package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"
)

// Store defines the session storage interface.
// Implementations should be thread-safe.
type Store interface {
	// Read retrieves session data by ID.
	Read(id string) (map[string]any, error)

	// Write persists session data with the given TTL.
	Write(id string, data map[string]any, ttl time.Duration) error

	// Destroy removes the session data.
	Destroy(id string) error

	// GC performs garbage collection on expired sessions.
	GC(maxLifetime time.Duration) error
}

// Session represents an HTTP session.
type Session struct {
	id         string
	store      Store
	data       map[string]any
	flash      map[string]any
	newFlash   map[string]any
	token      string
	started    bool
	mu         sync.RWMutex
}

// NewSession creates a new session instance.
func NewSession(store Store, id string) *Session {
	if id == "" {
		id = generateID()
	}
	return &Session{
		id:       id,
		store:    store,
		data:     make(map[string]any),
		flash:    make(map[string]any),
		newFlash: make(map[string]any),
	}
}

// Start loads session data from the store.
func (s *Session) Start() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.started {
		return nil
	}

	data, err := s.store.Read(s.id)
	if err != nil {
		return err
	}

	if data == nil {
		data = make(map[string]any)
	}

	s.data = data

	// Extract flash data
	if flashData, ok := s.data["_flash"].(map[string]any); ok {
		if newData, ok := flashData["new"].(map[string]any); ok {
			s.flash = newData
		}
		delete(s.data, "_flash")
	}

	// Extract token
	if token, ok := s.data["_token"].(string); ok {
		s.token = token
	} else {
		s.token = generateID()
	}

	s.newFlash = make(map[string]any)
	s.started = true

	return nil
}

// Save persists session data to the store.
func (s *Session) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Prepare flash data
	if len(s.newFlash) > 0 {
		s.data["_flash"] = map[string]any{
			"new": s.newFlash,
		}
	} else {
		delete(s.data, "_flash")
	}

	// Store token
	s.data["_token"] = s.token

	return s.store.Write(s.id, s.data, 2*time.Hour)
}

// Put stores a value in the session.
func (s *Session) Put(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// PutMany stores multiple values in the session.
func (s *Session) PutMany(data map[string]any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range data {
		s.data[k] = v
	}
}

// Get retrieves a value from the session.
func (s *Session) Get(key string, defaultVal ...any) any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if val, ok := s.data[key]; ok {
		return val
	}

	if val, ok := s.flash[key]; ok {
		return val
	}

	if len(defaultVal) > 0 {
		return defaultVal[0]
	}

	return nil
}

// GetString retrieves a string value from the session.
func (s *Session) GetString(key string, defaultVal ...string) string {
	val := s.Get(key)
	if str, ok := val.(string); ok {
		return str
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return ""
}

// GetInt retrieves an integer value from the session.
func (s *Session) GetInt(key string, defaultVal ...int) int {
	val := s.Get(key)
	if i, ok := val.(int); ok {
		return i
	}
	if len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return 0
}

// Pull retrieves a value and then removes it from the session.
func (s *Session) Pull(key string) any {
	s.mu.Lock()
	defer s.mu.Unlock()

	if val, ok := s.data[key]; ok {
		delete(s.data, key)
		return val
	}

	if val, ok := s.flash[key]; ok {
		delete(s.flash, key)
		return val
	}

	return nil
}

// Has checks if a key exists in the session (including flash).
func (s *Session) Has(key string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.data[key]; ok {
		return true
	}

	_, ok := s.flash[key]
	return ok
}

// Exists checks if a key exists and is not nil.
func (s *Session) Exists(key string) bool {
	val := s.Get(key)
	return val != nil
}

// Missing checks if a key does not exist in the session.
func (s *Session) Missing(key string) bool {
	return !s.Has(key)
}

// All returns all session data (excluding flash).
func (s *Session) All() map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any, len(s.data))
	for k, v := range s.data {
		result[k] = v
	}
	return result
}

// Only returns only the specified keys from the session.
func (s *Session) Only(keys []string) map[string]any {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]any)
	for _, key := range keys {
		if val, ok := s.data[key]; ok {
			result[key] = val
		}
	}
	return result
}

// Forget removes one or more keys from the session.
func (s *Session) Forget(keys ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.data, key)
	}
}

// Flush removes all data from the session.
func (s *Session) Flush() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data = make(map[string]any)
	s.flash = make(map[string]any)
	s.newFlash = make(map[string]any)
}

// Regenerate generates a new session ID while keeping the data.
func (s *Session) Regenerate() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.id = generateID()
	return s.id
}

// Invalidate flushes all data and regenerates the ID.
func (s *Session) Invalidate() string {
	s.Flush()
	return s.Regenerate()
}

// Token returns the CSRF token for the session.
func (s *Session) Token() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.token
}

// GetID returns the session ID.
func (s *Session) GetID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.id
}

// SetID sets the session ID.
func (s *Session) SetID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.id = id
}

// Flash stores a value that will be available only for the next request.
func (s *Session) Flash(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.newFlash[key] = value
}

// Now stores a value that will be available for the current request only.
func (s *Session) Now(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.flash[key] = value
}

// Reflash keeps all flash data for one more request.
func (s *Session) Reflash() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for k, v := range s.flash {
		s.newFlash[k] = v
	}
}

// KeepFlash keeps specific flash data for one more request.
func (s *Session) KeepFlash(keys ...string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		if val, ok := s.flash[key]; ok {
			s.newFlash[key] = val
		}
	}
}

// OldInput retrieves old input data (typically from flash).
func (s *Session) OldInput(key string) any {
	return s.Get("_old_input." + key)
}

// Increment increments a numeric session value.
func (s *Session) Increment(key string, amount ...int) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delta := 1
	if len(amount) > 0 {
		delta = amount[0]
	}

	current := 0
	if val, ok := s.data[key].(int); ok {
		current = val
	}

	s.data[key] = current + delta
}

// Decrement decrements a numeric session value.
func (s *Session) Decrement(key string, amount ...int) {
	delta := 1
	if len(amount) > 0 {
		delta = amount[0]
	}
	s.Increment(key, -delta)
}

// Push appends a value to an array session value.
func (s *Session) Push(key string, value any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	var arr []any
	if existing, ok := s.data[key].([]any); ok {
		arr = existing
	}

	arr = append(arr, value)
	s.data[key] = arr
}

// generateID generates a secure random session ID.
func generateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Fallback to timestamp-based ID
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b)
}
