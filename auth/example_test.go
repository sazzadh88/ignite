package auth_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/sazzad/ignite/auth"
)

// User represents a simple user model.
type User struct {
	ID       int
	Email    string
	Password string
}

// GetAuthIdentifier returns the user's unique identifier.
func (u *User) GetAuthIdentifier() any {
	return u.ID
}

// GetAuthPassword returns the user's hashed password.
func (u *User) GetAuthPassword() string {
	return u.Password
}

// SimpleSession is a basic session implementation.
type SimpleSession struct {
	data map[string]any
}

func (s *SimpleSession) Get(key string) (any, bool) {
	val, exists := s.data[key]
	return val, exists
}

func (s *SimpleSession) Put(key, value any) {
	s.data[key.(string)] = value
}

func (s *SimpleSession) Forget(key string) {
	delete(s.data, key)
}

func (s *SimpleSession) Flush() {
	s.data = make(map[string]any)
}

func ExampleSessionGuard() {
	// Create a user provider
	users := map[int]*User{
		1: {ID: 1, Email: "john@example.com", Password: "hashed_password"},
	}

	provider := auth.NewCallbackProvider(
		func(id any) (auth.Authenticatable, error) {
			user, exists := users[id.(int)]
			if !exists {
				return nil, auth.ErrUserNotFound
			}
			return user, nil
		},
		func(credentials map[string]string) (auth.Authenticatable, error) {
			email := credentials["email"]
			for _, user := range users {
				if user.Email == email {
					return user, nil
				}
			}
			return nil, auth.ErrUserNotFound
		},
		func(user auth.Authenticatable, credentials map[string]string) bool {
			password := credentials["password"]
			return user.GetAuthPassword() == password
		},
	)

	// Create session and guard
	session := &SimpleSession{data: make(map[string]any)}
	guard := auth.NewSessionGuard(provider, session)

	// Attempt login
	success := guard.Attempt(map[string]string{
		"email":    "john@example.com",
		"password": "hashed_password",
	})

	fmt.Println("Login successful:", success)
	fmt.Println("Authenticated:", guard.Check())
	fmt.Println("User ID:", guard.ID())

	// Logout
	guard.Logout()
	fmt.Println("After logout, guest:", guard.Guest())

	// Output:
	// Login successful: true
	// Authenticated: true
	// User ID: 1
	// After logout, guest: true
}

func ExampleTokenGuard() {
	// Create a user provider
	users := map[int]*User{
		1: {ID: 1, Email: "john@example.com", Password: "hashed_password"},
	}

	provider := auth.NewCallbackProvider(
		func(id any) (auth.Authenticatable, error) {
			user, exists := users[id.(int)]
			if !exists {
				return nil, auth.ErrUserNotFound
			}
			return user, nil
		},
		nil,
		nil,
	)

	// Create token storage and guard
	storage := auth.NewMemoryTokenStorage()
	guard := auth.NewTokenGuard(provider, storage)

	// Create a token
	user := users[1]
	token, err := guard.CreateToken(user, "api-token", []string{"read", "write"})
	if err != nil {
		panic(err)
	}

	fmt.Println("Token created:", token.Name)
	fmt.Println("Has abilities:", len(token.Abilities) > 0)

	// Authenticate with token
	err = guard.SetToken(token.PlainTextToken)
	fmt.Println("Authentication error:", err)
	fmt.Println("Authenticated:", guard.Check())

	// Output:
	// Token created: api-token
	// Has abilities: true
	// Authentication error: <nil>
	// Authenticated: true
}

func ExampleAuthMiddleware() {
	// Setup guard
	users := map[int]*User{
		1: {ID: 1, Email: "john@example.com", Password: "hashed_password"},
	}

	provider := auth.NewCallbackProvider(
		func(id any) (auth.Authenticatable, error) {
			user, exists := users[id.(int)]
			if !exists {
				return nil, auth.ErrUserNotFound
			}
			return user, nil
		},
		nil,
		nil,
	)

	session := &SimpleSession{data: make(map[string]any)}
	guard := auth.NewSessionGuard(provider, session)

	// Create middleware
	middleware := auth.NewAuthMiddleware(guard).SetAbortOnFail(true)

	// Test unauthenticated request
	req := httptest.NewRequest("GET", "/protected", nil)
	w := httptest.NewRecorder()

	middleware.Handle(w, req, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Unauthenticated status:", w.Code)

	// Login and test authenticated request
	guard.Login(users[1])
	req = httptest.NewRequest("GET", "/protected", nil)
	w = httptest.NewRecorder()

	middleware.Handle(w, req, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	fmt.Println("Authenticated status:", w.Code)

	// Output:
	// Unauthenticated status: 401
	// Authenticated status: 200
}

func ExampleManager() {
	// Create multiple guards
	users := map[int]*User{
		1: {ID: 1, Email: "john@example.com", Password: "hashed_password"},
	}

	provider := auth.NewCallbackProvider(
		func(id any) (auth.Authenticatable, error) {
			user, exists := users[id.(int)]
			if !exists {
				return nil, auth.ErrUserNotFound
			}
			return user, nil
		},
		func(credentials map[string]string) (auth.Authenticatable, error) {
			email := credentials["email"]
			for _, user := range users {
				if user.Email == email {
					return user, nil
				}
			}
			return nil, auth.ErrUserNotFound
		},
		func(user auth.Authenticatable, credentials map[string]string) bool {
			password := credentials["password"]
			return user.GetAuthPassword() == password
		},
	)

	// Session guard
	session := &SimpleSession{data: make(map[string]any)}
	sessionGuard := auth.NewSessionGuard(provider, session)

	// Token guard
	storage := auth.NewMemoryTokenStorage()
	tokenGuard := auth.NewTokenGuard(provider, storage)

	// Setup manager
	manager := auth.NewManager()
	manager.AddGuard("session", sessionGuard)
	manager.AddGuard("token", tokenGuard)
	manager.SetDefaultGuard("session")

	fmt.Println("Has session guard:", manager.HasGuard("session"))
	fmt.Println("Has token guard:", manager.HasGuard("token"))

	// Use default guard
	guard := manager.Guard("")
	fmt.Println("Default guard is guest:", guard.Guest())

	// Output:
	// Has session guard: true
	// Has token guard: true
	// Default guard is guest: true
}
