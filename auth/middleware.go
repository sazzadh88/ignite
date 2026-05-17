package auth

import (
	"net/http"
)

// Next is the middleware next function type.
type Next func(http.ResponseWriter, *http.Request)

// AuthMiddleware is a middleware that requires authentication.
type AuthMiddleware struct {
	guard        Guard
	redirectTo   string
	abortOnFail  bool
}

// NewAuthMiddleware creates a new authentication middleware.
func NewAuthMiddleware(guard Guard) *AuthMiddleware {
	return &AuthMiddleware{
		guard:       guard,
		redirectTo:  "/login",
		abortOnFail: false,
	}
}

// SetRedirectTo sets the redirect URL for unauthenticated users.
func (m *AuthMiddleware) SetRedirectTo(url string) *AuthMiddleware {
	m.redirectTo = url
	return m
}

// SetAbortOnFail sets whether to abort with 401 instead of redirecting.
func (m *AuthMiddleware) SetAbortOnFail(abort bool) *AuthMiddleware {
	m.abortOnFail = abort
	return m
}

// Handle implements the middleware interface.
func (m *AuthMiddleware) Handle(w http.ResponseWriter, r *http.Request, next Next) {
	if m.guard == nil || m.guard.Guest() {
		if m.abortOnFail {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		http.Redirect(w, r, m.redirectTo, http.StatusFound)
		return
	}

	next(w, r)
}

// GuestMiddleware is a middleware that requires the user to be a guest.
type GuestMiddleware struct {
	guard      Guard
	redirectTo string
}

// NewGuestMiddleware creates a new guest middleware.
func NewGuestMiddleware(guard Guard) *GuestMiddleware {
	return &GuestMiddleware{
		guard:      guard,
		redirectTo: "/home",
	}
}

// SetRedirectTo sets the redirect URL for authenticated users.
func (m *GuestMiddleware) SetRedirectTo(url string) *GuestMiddleware {
	m.redirectTo = url
	return m
}

// Handle implements the middleware interface.
func (m *GuestMiddleware) Handle(w http.ResponseWriter, r *http.Request, next Next) {
	if m.guard != nil && m.guard.Check() {
		http.Redirect(w, r, m.redirectTo, http.StatusFound)
		return
	}

	next(w, r)
}

// Authenticate creates an authentication middleware using the default guard.
func Authenticate() *AuthMiddleware {
	return NewAuthMiddleware(defaultManager.DefaultGuard())
}

// RedirectIfAuthenticated creates a guest middleware using the default guard.
func RedirectIfAuthenticated() *GuestMiddleware {
	return NewGuestMiddleware(defaultManager.DefaultGuard())
}

// TokenAuthMiddleware is a middleware that authenticates using Bearer tokens.
type TokenAuthMiddleware struct {
	guard *TokenGuard
}

// NewTokenAuthMiddleware creates a new token authentication middleware.
func NewTokenAuthMiddleware(guard *TokenGuard) *TokenAuthMiddleware {
	return &TokenAuthMiddleware{
		guard: guard,
	}
}

// Handle implements the middleware interface.
func (m *TokenAuthMiddleware) Handle(w http.ResponseWriter, r *http.Request, next Next) {
	// Extract Bearer token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse Bearer token
	const prefix = "Bearer "
	if len(authHeader) < len(prefix) || authHeader[:len(prefix)] != prefix {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := authHeader[len(prefix):]
	if err := m.guard.SetToken(token); err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if m.guard.Guest() {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	next(w, r)
}
