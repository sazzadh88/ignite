package routing

import "strings"

// ResourceController defines the standard RESTful resource methods.
// Controllers implementing resource routes should implement these methods.
type ResourceController interface{}

// resourceRoute represents a single RESTful route definition.
type resourceRoute struct {
	method string
	path   string
	action string
}

// standardResourceRoutes returns the 7 standard RESTful routes.
func standardResourceRoutes(resource string) []resourceRoute {
	return []resourceRoute{
		{"GET", "/" + resource, "Index"},
		{"GET", "/" + resource + "/create", "Create"},
		{"POST", "/" + resource, "Store"},
		{"GET", "/" + resource + "/{id}", "Show"},
		{"GET", "/" + resource + "/{id}/edit", "Edit"},
		{"PUT", "/" + resource + "/{id}", "Update"},
		{"DELETE", "/" + resource + "/{id}", "Destroy"},
	}
}

// apiResourceRoutes returns the 5 API resource routes (no create/edit).
func apiResourceRoutes(resource string) []resourceRoute {
	return []resourceRoute{
		{"GET", "/" + resource, "Index"},
		{"POST", "/" + resource, "Store"},
		{"GET", "/" + resource + "/{id}", "Show"},
		{"PUT", "/" + resource + "/{id}", "Update"},
		{"DELETE", "/" + resource + "/{id}", "Destroy"},
	}
}

// ResourceRegistrar handles resource route registration with filtering.
type ResourceRegistrar struct {
	router         *Router
	resource       string
	handler        HandlerFunc
	only           []string
	except         []string
	isAPI          bool
	routesBefore   int // Track how many routes existed before registration
	alreadyCleared bool
}

// newResourceRegistrar creates a new resource registrar.
func newResourceRegistrar(router *Router, resource string, handler HandlerFunc, isAPI bool) *ResourceRegistrar {
	router.mu.RLock()
	routesBefore := len(router.routes)
	router.mu.RUnlock()

	registrar := &ResourceRegistrar{
		router:       router,
		resource:     resource,
		handler:      handler,
		isAPI:        isAPI,
		routesBefore: routesBefore,
	}
	// Auto-register immediately (will be cleared and re-registered if Only/Except is called)
	registrar.Register()
	return registrar
}

// Only restricts resource routes to the specified actions.
func (r *ResourceRegistrar) Only(actions ...string) *ResourceRegistrar {
	r.only = actions
	r.clearAndReRegister()
	return r
}

// Except excludes the specified actions from resource routes.
func (r *ResourceRegistrar) Except(actions ...string) *ResourceRegistrar {
	r.except = actions
	r.clearAndReRegister()
	return r
}

// clearAndReRegister removes previously registered routes and re-registers with filters.
func (r *ResourceRegistrar) clearAndReRegister() {
	if r.alreadyCleared {
		return
	}

	r.router.mu.Lock()
	// Remove the routes we added
	r.router.routes = r.router.routes[:r.routesBefore]
	r.router.mu.Unlock()

	r.alreadyCleared = true
	r.Register()
}

// Register registers all filtered resource routes.
func (r *ResourceRegistrar) Register() {
	var routes []resourceRoute
	if r.isAPI {
		routes = apiResourceRoutes(r.resource)
	} else {
		routes = standardResourceRoutes(r.resource)
	}

	for _, route := range routes {
		if r.shouldRegisterAction(route.action) {
			r.registerRoute(route)
		}
	}
}

// shouldRegisterAction checks if an action should be registered based on only/except filters.
func (r *ResourceRegistrar) shouldRegisterAction(action string) bool {
	actionLower := strings.ToLower(action)

	// If only is specified, action must be in only list
	if len(r.only) > 0 {
		for _, allowed := range r.only {
			if strings.ToLower(allowed) == actionLower {
				return true
			}
		}
		return false
	}

	// If except is specified, action must not be in except list
	if len(r.except) > 0 {
		for _, excluded := range r.except {
			if strings.ToLower(excluded) == actionLower {
				return false
			}
		}
	}

	return true
}

// registerRoute registers a single resource route.
func (r *ResourceRegistrar) registerRoute(route resourceRoute) {
	name := r.resource + "." + strings.ToLower(route.action)

	switch route.method {
	case "GET":
		r.router.Get(route.path, r.handler).Name(name)
	case "POST":
		r.router.Post(route.path, r.handler).Name(name)
	case "PUT":
		r.router.Put(route.path, r.handler).Name(name)
	case "DELETE":
		r.router.Delete(route.path, r.handler).Name(name)
	}
}
