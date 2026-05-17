package session_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/sazzadh88/ignite/session"
)

// Example demonstrates basic session usage.
func Example() {
	// Create session manager with default configuration
	config := session.DefaultConfig()
	mgr, err := session.NewManager(config)
	if err != nil {
		panic(err)
	}

	// Create HTTP handler with session middleware
	handler := session.StartSession(mgr)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get session from request
		sess := session.FromRequest(r)

		// Store data
		sess.Put("user_id", 123)
		sess.Put("username", "john")

		// Flash data (available only in next request)
		sess.Flash("message", "Welcome!")

		// Retrieve data
		userID := sess.GetInt("user_id")
		username := sess.GetString("username")

		fmt.Fprintf(w, "User: %s (ID: %d)", username, userID)
	}))

	// Test the handler
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	fmt.Println(w.Body.String())
	// Output: User: john (ID: 123)
}

// ExampleSession_Flash demonstrates flash data usage.
func ExampleSession_Flash() {
	store := session.NewMemoryStore()
	sess := session.NewSession(store, "")

	// Start session
	sess.Start()

	// Set flash data
	sess.Flash("success", "Record saved!")
	sess.Save()

	// Flash data will be available in the next request
	// and then automatically removed
}

// ExampleSession_Increment demonstrates counter operations.
func ExampleSession_Increment() {
	store := session.NewMemoryStore()
	sess := session.NewSession(store, "")

	// Increment page views
	sess.Increment("page_views")
	sess.Increment("page_views")
	sess.Increment("page_views")

	views := sess.GetInt("page_views")
	fmt.Println(views)
	// Output: 3
}

// ExampleSession_Push demonstrates array operations.
func ExampleSession_Push() {
	store := session.NewMemoryStore()
	sess := session.NewSession(store, "")

	// Build a list
	sess.Push("cart", "item1")
	sess.Push("cart", "item2")
	sess.Push("cart", "item3")

	cart := sess.Get("cart")
	items := cart.([]any)
	fmt.Println(len(items))
	// Output: 3
}

// ExampleNewFileStore demonstrates file-based session storage.
func ExampleNewFileStore() {
	// Create file store
	store, err := session.NewFileStore("/tmp/sessions")
	if err != nil {
		panic(err)
	}

	// Use with session manager
	config := session.SessionConfig{
		Driver:     "file",
		CookieName: "app_session",
		Lifetime:   120,
		Files:      "/tmp/sessions",
	}

	manager, err := session.NewManager(config)
	if err != nil {
		panic(err)
	}

	_ = store
	_ = manager
}
