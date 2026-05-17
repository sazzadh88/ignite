package auth

import (
	"sync"
)

// Session defines the interface for session storage.
type Session interface {
	// Get retrieves a value from the session.
	Get(key string) (any, bool)
	// Put stores a value in the session.
	Put(key, value any)
	// Forget removes a value from the session.
	Forget(key string)
	// Flush clears all session data.
	Flush()
}

// SessionGuard implements session-based authentication.
type SessionGuard struct {
	provider     UserProvider
	session      Session
	user         Authenticatable
	sessionKey   string
	rememberKey  string
	loggedOut    bool
	mu           sync.RWMutex
}

// NewSessionGuard creates a new session guard.
func NewSessionGuard(provider UserProvider, session Session) *SessionGuard {
	return &SessionGuard{
		provider:    provider,
		session:     session,
		sessionKey:  "auth_user_id",
		rememberKey: "auth_remember",
	}
}

// User returns the currently authenticated user.
func (g *SessionGuard) User() Authenticatable {
	g.mu.RLock()
	if g.user != nil {
		g.mu.RUnlock()
		return g.user
	}
	g.mu.RUnlock()

	// Try to load user from session
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.loggedOut {
		return nil
	}

	// Check if user ID exists in session
	userID, exists := g.session.Get(g.sessionKey)
	if !exists || userID == nil {
		return nil
	}

	// Retrieve user from provider
	user, err := g.provider.RetrieveByID(userID)
	if err != nil {
		return nil
	}

	g.user = user
	return g.user
}

// ID returns the identifier for the currently authenticated user.
func (g *SessionGuard) ID() any {
	user := g.User()
	if user == nil {
		return nil
	}
	return user.GetAuthIdentifier()
}

// Check returns true if a user is authenticated.
func (g *SessionGuard) Check() bool {
	return g.User() != nil
}

// Guest returns true if no user is authenticated.
func (g *SessionGuard) Guest() bool {
	return !g.Check()
}

// Validate checks if the given credentials are valid without logging in.
func (g *SessionGuard) Validate(credentials map[string]string) bool {
	user, err := g.provider.RetrieveByCredentials(credentials)
	if err != nil {
		return false
	}

	return g.provider.ValidateCredentials(user, credentials)
}

// Attempt attempts to authenticate with the given credentials.
// Returns true if authentication was successful.
func (g *SessionGuard) Attempt(credentials map[string]string) bool {
	user, err := g.provider.RetrieveByCredentials(credentials)
	if err != nil {
		return false
	}

	if !g.provider.ValidateCredentials(user, credentials) {
		return false
	}

	g.Login(user)
	return true
}

// Login logs in the given user.
func (g *SessionGuard) Login(user Authenticatable) {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.user = user
	g.loggedOut = false
	g.session.Put(g.sessionKey, user.GetAuthIdentifier())
}

// LoginUsingID logs in a user by their identifier.
func (g *SessionGuard) LoginUsingID(id any) error {
	user, err := g.provider.RetrieveByID(id)
	if err != nil {
		return err
	}

	g.Login(user)
	return nil
}

// Logout logs out the currently authenticated user.
func (g *SessionGuard) Logout() {
	g.mu.Lock()
	defer g.mu.Unlock()

	g.user = nil
	g.loggedOut = true
	g.session.Forget(g.sessionKey)
	g.session.Forget(g.rememberKey)
}

// Once attempts to authenticate for a single request without session.
// This is useful for stateless authentication in specific contexts.
func (g *SessionGuard) Once(credentials map[string]string) bool {
	user, err := g.provider.RetrieveByCredentials(credentials)
	if err != nil {
		return false
	}

	if !g.provider.ValidateCredentials(user, credentials) {
		return false
	}

	// Set user but don't persist to session
	g.mu.Lock()
	g.user = user
	g.loggedOut = false
	g.mu.Unlock()

	return true
}

// SetSessionKey sets the session key name.
func (g *SessionGuard) SetSessionKey(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.sessionKey = key
}

// SetRememberKey sets the remember token key name.
func (g *SessionGuard) SetRememberKey(key string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.rememberKey = key
}
