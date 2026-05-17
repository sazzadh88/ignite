package gate

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Mock user types for testing
type User struct {
	ID      int
	IsAdmin bool
}

type Post struct {
	ID       int
	AuthorID int
}

type PostPolicy struct{}

func (p *PostPolicy) View(user any, post any) bool {
	u, ok := user.(*User)
	if !ok {
		return false
	}

	po, ok := post.(*Post)
	if !ok {
		return false
	}

	return u.ID == po.AuthorID || u.IsAdmin
}

func (p *PostPolicy) Update(user any, post any) bool {
	u, ok := user.(*User)
	if !ok {
		return false
	}

	po, ok := post.(*Post)
	if !ok {
		return false
	}

	return u.ID == po.AuthorID
}

func (p *PostPolicy) Create(user any) bool {
	u, ok := user.(*User)
	if !ok {
		return false
	}
	return u.ID > 0
}

func TestDefineAndCheckAbility(t *testing.T) {
	g := NewGate()

	g.Define("edit-settings", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	if !g.Has("edit-settings") {
		t.Error("Expected ability to be defined")
	}

	if g.Has("non-existent") {
		t.Error("Expected ability to not be defined")
	}
}

func TestAllowsAndDenies(t *testing.T) {
	g := NewGate()

	g.Define("edit-settings", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	adminUser := &User{ID: 1, IsAdmin: true}
	normalUser := &User{ID: 2, IsAdmin: false}

	g.SetUser(adminUser)
	if !g.Allows("edit-settings") {
		t.Error("Expected admin to be allowed")
	}
	if g.Denies("edit-settings") {
		t.Error("Expected admin to not be denied")
	}

	g.SetUser(normalUser)
	if g.Allows("edit-settings") {
		t.Error("Expected normal user to be denied")
	}
	if !g.Denies("edit-settings") {
		t.Error("Expected normal user to be denied")
	}
}

func TestBeforeHook(t *testing.T) {
	g := NewGate()

	g.Define("edit-post", func(user any, args ...any) bool {
		return false // Should normally deny
	})

	// Before hook grants access to admins
	g.Before(func(user any, ability string) *bool {
		u, ok := user.(*User)
		if ok && u.IsAdmin {
			result := true
			return &result
		}
		return nil
	})

	adminUser := &User{ID: 1, IsAdmin: true}
	normalUser := &User{ID: 2, IsAdmin: false}

	g.SetUser(adminUser)
	if !g.Allows("edit-post") {
		t.Error("Expected before hook to grant access to admin")
	}

	g.SetUser(normalUser)
	if g.Allows("edit-post") {
		t.Error("Expected normal user to be denied")
	}
}

func TestAfterHook(t *testing.T) {
	g := NewGate()

	g.Define("view-post", func(user any, args ...any) bool {
		return true // Should normally allow
	})

	// After hook denies access to banned users
	g.After(func(user any, ability string, result bool) *bool {
		u, ok := user.(*User)
		if ok && u.ID == 999 { // Banned user
			denied := false
			return &denied
		}
		return nil
	})

	normalUser := &User{ID: 1, IsAdmin: false}
	bannedUser := &User{ID: 999, IsAdmin: false}

	g.SetUser(normalUser)
	if !g.Allows("view-post") {
		t.Error("Expected normal user to be allowed")
	}

	g.SetUser(bannedUser)
	if g.Allows("view-post") {
		t.Error("Expected after hook to deny banned user")
	}
}

func TestPolicyRegistration(t *testing.T) {
	g := NewGate()

	policy := &PostPolicy{}
	g.Register(Post{}, policy)

	author := &User{ID: 1, IsAdmin: false}
	otherUser := &User{ID: 2, IsAdmin: false}
	post := &Post{ID: 1, AuthorID: 1}

	g.SetUser(author)
	if !g.Allows("view", post) {
		t.Error("Expected author to view their own post")
	}
	if !g.Allows("update", post) {
		t.Error("Expected author to update their own post")
	}

	g.SetUser(otherUser)
	if g.Allows("update", post) {
		t.Error("Expected other user to not update post")
	}
}

func TestPolicyWithAdmin(t *testing.T) {
	g := NewGate()

	policy := &PostPolicy{}
	g.Register(Post{}, policy)

	admin := &User{ID: 3, IsAdmin: true}
	otherUser := &User{ID: 2, IsAdmin: false}
	post := &Post{ID: 1, AuthorID: 1}

	g.SetUser(admin)
	if !g.Allows("view", post) {
		t.Error("Expected admin to view any post")
	}

	g.SetUser(otherUser)
	if g.Allows("view", post) {
		t.Error("Expected non-admin, non-author to not view post")
	}
}

func TestAnyAbility(t *testing.T) {
	g := NewGate()

	g.Define("edit-post", func(user any, args ...any) bool {
		return false
	})

	g.Define("delete-post", func(user any, args ...any) bool {
		return true
	})

	user := &User{ID: 1, IsAdmin: false}
	g.SetUser(user)

	if !g.Any([]string{"edit-post", "delete-post"}) {
		t.Error("Expected Any to return true when at least one ability passes")
	}

	if g.Any([]string{"edit-post", "non-existent"}) {
		t.Error("Expected Any to return false when no abilities pass")
	}
}

func TestNoneAbility(t *testing.T) {
	g := NewGate()

	g.Define("edit-post", func(user any, args ...any) bool {
		return false
	})

	g.Define("delete-post", func(user any, args ...any) bool {
		return false
	})

	user := &User{ID: 1, IsAdmin: false}
	g.SetUser(user)

	if !g.None([]string{"edit-post", "delete-post"}) {
		t.Error("Expected None to return true when all abilities fail")
	}

	g.Define("view-post", func(user any, args ...any) bool {
		return true
	})

	if g.None([]string{"edit-post", "view-post"}) {
		t.Error("Expected None to return false when any ability passes")
	}
}

func TestCheckAbility(t *testing.T) {
	g := NewGate()

	g.Define("view-post", func(user any, args ...any) bool {
		return true
	})

	g.Define("edit-post", func(user any, args ...any) bool {
		return true
	})

	user := &User{ID: 1, IsAdmin: false}
	g.SetUser(user)

	if !g.Check([]string{"view-post", "edit-post"}) {
		t.Error("Expected Check to return true when all abilities pass")
	}

	g.Define("delete-post", func(user any, args ...any) bool {
		return false
	})

	if g.Check([]string{"view-post", "delete-post"}) {
		t.Error("Expected Check to return false when any ability fails")
	}
}

func TestAuthorize(t *testing.T) {
	g := NewGate()

	g.Define("edit-settings", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	adminUser := &User{ID: 1, IsAdmin: true}
	normalUser := &User{ID: 2, IsAdmin: false}

	g.SetUser(adminUser)
	if err := g.Authorize("edit-settings"); err != nil {
		t.Errorf("Expected no error for admin, got: %v", err)
	}

	g.SetUser(normalUser)
	if err := g.Authorize("edit-settings"); err == nil {
		t.Error("Expected error for normal user")
	}
}

func TestForUser(t *testing.T) {
	g := NewGate()

	g.Define("edit-post", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	adminUser := &User{ID: 1, IsAdmin: true}
	normalUser := &User{ID: 2, IsAdmin: false}

	// Check against specific user
	if !g.ForUser(adminUser).Allows("edit-post") {
		t.Error("Expected admin to be allowed")
	}

	if g.ForUser(normalUser).Allows("edit-post") {
		t.Error("Expected normal user to be denied")
	}

	// Test Denies
	if g.ForUser(adminUser).Denies("edit-post") {
		t.Error("Expected admin to not be denied")
	}

	// Test Authorize
	if err := g.ForUser(adminUser).Authorize("edit-post"); err != nil {
		t.Error("Expected no error for admin")
	}

	if err := g.ForUser(normalUser).Authorize("edit-post"); err == nil {
		t.Error("Expected error for normal user")
	}
}

func TestForUserWithMultipleAbilities(t *testing.T) {
	g := NewGate()

	g.Define("view-post", func(user any, args ...any) bool {
		return true
	})

	g.Define("edit-post", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	g.Define("delete-post", func(user any, args ...any) bool {
		return false
	})

	adminUser := &User{ID: 1, IsAdmin: true}

	userGate := g.ForUser(adminUser)

	if !userGate.Any([]string{"edit-post", "delete-post"}) {
		t.Error("Expected Any to return true")
	}

	if userGate.None([]string{"view-post", "edit-post"}) {
		t.Error("Expected None to return false")
	}

	if !userGate.Check([]string{"view-post", "edit-post"}) {
		t.Error("Expected Check to return true")
	}
}

func TestResponse(t *testing.T) {
	allow := Allow()
	if !allow.Allowed {
		t.Error("Expected Allow to set Allowed to true")
	}
	if allow.Code != http.StatusOK {
		t.Error("Expected Allow to set Code to 200")
	}

	allowWithMsg := Allow("Custom message")
	if allowWithMsg.Message != "Custom message" {
		t.Error("Expected custom message")
	}

	deny := Deny()
	if deny.Allowed {
		t.Error("Expected Deny to set Allowed to false")
	}
	if deny.Code != http.StatusForbidden {
		t.Error("Expected Deny to set Code to 403")
	}
	if deny.Message != "This action is unauthorized." {
		t.Error("Expected default deny message")
	}

	denyWithMsg := Deny("Custom deny message")
	if denyWithMsg.Message != "Custom deny message" {
		t.Error("Expected custom deny message")
	}

	denyWithStatus := DenyWithStatus(http.StatusUnauthorized, "Unauthorized")
	if denyWithStatus.Code != http.StatusUnauthorized {
		t.Error("Expected custom status code")
	}
	if denyWithStatus.Message != "Unauthorized" {
		t.Error("Expected custom message")
	}

	notFound := DenyAsNotFound()
	if notFound.Code != http.StatusNotFound {
		t.Error("Expected 404 status")
	}
	if notFound.Message != "Not found." {
		t.Error("Expected not found message")
	}

	notFoundWithMsg := DenyAsNotFound("Resource not found")
	if notFoundWithMsg.Message != "Resource not found" {
		t.Error("Expected custom not found message")
	}
}

func TestCanMiddleware(t *testing.T) {
	g := NewGate()

	g.Define("view-dashboard", func(user any, args ...any) bool {
		u, ok := user.(*User)
		return ok && u.IsAdmin
	})

	adminUser := &User{ID: 1, IsAdmin: true}
	normalUser := &User{ID: 2, IsAdmin: false}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Dashboard"))
	})

	// Test with admin user
	g.SetUser(adminUser)
	middleware := CanFunc(g, "view-dashboard")
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest("GET", "/dashboard", nil)
	w := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200 for admin, got %d", w.Code)
	}

	// Test with normal user
	g.SetUser(normalUser)
	req = httptest.NewRequest("GET", "/dashboard", nil)
	w = httptest.NewRecorder()
	wrappedHandler.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Errorf("Expected status 403 for normal user, got %d", w.Code)
	}
}

func TestGlobalAccessGate(t *testing.T) {
	// Reset the global gate for testing
	Access = NewGate()

	Access.Define("test-ability", func(user any, args ...any) bool {
		return true
	})

	if !Access.Has("test-ability") {
		t.Error("Expected global Access gate to have ability")
	}

	user := &User{ID: 1, IsAdmin: false}
	Access.SetUser(user)

	if !Access.Allows("test-ability") {
		t.Error("Expected global Access gate to allow ability")
	}
}

func TestRegisterPolicyGlobal(t *testing.T) {
	// Reset the global gate for testing
	Access = NewGate()

	policy := &PostPolicy{}
	RegisterPolicy(Post{}, policy)

	author := &User{ID: 1, IsAdmin: false}
	post := &Post{ID: 1, AuthorID: 1}

	Access.SetUser(author)
	if !Access.Allows("view", post) {
		t.Error("Expected global RegisterPolicy to work")
	}
}
