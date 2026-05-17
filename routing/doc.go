// Package routing provides Laravel-style HTTP routing for Ignite.
//
// The routing package mirrors Laravel's Router API, offering a fluent interface
// for registering routes, route groups, middleware, and RESTful resources.
//
// # Basic Usage
//
// Register routes using HTTP verb methods:
//
//	router := routing.NewRouter()
//	router.Get("/users", userHandler)
//	router.Post("/users", createUserHandler)
//	router.Put("/users/{id}", updateUserHandler)
//	router.Delete("/users/{id}", deleteUserHandler)
//
// # Route Parameters
//
// Define route parameters using curly braces:
//
//	router.Get("/users/{id}", showUserHandler)
//	router.Get("/posts/{post}/comments/{comment}", showCommentHandler)
//
// # Route Constraints
//
// Add regex constraints to route parameters:
//
//	router.Get("/users/{id}", handler).Where("id", `\d+`)
//	router.Get("/users/{name}", handler).Where("name", `[a-zA-Z]+`)
//
// # Named Routes
//
// Name routes for URL generation:
//
//	router.Get("/users/{id}", handler).Name("user.show")
//	url := router.URL("user.show")                          // "/users/{id}"
//	url := router.URLWith("user.show", map[string]string{"id": "5"}) // "/users/5"
//
// # Route Groups
//
// Group routes with shared attributes:
//
//	router.Group(func(r *routing.Router) {
//	    r.Get("/dashboard", dashboardHandler)
//	    r.Get("/settings", settingsHandler)
//	}).Prefix("/admin").Middleware("auth", "admin")
//
// # RESTful Resources
//
// Register all RESTful routes for a resource:
//
//	router.Resource("posts", postHandler)
//	// Registers: Index, Create, Store, Show, Edit, Update, Destroy
//
//	router.ApiResource("posts", postHandler)
//	// Registers: Index, Store, Show, Update, Destroy (no Create/Edit)
//
// Filter resource routes:
//
//	router.Resource("posts", handler).Only("index", "show")
//	router.Resource("posts", handler).Except("destroy")
//
// # Middleware
//
// Apply middleware to routes or groups:
//
//	router.Get("/profile", handler).Middleware("auth")
//	router.Group(func(r *routing.Router) {
//	    r.Get("/users", usersHandler)
//	}).Middleware("auth", "verified")
//
// # Domain Routing
//
// Restrict routes to specific domains:
//
//	router.Domain("api.example.com").Group(func(r *routing.Router) {
//	    r.Get("/users", apiUsersHandler)
//	})
//
//	router.Domain("{account}.example.com").Group(func(r *routing.Router) {
//	    r.Get("/dashboard", tenantDashboardHandler)
//	})
//
// # Redirects
//
//	router.Redirect("/old", "/new", http.StatusMovedPermanently)
//	router.PermanentRedirect("/old", "/new")
//
// # Fallback Routes
//
// Handle unmatched routes:
//
//	router.Fallback(notFoundHandler)
//
// # HTTP Server Integration
//
// Router implements http.Handler:
//
//	http.ListenAndServe(":8080", router)
//
// # Thread Safety
//
// Router is safe for concurrent route registration and matching.
//
// # Package-level Facade
//
// Use DefaultRouter for global route registration:
//
//	routing.DefaultRouter.Get("/", homeHandler)
package routing
