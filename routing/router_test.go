package routing

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Test helper: create a simple handler
func simpleHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func TestRouterBasicRegistration(t *testing.T) {
	router := NewRouter()

	route := router.Get("/users", HandlerFunc(simpleHandler))

	if route.Method() != "GET" {
		t.Errorf("Expected method GET, got %s", route.Method())
	}

	if route.Path() != "/users" {
		t.Errorf("Expected path /users, got %s", route.Path())
	}

	if len(router.Routes()) != 1 {
		t.Errorf("Expected 1 route, got %d", len(router.Routes()))
	}
}

func TestRouterAllMethods(t *testing.T) {
	router := NewRouter()

	router.Get("/get", HandlerFunc(simpleHandler))
	router.Post("/post", HandlerFunc(simpleHandler))
	router.Put("/put", HandlerFunc(simpleHandler))
	router.Patch("/patch", HandlerFunc(simpleHandler))
	router.Delete("/delete", HandlerFunc(simpleHandler))
	router.Options("/options", HandlerFunc(simpleHandler))

	if len(router.Routes()) != 6 {
		t.Errorf("Expected 6 routes, got %d", len(router.Routes()))
	}

	routes := router.Routes()
	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}

	for i, route := range routes {
		if route.Method() != expectedMethods[i] {
			t.Errorf("Expected method %s, got %s", expectedMethods[i], route.Method())
		}
	}
}

func TestRouterAny(t *testing.T) {
	router := NewRouter()

	routes := router.Any("/any", HandlerFunc(simpleHandler))

	if len(routes) != 7 {
		t.Errorf("Expected 7 routes for Any, got %d", len(routes))
	}

	expectedMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	for i, route := range routes {
		if route.Method() != expectedMethods[i] {
			t.Errorf("Expected method %s, got %s", expectedMethods[i], route.Method())
		}
	}
}

func TestRouterMatch(t *testing.T) {
	router := NewRouter()

	routes := router.MatchMethods([]string{"GET", "POST"}, "/match", HandlerFunc(simpleHandler))

	if len(routes) != 2 {
		t.Errorf("Expected 2 routes for Match, got %d", len(routes))
	}

	if routes[0].Method() != "GET" || routes[1].Method() != "POST" {
		t.Errorf("Expected GET and POST methods")
	}
}

func TestRouteNaming(t *testing.T) {
	router := NewRouter()

	router.Get("/users/{id}", HandlerFunc(simpleHandler)).Name("user.show")

	url := router.URL("user.show")
	if url != "/users/{id}" {
		t.Errorf("Expected /users/{id}, got %s", url)
	}

	urlWithParams := router.URLWith("user.show", map[string]string{"id": "123"})
	if urlWithParams != "/users/123" {
		t.Errorf("Expected /users/123, got %s", urlWithParams)
	}
}

func TestRouteMiddleware(t *testing.T) {
	router := NewRouter()

	route := router.Get("/protected", HandlerFunc(simpleHandler)).Middleware("auth", "verified")

	middlewares := route.Middlewares()
	if len(middlewares) != 2 {
		t.Errorf("Expected 2 middlewares, got %d", len(middlewares))
	}

	if middlewares[0] != "auth" || middlewares[1] != "verified" {
		t.Errorf("Middleware names don't match")
	}
}

func TestRouteWhere(t *testing.T) {
	router := NewRouter()

	route := router.Get("/users/{id}", HandlerFunc(simpleHandler)).Where("id", `\d+`)

	constraints := route.Constraints()
	if len(constraints) != 1 {
		t.Errorf("Expected 1 constraint, got %d", len(constraints))
	}

	if _, exists := constraints["id"]; !exists {
		t.Errorf("Expected 'id' constraint to exist")
	}
}

func TestRouteGroup(t *testing.T) {
	router := NewRouter()

	router.Group(func(r *Router) {
		r.Get("/dashboard", HandlerFunc(simpleHandler))
		r.Get("/settings", HandlerFunc(simpleHandler))
	}).Prefix("/admin").Middleware("auth")

	routes := router.Routes()
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes, got %d", len(routes))
	}

	for _, route := range routes {
		// Check prefix
		if route.Path() != "/admin/dashboard" && route.Path() != "/admin/settings" {
			t.Errorf("Expected path with /admin prefix, got %s", route.Path())
		}

		// Check middleware
		middlewares := route.Middlewares()
		if len(middlewares) != 1 || middlewares[0] != "auth" {
			t.Errorf("Expected 'auth' middleware")
		}
	}
}

