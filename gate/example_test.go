package gate_test

import (
	"fmt"
	"net/http"

	"github.com/sazzadh88/ignite/gate"
)

// User represents an authenticated user.
type User struct {
	ID      int
	IsAdmin bool
}

// Post represents a blog post.
type Post struct {
	ID       int
	AuthorID int
}

// PostPolicy defines authorization rules for posts.
type PostPolicy struct{}

func (p *PostPolicy) View(user any, post any) bool {
	u := user.(*User)
	po := post.(*Post)
	return u.ID == po.AuthorID || u.IsAdmin
}

func (p *PostPolicy) Update(user any, post any) bool {
	u := user.(*User)
	po := post.(*Post)
	return u.ID == po.AuthorID
}

func (p *PostPolicy) Delete(user any, post any) bool {
	u := user.(*User)
	return u.IsAdmin
}

func ExampleGate_Define() {
	g := gate.NewGate()

	// Define a simple ability
	g.Define("edit-settings", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	admin := &User{ID: 1, IsAdmin: true}
	g.SetUser(admin)

	if g.Allows("edit-settings") {
		fmt.Println("Admin can edit settings")
	}

	// Output: Admin can edit settings
}

func ExampleGate_Before() {
	g := gate.NewGate()

	// Grant all permissions to admins
	g.Before(func(user any, ability string) *bool {
		u, ok := user.(*User)
		if ok && u.IsAdmin {
			result := true
			return &result
		}
		return nil
	})

	g.Define("delete-user", func(user any, args ...any) bool {
		return false // Normally deny
	})

	admin := &User{ID: 1, IsAdmin: true}
	g.SetUser(admin)

	if g.Allows("delete-user") {
		fmt.Println("Admin bypasses all checks")
	}

	// Output: Admin bypasses all checks
}

func ExampleGate_Register() {
	g := gate.NewGate()

	// Register policy for Post model
	policy := &PostPolicy{}
	g.Register(Post{}, policy)

	author := &User{ID: 1, IsAdmin: false}
	post := &Post{ID: 100, AuthorID: 1}

	g.SetUser(author)

	if g.Allows("update", post) {
		fmt.Println("Author can update their post")
	}

	if g.Denies("delete", post) {
		fmt.Println("Author cannot delete post")
	}

	// Output:
	// Author can update their post
	// Author cannot delete post
}

func ExampleGate_ForUser() {
	g := gate.NewGate()

	g.Define("publish-post", func(user any, args ...any) bool {
		u := user.(*User)
		return u.ID > 0
	})

	user1 := &User{ID: 1, IsAdmin: false}
	user2 := &User{ID: 2, IsAdmin: false}

	// Check authorization for specific users
	if g.ForUser(user1).Allows("publish-post") {
		fmt.Println("User 1 can publish")
	}

	if g.ForUser(user2).Allows("publish-post") {
		fmt.Println("User 2 can publish")
	}

	// Output:
	// User 1 can publish
	// User 2 can publish
}

func ExampleGate_Any() {
	g := gate.NewGate()

	g.Define("view-post", func(user any, args ...any) bool {
		return true
	})

	g.Define("edit-post", func(user any, args ...any) bool {
		return false
	})

	user := &User{ID: 1, IsAdmin: false}
	g.SetUser(user)

	// Check if user has any of the abilities
	if g.Any([]string{"view-post", "edit-post"}) {
		fmt.Println("User can view or edit")
	}

	// Output: User can view or edit
}

func ExampleGate_Check() {
	g := gate.NewGate()

	g.Define("view-dashboard", func(user any, args ...any) bool {
		return true
	})

	g.Define("view-analytics", func(user any, args ...any) bool {
		return true
	})

	user := &User{ID: 1, IsAdmin: false}
	g.SetUser(user)

	// Check if user has all abilities
	if g.Check([]string{"view-dashboard", "view-analytics"}) {
		fmt.Println("User can access full dashboard")
	}

	// Output: User can access full dashboard
}

func ExampleCan() {
	g := gate.NewGate()

	g.Define("view-admin", func(user any, args ...any) bool {
		u := user.(*User)
		return u.IsAdmin
	})

	admin := &User{ID: 1, IsAdmin: true}
	g.SetUser(admin)

	// Create a protected handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Admin Dashboard"))
	})

	// Wrap with authorization middleware
	protected := gate.CanFunc(g, "view-admin")(handler)

	// In a real application, you would use this with your router
	_ = protected

	fmt.Println("Middleware configured")

	// Output: Middleware configured
}

func ExampleResponse() {
	// Allow with message
	allowResp := gate.Allow("Access granted")
	fmt.Printf("Allowed: %v, Message: %s\n", allowResp.Allowed, allowResp.Message)

	// Deny with default message
	denyResp := gate.Deny()
	fmt.Printf("Allowed: %v, Code: %d\n", denyResp.Allowed, denyResp.Code)

	// Deny as not found
	notFoundResp := gate.DenyAsNotFound()
	fmt.Printf("Code: %d\n", notFoundResp.Code)

	// Output:
	// Allowed: true, Message: Access granted
	// Allowed: false, Code: 403
	// Code: 404
}
