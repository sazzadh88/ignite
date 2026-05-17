# Routing Package

Laravel-inspired HTTP routing for GoFrame.

## Features

- **HTTP verb methods**: Get, Post, Put, Patch, Delete, Options, Any, Match
- **Route parameters**: `/users/{id}` with regex constraints
- **Named routes**: URL generation with parameters
- **Route groups**: Shared prefix, middleware, and domain
- **RESTful resources**: Auto-generate CRUD routes
- **Middleware**: Per-route or per-group
- **Domain routing**: Multi-tenant support with subdomains
- **Redirects**: Permanent and temporary
- **Fallback routes**: Custom 404 handling
- **Thread-safe**: Concurrent route registration and matching

## Quick Start

```go
package main

import (
    "net/http"
    "github.com/sazzad/goframe/routing"
)

func main() {
    router := routing.NewRouter()

    // Basic routes
    router.Get("/", homeHandler)
    router.Post("/users", createUserHandler)
    router.Get("/users/{id}", showUserHandler).Where("id", `\d+`)

    // Named routes
    router.Get("/profile", profileHandler).Name("profile")
    url := router.URL("profile") // "/profile"

    // Route groups
    router.Group(func(r *routing.Router) {
        r.Get("/dashboard", dashboardHandler)
        r.Get("/settings", settingsHandler)
    }).Prefix("/admin").Middleware("auth", "admin")

    // RESTful resources
    router.Resource("posts", postController)
    router.ApiResource("api/v1/posts", apiPostController).Only("index", "show")

    // Domain routing
    router.Domain("{account}.example.com").Group(func(r *routing.Router) {
        r.Get("/dashboard", tenantDashboardHandler)
    })

    // Start server
    http.ListenAndServe(":8080", router)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Welcome to GoFrame"))
}
```

## Route Parameters

```go
// Define parameters
router.Get("/users/{id}", handler)
router.Get("/posts/{post}/comments/{comment}", handler)

// With constraints
router.Get("/users/{id}", handler).Where("id", `\d+`)
router.Get("/users/{name}", handler).Where("name", `[a-zA-Z]+`)

// Access parameters (TODO: Context support coming soon)
// params := request.RouteParams(r)
// id := params["id"]
```

## Route Groups

```go
// Prefix
router.Group(func(r *routing.Router) {
    r.Get("/users", usersHandler)     // /admin/users
    r.Get("/posts", postsHandler)     // /admin/posts
}).Prefix("/admin")

// Middleware
router.Group(func(r *routing.Router) {
    r.Get("/dashboard", dashboardHandler)
}).Middleware("auth", "verified")

// Nested groups
router.Group(func(r *routing.Router) {
    r.Group(func(r2 *routing.Router) {
        r2.Get("/profile", profileHandler)  // /admin/user/profile
    }).Prefix("/user")
}).Prefix("/admin").Middleware("auth")

// Domain
router.Domain("api.example.com").Group(func(r *routing.Router) {
    r.Get("/users", apiUsersHandler)
})
```

## RESTful Resources

```go
// Standard resource (7 routes)
router.Resource("photos", photoController)
// GET    /photos           → Index
// GET    /photos/create    → Create
// POST   /photos           → Store
// GET    /photos/{id}      → Show
// GET    /photos/{id}/edit → Edit
// PUT    /photos/{id}      → Update
// DELETE /photos/{id}      → Destroy

// API resource (5 routes, no create/edit)
router.ApiResource("photos", photoController)

// Filter routes
router.Resource("photos", photoController).Only("index", "show")
router.Resource("photos", photoController).Except("destroy")
```

## Named Routes

```go
// Register named route
router.Get("/users/{id}", handler).Name("user.show")

// Generate URL
url := router.URL("user.show")  // "/users/{id}"

// Generate URL with parameters
url := router.URLWith("user.show", map[string]string{
    "id": "123",
})  // "/users/123"
```

## Middleware

```go
// Per-route
router.Get("/profile", profileHandler).Middleware("auth")

// Per-group
router.Group(func(r *routing.Router) {
    r.Get("/dashboard", dashboardHandler)
}).Middleware("auth", "verified")

// Multiple middleware
router.Get("/admin", adminHandler).Middleware("auth", "admin", "2fa")
```

## Redirects

```go
// Temporary redirect (302)
router.Redirect("/old", "/new", http.StatusFound)

// Permanent redirect (301)
router.PermanentRedirect("/old", "/new")
```

## Fallback

```go
// Custom 404 handler
router.Fallback(routing.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusNotFound)
    w.Write([]byte("Custom 404 Page"))
}))
```

## Thread Safety

All router methods are thread-safe. You can register routes from multiple goroutines:

```go
router := routing.NewRouter()

go func() {
    router.Get("/route1", handler1)
}()

go func() {
    router.Get("/route2", handler2)
}()
```

## Testing

```go
import "net/http/httptest"

router := routing.NewRouter()
router.Get("/test", handler)

req := httptest.NewRequest("GET", "/test", nil)
route, params, matched := router.MatchRequest(req)
// Check route, params, matched
```

## Package Facade

Use the global router instance:

```go
import "github.com/sazzad/goframe/routing"

routing.DefaultRouter.Get("/", homeHandler)
http.ListenAndServe(":8080", routing.DefaultRouter)
```

## Implementation Notes

- Uses Go 1.22+ `{param}` syntax for route parameters
- Parameters extracted with simple pattern matching
- Regex constraints anchored automatically (^pattern$)
- Domain matching supports wildcards: `{account}.example.com`
- Thread-safe with RWMutex for route storage
- Zero external dependencies

## Roadmap

- [ ] Context-aware handlers with Request/Response types
- [ ] Automatic route model binding
- [ ] Route caching for production
- [ ] Route:list CLI command
- [ ] View routes (render template directly)
- [ ] Rate limiting per route
- [ ] CORS middleware integration

## License

Part of the GoFrame framework.
