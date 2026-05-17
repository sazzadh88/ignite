package session

import (
	"fmt"
	"net/http"
	"time"
)

// SessionConfig holds session configuration.
type SessionConfig struct {
	// Driver specifies the session storage driver ("file", "memory", "cookie").
	Driver string

	// CookieName is the name of the session cookie.
	CookieName string

	// Lifetime is the session lifetime in minutes.
	Lifetime int

	// Path is the cookie path.
	Path string

	// Domain is the cookie domain.
	Domain string

	// Secure indicates if the cookie should only be sent over HTTPS.
	Secure bool

	// HTTPOnly indicates if the cookie should be HTTP-only.
	HTTPOnly bool

	// SameSite sets the SameSite cookie attribute.
	SameSite http.SameSite

	// Files is the directory for file-based sessions.
	Files string

	// EncryptionKey is used for cookie encryption (16, 24, or 32 bytes).
	EncryptionKey []byte

	// SigningKey is used for cookie signing.
	SigningKey []byte
}

// DefaultConfig returns a default session configuration.
func DefaultConfig() SessionConfig {
	return SessionConfig{
		Driver:     "memory",
		CookieName: "session",
		Lifetime:   120, // 2 hours
		Path:       "/",
		Domain:     "",
		Secure:     false,
		HTTPOnly:   true,
		SameSite:   http.SameSiteLaxMode,
		Files:      "/tmp/sessions",
	}
}

// Manager manages session stores and lifecycle.
type Manager struct {
	config SessionConfig
	stores map[string]Store
}

// NewManager creates a new session manager.
func NewManager(config SessionConfig) (*Manager, error) {
	m := &Manager{
		config: config,
		stores: make(map[string]Store),
	}

	// Initialize default store based on driver
	store, err := m.createStore(config.Driver)
	if err != nil {
		return nil, fmt.Errorf("failed to create default store: %w", err)
	}

	m.stores[config.Driver] = store

	return m, nil
}

// Driver returns a store by name.
func (m *Manager) Driver(name string) Store {
	if store, ok := m.stores[name]; ok {
		return store
	}

	// Try to create the store
	store, err := m.createStore(name)
	if err != nil {
		return nil
	}

	m.stores[name] = store
	return store
}

// Start loads a session from the request.
func (m *Manager) Start(r *http.Request) (*Session, error) {
	cookie, err := r.Cookie(m.config.CookieName)
	if err != nil && err != http.ErrNoCookie {
		return nil, fmt.Errorf("failed to read session cookie: %w", err)
	}

	var sessionID string
	if cookie != nil {
		sessionID = cookie.Value
	}

	store := m.Driver(m.config.Driver)
	if store == nil {
		return nil, fmt.Errorf("session store not available")
	}

	session := NewSession(store, sessionID)
	if err := session.Start(); err != nil {
		return nil, fmt.Errorf("failed to start session: %w", err)
	}

	return session, nil
}

// Save persists the session and sets the cookie.
func (m *Manager) Save(session *Session, w http.ResponseWriter) error {
	if err := session.Save(); err != nil {
		return fmt.Errorf("failed to save session: %w", err)
	}

	// Set session cookie
	cookie := &http.Cookie{
		Name:     m.config.CookieName,
		Value:    session.GetID(),
		Path:     m.config.Path,
		Domain:   m.config.Domain,
		MaxAge:   m.config.Lifetime * 60,
		Secure:   m.config.Secure,
		HttpOnly: m.config.HTTPOnly,
		SameSite: m.config.SameSite,
	}

	http.SetCookie(w, cookie)

	return nil
}

// GC performs garbage collection on all stores.
func (m *Manager) GC() error {
	maxLifetime := time.Duration(m.config.Lifetime) * time.Minute

	for _, store := range m.stores {
		if err := store.GC(maxLifetime); err != nil {
			return fmt.Errorf("garbage collection failed: %w", err)
		}
	}

	return nil
}

// createStore creates a store instance by name.
func (m *Manager) createStore(name string) (Store, error) {
	switch name {
	case "file":
		return NewFileStore(m.config.Files)
	case "memory":
		return NewMemoryStore(), nil
	case "cookie":
		if len(m.config.EncryptionKey) == 0 {
			return nil, fmt.Errorf("encryption key required for cookie store")
		}
		return NewCookieStore(m.config.EncryptionKey, m.config.SigningKey)
	default:
		return nil, fmt.Errorf("unknown session driver: %s", name)
	}
}
