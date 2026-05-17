// Package auth provides authentication functionality for Ignite.
// It implements a Laravel-inspired authentication system with support for
// multiple guards (session, token) and user providers.
package auth

import (
	"errors"
	"sync"
)

var (
	// ErrInvalidCredentials is returned when credentials are invalid.
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrNotAuthenticated is returned when user is not authenticated.
	ErrNotAuthenticated = errors.New("user not authenticated")
	// ErrGuardNotFound is returned when a guard is not found.
	ErrGuardNotFound = errors.New("guard not found")
)

// Authenticatable represents a user that can be authenticated.
// Any type implementing this interface can be used with the auth system.
type Authenticatable interface {
	// GetAuthIdentifier returns the unique identifier for the user.
	GetAuthIdentifier() any
	// GetAuthPassword returns the hashed password for the user.
	GetAuthPassword() string
}

// Guard defines the interface for authentication guards.
// Guards handle the actual authentication logic (session, token, etc).
type Guard interface {
	// User returns the currently authenticated user.
	User() Authenticatable
	// ID returns the identifier for the currently authenticated user.
	ID() any
	// Check returns true if a user is authenticated.
	Check() bool
	// Guest returns true if no user is authenticated.
	Guest() bool
	// Validate checks if the given credentials are valid without logging in.
	Validate(credentials map[string]string) bool
}

// Manager manages multiple authentication guards.
type Manager struct {
	guards       map[string]Guard
	defaultGuard string
	mu           sync.RWMutex
}

// NewManager creates a new authentication manager.
func NewManager() *Manager {
	return &Manager{
		guards:       make(map[string]Guard),
		defaultGuard: "session",
	}
}

// Guard returns the named guard or the default guard if name is empty.
func (m *Manager) Guard(name string) Guard {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if name == "" {
		name = m.defaultGuard
	}

	guard, exists := m.guards[name]
	if !exists {
		return nil
	}
	return guard
}

// AddGuard registers a guard with the manager.
func (m *Manager) AddGuard(name string, guard Guard) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.guards[name] = guard
}

// SetDefaultGuard sets the default guard name.
func (m *Manager) SetDefaultGuard(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.defaultGuard = name
}

// DefaultGuard returns the default guard.
func (m *Manager) DefaultGuard() Guard {
	return m.Guard("")
}

// HasGuard checks if a guard exists.
func (m *Manager) HasGuard(name string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	_, exists := m.guards[name]
	return exists
}

// Package-level default manager for facade pattern
var defaultManager = NewManager()

// SetManager replaces the default manager (useful for testing).
func SetManager(manager *Manager) {
	defaultManager = manager
}

// GetManager returns the default manager.
func GetManager() *Manager {
	return defaultManager
}

// User returns the currently authenticated user from the default guard.
func User() Authenticatable {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return nil
	}
	return guard.User()
}

// ID returns the identifier for the currently authenticated user.
func ID() any {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return nil
	}
	return guard.ID()
}

// Check returns true if a user is authenticated.
func Check() bool {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return false
	}
	return guard.Check()
}

// Guest returns true if no user is authenticated.
func Guest() bool {
	return !Check()
}

// Attempt attempts to authenticate with the given credentials.
func Attempt(credentials map[string]string) bool {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return false
	}

	// Type assert to get Attempt method
	if sg, ok := guard.(interface{ Attempt(map[string]string) bool }); ok {
		return sg.Attempt(credentials)
	}
	return false
}

// Login logs in the given user.
func Login(user Authenticatable) {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return
	}

	// Type assert to get Login method
	if sg, ok := guard.(interface{ Login(Authenticatable) }); ok {
		sg.Login(user)
	}
}

// LoginUsingID logs in a user by their identifier.
func LoginUsingID(id any) error {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return ErrGuardNotFound
	}

	// Type assert to get LoginUsingID method
	if sg, ok := guard.(interface{ LoginUsingID(any) error }); ok {
		return sg.LoginUsingID(id)
	}
	return errors.New("guard does not support LoginUsingID")
}

// Logout logs out the currently authenticated user.
func Logout() {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return
	}

	// Type assert to get Logout method
	if sg, ok := guard.(interface{ Logout() }); ok {
		sg.Logout()
	}
}

// Once attempts to authenticate for a single request without session.
func Once(credentials map[string]string) bool {
	guard := defaultManager.DefaultGuard()
	if guard == nil {
		return false
	}

	// Type assert to get Once method
	if sg, ok := guard.(interface{ Once(map[string]string) bool }); ok {
		return sg.Once(credentials)
	}
	return false
}
