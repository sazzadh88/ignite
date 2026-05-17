// Package gate provides Laravel-inspired authorization functionality for Ignite.
//
// The gate package implements Gates and Policies for fine-grained access control
// in your application. It allows you to define authorization logic in a centralized
// location and check permissions throughout your application.
//
// # Basic Usage
//
// Create a new gate and define abilities:
//
//	g := gate.NewGate()
//	g.Define("edit-settings", func(user any, args ...any) bool {
//	    u := user.(*User)
//	    return u.IsAdmin
//	})
//
// Check authorization:
//
//	g.SetUser(currentUser)
//	if g.Allows("edit-settings") {
//	    // User can edit settings
//	}
//
// # Global Access Gate
//
// Use the global Access gate for package-level operations:
//
//	gate.Access.Define("delete-user", func(user any, args ...any) bool {
//	    u := user.(*User)
//	    return u.IsAdmin
//	})
//
//	gate.Access.SetUser(currentUser)
//	if gate.Access.Allows("delete-user") {
//	    // Perform deletion
//	}
//
// # Policies
//
// Policies group authorization logic for a particular model. Define a policy:
//
//	type PostPolicy struct{}
//
//	func (p *PostPolicy) View(user any, post any) bool {
//	    u := user.(*User)
//	    po := post.(*Post)
//	    return u.ID == po.AuthorID || u.IsAdmin
//	}
//
//	func (p *PostPolicy) Update(user any, post any) bool {
//	    u := user.(*User)
//	    po := post.(*Post)
//	    return u.ID == po.AuthorID
//	}
//
//	func (p *PostPolicy) Delete(user any, post any) bool {
//	    u := user.(*User)
//	    return u.IsAdmin
//	}
//
// Register the policy:
//
//	policy := &PostPolicy{}
//	g.Register(Post{}, policy)
//
// Use the policy (automatically resolved):
//
//	post := &Post{ID: 1, AuthorID: 5}
//	g.SetUser(currentUser)
//	if g.Allows("update", post) {
//	    // User can update this post
//	}
//
// # Before Hooks
//
// Run logic before all authorization checks. Useful for granting super admin access:
//
//	g.Before(func(user any, ability string) *bool {
//	    u := user.(*User)
//	    if u.IsAdmin {
//	        result := true
//	        return &result // Grant access immediately
//	    }
//	    return nil // Continue with normal checks
//	})
//
// # After Hooks
//
// Run logic after authorization checks to override or log results:
//
//	g.After(func(user any, ability string, result bool) *bool {
//	    // Log authorization attempts
//	    log.Printf("User %v attempted %s: %v", user, ability, result)
//
//	    // Optionally override
//	    if isBannedUser(user) {
//	        denied := false
//	        return &denied
//	    }
//	    return nil // Keep original result
//	})
//
// # Checking Multiple Abilities
//
// Check if user has any of the abilities:
//
//	if g.Any([]string{"edit-post", "delete-post"}) {
//	    // User can edit OR delete
//	}
//
// Check if user has all abilities:
//
//	if g.Check([]string{"view-post", "edit-post"}) {
//	    // User can view AND edit
//	}
//
// Check if user has none of the abilities:
//
//	if g.None([]string{"edit-post", "delete-post"}) {
//	    // User cannot edit or delete
//	}
//
// # Authorization for Specific Users
//
// Check authorization against a specific user (without setting global user):
//
//	admin := &User{ID: 1, IsAdmin: true}
//	if g.ForUser(admin).Allows("delete-user") {
//	    // This specific user can delete
//	}
//
// # Error Handling
//
// Use Authorize to get an error instead of a boolean:
//
//	if err := g.Authorize("edit-settings"); err != nil {
//	    return err // Returns "this action is unauthorized"
//	}
//
// # Middleware
//
// Protect HTTP routes with authorization middleware:
//
//	g.Define("view-admin", func(user any, args ...any) bool {
//	    u := user.(*User)
//	    return u.IsAdmin
//	})
//
//	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	    w.Write([]byte("Admin Dashboard"))
//	})
//
//	// Wrap with authorization
//	protected := gate.CanFunc(g, "view-admin")(handler)
//	http.Handle("/admin", protected)
//
// Using the global Access gate:
//
//	protected := gate.Can("view-admin")(handler)
//
// # Authorization Responses
//
// Create structured authorization responses:
//
//	// Allow access
//	resp := gate.Allow("Access granted")
//	if resp.Allowed {
//	    // Process request
//	}
//
//	// Deny access
//	resp := gate.Deny("You must be an admin")
//	// resp.Code is 403
//
//	// Deny with custom status
//	resp := gate.DenyWithStatus(401, "Authentication required")
//
//	// Deny as not found (hide resource existence)
//	resp := gate.DenyAsNotFound()
//	// resp.Code is 404
//
// # Integration Example
//
// Complete example with HTTP handler:
//
//	type User struct {
//	    ID      int
//	    IsAdmin bool
//	}
//
//	func main() {
//	    g := gate.NewGate()
//
//	    // Define abilities
//	    g.Define("view-admin", func(user any, args ...any) bool {
//	        u, ok := user.(*User)
//	        return ok && u.IsAdmin
//	    })
//
//	    // Create handler
//	    adminHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//	        // Get user from session/JWT
//	        user := getUserFromRequest(r)
//	        g.SetUser(user)
//
//	        if g.Denies("view-admin") {
//	            http.Error(w, "Forbidden", 403)
//	            return
//	        }
//
//	        w.Write([]byte("Admin Dashboard"))
//	    })
//
//	    // Or use middleware
//	    protected := gate.CanFunc(g, "view-admin")(adminHandler)
//	    http.Handle("/admin", protected)
//	}
//
// # Thread Safety
//
// The Gate is safe for concurrent use. All internal maps are protected by a sync.RWMutex.
// You can safely define abilities and register policies from multiple goroutines.
//
// # Policy Method Resolution
//
// When checking authorization with a model argument, the gate automatically looks for
// a registered policy for that model type. It calls the method matching the ability name
// (capitalized). Common policy methods include:
//
//   - ViewAny(user any) bool
//   - View(user any, model any) bool
//   - Create(user any) bool
//   - Update(user any, model any) bool
//   - Delete(user any, model any) bool
//   - Restore(user any, model any) bool
//   - ForceDelete(user any, model any) bool
//
// Example policy method resolution:
//
//	g.Allows("view", post)     // Calls policy.View(user, post)
//	g.Allows("update", post)   // Calls policy.Update(user, post)
//	g.Allows("delete", post)   // Calls policy.Delete(user, post)
package gate
