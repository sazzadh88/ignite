// Package auth provides authentication functionality for Ignite.
//
// This package implements a Laravel-inspired authentication system with support for
// multiple guards (session-based, token-based) and user providers. It follows the
// facade pattern for convenient access to authentication features.
//
// # Architecture
//
// The auth package is built around these core concepts:
//
//   - Authenticatable: Interface that user models must implement
//   - Guard: Interface for authentication logic (session, token, etc.)
//   - UserProvider: Interface for retrieving and validating users
//   - Manager: Coordinates multiple guards
//
// # Basic Usage
//
// Session-based authentication:
//
//	provider := auth.NewCallbackProvider(
//	    func(id any) (auth.Authenticatable, error) {
//	        // Retrieve user from database by ID
//	    },
//	    func(credentials map[string]string) (auth.Authenticatable, error) {
//	        // Retrieve user by email/username
//	    },
//	    func(user auth.Authenticatable, credentials map[string]string) bool {
//	        // Validate password
//	    },
//	)
//
//	session := &mySession{} // Your session implementation
//	guard := auth.NewSessionGuard(provider, session)
//
//	manager := auth.GetManager()
//	manager.AddGuard("session", guard)
//
//	// Attempt login
//	if auth.Attempt(map[string]string{
//	    "email":    "user@example.com",
//	    "password": "secret",
//	}) {
//	    // User is authenticated
//	    user := auth.User()
//	    userID := auth.ID()
//	}
//
//	// Check authentication status
//	if auth.Check() {
//	    // User is authenticated
//	}
//
//	if auth.Guest() {
//	    // User is not authenticated
//	}
//
//	// Logout
//	auth.Logout()
//
// Token-based authentication (API):
//
//	storage := auth.NewMemoryTokenStorage()
//	tokenGuard := auth.NewTokenGuard(provider, storage)
//
//	// Create token
//	user := getUser()
//	token, err := tokenGuard.CreateToken(user, "api-token", []string{"read", "write"})
//	if err != nil {
//	    // Handle error
//	}
//
//	// Use plain text token in API requests
//	fmt.Println(token.PlainTextToken)
//
//	// Authenticate with token
//	err = tokenGuard.SetToken(token.PlainTextToken)
//	if err == nil && tokenGuard.Check() {
//	    // User is authenticated
//	}
//
// # Middleware
//
// The package provides middleware for protecting routes:
//
//	// Require authentication
//	authMiddleware := auth.Authenticate()
//
//	// Guest only (redirect if authenticated)
//	guestMiddleware := auth.RedirectIfAuthenticated()
//
//	// Token authentication
//	tokenMiddleware := auth.NewTokenAuthMiddleware(tokenGuard)
//
// # Password Reset
//
// Password reset functionality:
//
//	storage := auth.NewMemoryPasswordResetStorage()
//	resetter := &myPasswordResetter{} // Implement PasswordResetter interface
//	broker := auth.NewPasswordBroker(provider, storage, resetter)
//
//	// Send reset link
//	err := broker.SendResetLink("user@example.com")
//
//	// Reset password with token
//	err = broker.Reset(token, "user@example.com", "new_password")
//
// # Custom Guards
//
// You can implement custom guards by implementing the Guard interface:
//
//	type CustomGuard struct {
//	    // Your fields
//	}
//
//	func (g *CustomGuard) User() auth.Authenticatable { /* ... */ }
//	func (g *CustomGuard) ID() any { /* ... */ }
//	func (g *CustomGuard) Check() bool { /* ... */ }
//	func (g *CustomGuard) Guest() bool { /* ... */ }
//	func (g *CustomGuard) Validate(credentials map[string]string) bool { /* ... */ }
//
// # Thread Safety
//
// All guards and the manager are thread-safe and can be used concurrently.
//
// # Zero Dependencies
//
// This package has zero external dependencies and only uses the Go standard library.
package auth