func TestNestedRouteGroups(t *testing.T) {
	router := NewRouter()

	router.Group(func(r *Router) {
		r.Group(func(r2 *Router) {
			r2.Get("/profile", HandlerFunc(simpleHandler))
		}).Prefix("/user")
	}).Prefix("/admin").Middleware("auth")

	routes := router.Routes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.Path() != "/admin/user/profile" {
		t.Errorf("Expected /admin/user/profile, got %s", route.Path())
	}

	middlewares := route.Middlewares()
	if len(middlewares) != 1 || middlewares[0] != "auth" {
		t.Errorf("Expected 'auth' middleware")
	}
}

func TestRouteMatching(t *testing.T) {
	router := NewRouter()

	router.Get("/users", HandlerFunc(simpleHandler))
	router.Get("/users/{id}", HandlerFunc(simpleHandler))
	router.Get("/users/{id}/posts", HandlerFunc(simpleHandler))

	tests := []struct {
		method      string
		path        string
		shouldMatch bool
		params      map[string]string
	}{
		{"GET", "/users", true, map[string]string{}},
		{"GET", "/users/123", true, map[string]string{"id": "123"}},
		{"GET", "/users/123/posts", true, map[string]string{"id": "123"}},
		{"POST", "/users", false, nil},
		{"GET", "/nonexistent", false, nil},
		{"GET", "/users/123/comments", false, nil},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(tt.method, tt.path, nil)
		route, params, matched := router.MatchRequest(req)

		if matched != tt.shouldMatch {
			t.Errorf("Path %s %s: expected match=%v, got %v", tt.method, tt.path, tt.shouldMatch, matched)
		}

		if matched && route == nil {
			t.Errorf("Path %s %s: matched but route is nil", tt.method, tt.path)
		}

		if matched && tt.params != nil {
			for key, expectedVal := range tt.params {
				if params[key] != expectedVal {
					t.Errorf("Path %s: param %s expected %s, got %s", tt.path, key, expectedVal, params[key])
				}
			}
		}
	}
}

func TestRouteConstraints(t *testing.T) {
	router := NewRouter()

	router.Get("/users/{id}", HandlerFunc(simpleHandler)).Where("id", `\d+`)

	tests := []struct {
		path        string
		shouldMatch bool
	}{
		{"/users/123", true},
		{"/users/abc", false},
		{"/users/12abc", false},
	}

	for _, tt := range tests {
		req := httptest.NewRequest("GET", tt.path, nil)
		_, _, matched := router.MatchRequest(req)

		if matched != tt.shouldMatch {
			t.Errorf("Path %s: expected match=%v, got %v", tt.path, tt.shouldMatch, matched)
		}
	}
}

func TestRedirect(t *testing.T) {
	router := NewRouter()

	router.Redirect("/old", "/new", http.StatusFound)

	req := httptest.NewRequest("GET", "/old", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Errorf("Expected status %d, got %d", http.StatusFound, w.Code)
	}

	location := w.Header().Get("Location")
	if location != "/new" {
		t.Errorf("Expected Location /new, got %s", location)
	}
}

func TestPermanentRedirect(t *testing.T) {
	router := NewRouter()

	router.PermanentRedirect("/old", "/new")

	req := httptest.NewRequest("GET", "/old", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusMovedPermanently {
		t.Errorf("Expected status %d, got %d", http.StatusMovedPermanently, w.Code)
	}
}

func TestFallback(t *testing.T) {
	router := NewRouter()

	fallbackCalled := false
	router.Fallback(HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fallbackCalled = true
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("Custom 404"))
	}))

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if !fallbackCalled {
		t.Error("Fallback handler was not called")
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestResourceRoutes(t *testing.T) {
	router := NewRouter()

	router.Resource("posts", HandlerFunc(simpleHandler))

	routes := router.Routes()
	if len(routes) != 7 {
		t.Errorf("Expected 7 resource routes, got %d", len(routes))
	}

	expectedRoutes := []struct {
		method string
		path   string
		name   string
	}{
		{"GET", "/posts", "posts.index"},
		{"GET", "/posts/create", "posts.create"},
		{"POST", "/posts", "posts.store"},
		{"GET", "/posts/{id}", "posts.show"},
		{"GET", "/posts/{id}/edit", "posts.edit"},
		{"PUT", "/posts/{id}", "posts.update"},
		{"DELETE", "/posts/{id}", "posts.destroy"},
	}

	for i, expected := range expectedRoutes {
		route := routes[i]
		if route.Method() != expected.method {
			t.Errorf("Route %d: expected method %s, got %s", i, expected.method, route.Method())
		}
		if route.Path() != expected.path {
			t.Errorf("Route %d: expected path %s, got %s", i, expected.path, route.Path())
		}
	}
}

func TestApiResourceRoutes(t *testing.T) {
	router := NewRouter()

	router.ApiResource("posts", HandlerFunc(simpleHandler))

	routes := router.Routes()
	if len(routes) != 5 {
		t.Errorf("Expected 5 API resource routes, got %d", len(routes))
	}

	expectedRoutes := []struct {
		method string
		path   string
	}{
		{"GET", "/posts"},
		{"POST", "/posts"},
		{"GET", "/posts/{id}"},
		{"PUT", "/posts/{id}"},
		{"DELETE", "/posts/{id}"},
	}

	for i, expected := range expectedRoutes {
		route := routes[i]
		if route.Method() != expected.method {
			t.Errorf("Route %d: expected method %s, got %s", i, expected.method, route.Method())
		}
		if route.Path() != expected.path {
			t.Errorf("Route %d: expected path %s, got %s", i, expected.path, route.Path())
		}
	}
}

func TestResourceOnly(t *testing.T) {
	router := NewRouter()

	router.Resource("posts", HandlerFunc(simpleHandler)).Only("index", "show")

	routes := router.Routes()
	if len(routes) != 2 {
		t.Errorf("Expected 2 routes (index, show), got %d", len(routes))
	}
}

func TestResourceExcept(t *testing.T) {
	router := NewRouter()

	router.Resource("posts", HandlerFunc(simpleHandler)).Except("destroy")

	routes := router.Routes()
	if len(routes) != 6 {
		t.Errorf("Expected 6 routes (all except destroy), got %d", len(routes))
	}

	// Verify destroy is not present
	for _, route := range routes {
		if route.Method() == "DELETE" {
			t.Error("Destroy route should not be registered")
		}
	}
}

func TestDomainRouting(t *testing.T) {
	router := NewRouter()

	router.Domain("api.example.com").Group(func(r *Router) {
		r.Get("/users", HandlerFunc(simpleHandler))
	})

	routes := router.Routes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	route := routes[0]
	if route.GetDomain() != "api.example.com" {
		t.Errorf("Expected domain api.example.com, got %s", route.GetDomain())
	}

	// Test domain matching
	req := httptest.NewRequest("GET", "http://api.example.com/users", nil)
	_, _, matched := router.MatchRequest(req)
	if !matched {
		t.Error("Expected route to match api.example.com domain")
	}

	// Test non-matching domain
	req2 := httptest.NewRequest("GET", "http://example.com/users", nil)
	_, _, matched2 := router.MatchRequest(req2)
	if matched2 {
		t.Error("Expected route not to match example.com domain")
	}
}

func TestDomainWithParameters(t *testing.T) {
	router := NewRouter()

	router.Domain("{account}.example.com").Group(func(r *Router) {
		r.Get("/dashboard", HandlerFunc(simpleHandler))
	})

	routes := router.Routes()
	if len(routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(routes))
	}

	// Test matching with parameter
	req := httptest.NewRequest("GET", "http://tenant1.example.com/dashboard", nil)
	_, _, matched := router.MatchRequest(req)
	if !matched {
		t.Error("Expected route to match tenant1.example.com")
	}

	req2 := httptest.NewRequest("GET", "http://tenant2.example.com/dashboard", nil)
	_, _, matched2 := router.MatchRequest(req2)
	if !matched2 {
		t.Error("Expected route to match tenant2.example.com")
	}
}

func TestServeHTTP(t *testing.T) {
	router := NewRouter()

	router.Get("/test", HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Test Response"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, w.Code)
	}

	body := w.Body.String()
	if body != "Test Response" {
		t.Errorf("Expected body 'Test Response', got '%s'", body)
	}
}

func TestServeHTTP404(t *testing.T) {
	router := NewRouter()

	req := httptest.NewRequest("GET", "/nonexistent", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestConcurrentRouteRegistration(t *testing.T) {
	router := NewRouter()

	done := make(chan bool)

	// Register routes concurrently
	for i := 0; i < 100; i++ {
		go func(n int) {
			router.Get("/test", HandlerFunc(simpleHandler))
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}

	routes := router.Routes()
	if len(routes) != 100 {
		t.Errorf("Expected 100 routes, got %d", len(routes))
	}
}

func TestConcurrentRouteMatching(t *testing.T) {
	router := NewRouter()

	router.Get("/users/{id}", HandlerFunc(simpleHandler))

	done := make(chan bool)

	// Match routes concurrently
	for i := 0; i < 100; i++ {
		go func(n int) {
			req := httptest.NewRequest("GET", "/users/123", nil)
			_, _, matched := router.MatchRequest(req)
			if !matched {
				t.Error("Expected route to match")
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 100; i++ {
		<-done
	}
}
