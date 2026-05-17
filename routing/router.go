package routing

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"sync"
)

// Router manages HTTP routes and dispatching.
type Router struct {
	mu              sync.RWMutex
	routes          []*Route
	namedRoutes     map[string]*Route
	currentGroup    *RouteGroup
	pendingGroup    *RouteGroup // Used for chaining Domain().Group()
	fallbackHandler HandlerFunc
}

// NewRouter creates a new Router instance.
func NewRouter() *Router {
	return &Router{
		routes:      []*Route{},
		namedRoutes: make(map[string]*Route),
	}
}

// Get registers a GET route.
func (r *Router) Get(path string, handler HandlerFunc) *Route {
	return r.addRoute("GET", path, handler)
}

// Post registers a POST route.
func (r *Router) Post(path string, handler HandlerFunc) *Route {
	return r.addRoute("POST", path, handler)
}

// Put registers a PUT route.
func (r *Router) Put(path string, handler HandlerFunc) *Route {
	return r.addRoute("PUT", path, handler)
}

// Patch registers a PATCH route.
func (r *Router) Patch(path string, handler HandlerFunc) *Route {
	return r.addRoute("PATCH", path, handler)
}

// Delete registers a DELETE route.
func (r *Router) Delete(path string, handler HandlerFunc) *Route {
	return r.addRoute("DELETE", path, handler)
}

// Options registers an OPTIONS route.
func (r *Router) Options(path string, handler HandlerFunc) *Route {
	return r.addRoute("OPTIONS", path, handler)
}

// Any registers a route for all HTTP methods.
func (r *Router) Any(path string, handler HandlerFunc) []*Route {
	methods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"}
	routes := make([]*Route, 0, len(methods))

	for _, method := range methods {
		route := r.addRoute(method, path, handler)
		routes = append(routes, route)
	}

	return routes
}

// Match registers a route for the specified HTTP methods.
func (r *Router) MatchMethods(methods []string, path string, handler HandlerFunc) []*Route {
	routes := make([]*Route, 0, len(methods))

	for _, method := range methods {
		route := r.addRoute(strings.ToUpper(method), path, handler)
		routes = append(routes, route)
	}

	return routes
}

// Group creates a route group with shared attributes.
func (r *Router) Group(callback func(*Router)) *RouteGroup {
	// Store previous group
	previousGroup := r.currentGroup

	// Create new group inheriting from previous or pending group
	group := newRouteGroup(r)
	if r.pendingGroup != nil {
		group.prefix = r.pendingGroup.prefix
		group.middlewares = append([]string{}, r.pendingGroup.middlewares...)
		group.domain = r.pendingGroup.domain
		r.pendingGroup = nil
	} else if previousGroup != nil {
		group.prefix = previousGroup.prefix
		group.middlewares = append([]string{}, previousGroup.middlewares...)
		group.domain = previousGroup.domain
	}

	r.currentGroup = group

	// Track which routes are added in this group
	r.mu.Lock()
	startIdx := len(r.routes)
	r.mu.Unlock()

	// Execute callback
	callback(r)

	// Record routes added during callback so group can retroactively apply attributes
	r.mu.Lock()
	group.routes = r.routes[startIdx:]
	r.mu.Unlock()

	// Restore previous group
	r.currentGroup = previousGroup

	return group
}

// Prefix sets a prefix for the current group (chainable with Group).
func (r *Router) Prefix(prefix string) *Router {
	if r.currentGroup != nil {
		r.currentGroup.Prefix(prefix)
	} else {
		if r.pendingGroup == nil {
			r.pendingGroup = newRouteGroup(r)
		}
		r.pendingGroup.Prefix(prefix)
	}
	return r
}

// Middleware adds middleware to the current group.
func (r *Router) Middleware(middleware ...string) *Router {
	if r.currentGroup != nil {
		r.currentGroup.Middleware(middleware...)
	} else {
		if r.pendingGroup == nil {
			r.pendingGroup = newRouteGroup(r)
		}
		r.pendingGroup.Middleware(middleware...)
	}
	return r
}

// Domain restricts routes to a specific domain.
// Can be chained before Group() to set domain for upcoming group.
func (r *Router) Domain(domain string) *Router {
	if r.currentGroup != nil {
		r.currentGroup.Domain(domain)
	} else {
		r.pendingGroup = &RouteGroup{router: r, domain: domain, middlewares: []string{}}
	}
	return r
}

// Resource registers RESTful resource routes.
func (r *Router) Resource(name string, handler HandlerFunc) *ResourceRegistrar {
	registrar := newResourceRegistrar(r, name, handler, false)
	return registrar
}

// ApiResource registers API resource routes (no create/edit).
func (r *Router) ApiResource(name string, handler HandlerFunc) *ResourceRegistrar {
	registrar := newResourceRegistrar(r, name, handler, true)
	return registrar
}

// Redirect registers a redirect route.
func (r *Router) Redirect(from, to string, status int) *Route {
	handler := func(w http.ResponseWriter, req *http.Request) {
		http.Redirect(w, req, to, status)
	}
	return r.Get(from, HandlerFunc(handler))
}

// PermanentRedirect registers a permanent redirect (301).
func (r *Router) PermanentRedirect(from, to string) *Route {
	return r.Redirect(from, to, http.StatusMovedPermanently)
}

// Fallback sets a fallback handler for unmatched routes.
func (r *Router) Fallback(handler HandlerFunc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallbackHandler = handler
}

// addRoute adds a new route to the router.
func (r *Router) addRoute(method, path string, handler HandlerFunc) *Route {
	r.mu.Lock()
	defer r.mu.Unlock()

	route := newRoute(method, path, handler)
	route.router = r

	// Apply current group attributes if inside a group
	if r.currentGroup != nil {
		r.currentGroup.applyToRoute(route)
	}

	r.routes = append(r.routes, route)
	return route
}

// Name registers a named route for the last added route.
// This is called via route.Name(), but we need to track it here.
func (r *Router) registerNamedRoute(name string, route *Route) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.namedRoutes[name] = route
}

// URL generates a URL for a named route.
func (r *Router) URL(name string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	route, exists := r.namedRoutes[name]
	if !exists {
		return ""
	}

	return route.path
}

// URLWith generates a URL for a named route with parameters.
func (r *Router) URLWith(name string, params map[string]string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	route, exists := r.namedRoutes[name]
	if !exists {
		return ""
	}

	url := route.path

	// Replace parameters in the URL
	for key, value := range params {
		placeholder := "{" + key + "}"
		url = strings.ReplaceAll(url, placeholder, value)
	}

	return url
}

// MatchRequest finds the matching route for the given request.
func (r *Router) MatchRequest(req *http.Request) (*Route, map[string]string, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	method := req.Method
	path := req.URL.Path
	host := req.Host

	// Try to match routes
	for _, route := range r.routes {
		// Check method
		if route.method != method {
			continue
		}

		// Check domain if specified
		if route.domain != "" && !matchDomain(route.domain, host) {
			continue
		}

		// Check path pattern
		if params, matched := matchPath(route.path, path, route.constraints); matched {
			return route, params, true
		}
	}

	return nil, nil, false
}

// ServeHTTP implements http.Handler interface.
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	route, params, matched := r.MatchRequest(req)

	if matched {
		if len(params) > 0 {
			ctx := req.Context()
			for k, v := range params {
				ctx = contextWithParam(ctx, k, v)
			}
			req = req.WithContext(ctx)
		}
		route.handler(w, req)
		return
	}

	// Call fallback handler if registered
	if r.fallbackHandler != nil {
		r.fallbackHandler(w, req)
		return
	}

	// Default 404
	http.NotFound(w, req)
}

// Routes returns all registered routes.
func (r *Router) Routes() []*Route {
	r.mu.RLock()
	defer r.mu.RUnlock()

	routes := make([]*Route, len(r.routes))
	copy(routes, r.routes)
	return routes
}

// RouteList returns a formatted string table of all registered routes.
func (r *Router) RouteList() string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.routes) == 0 {
		return "No routes registered."
	}

	// Calculate column widths
	methodW, pathW, nameW, mwW := 6, 4, 4, 10
	for _, route := range r.routes {
		if len(route.method) > methodW {
			methodW = len(route.method)
		}
		if len(route.path) > pathW {
			pathW = len(route.path)
		}
		if len(route.name) > nameW {
			nameW = len(route.name)
		}
		mw := strings.Join(route.middlewares, ", ")
		if len(mw) > mwW {
			mwW = len(mw)
		}
	}

	var sb strings.Builder
	format := "  %-" + fmt.Sprintf("%d", methodW) + "s  %-" + fmt.Sprintf("%d", pathW) + "s  %-" + fmt.Sprintf("%d", nameW) + "s  %s\n"

	sb.WriteString(fmt.Sprintf(format, "METHOD", "URI", "NAME", "MIDDLEWARE"))
	sb.WriteString(fmt.Sprintf("  %s  %s  %s  %s\n",
		strings.Repeat("-", methodW),
		strings.Repeat("-", pathW),
		strings.Repeat("-", nameW),
		strings.Repeat("-", mwW)))

	for _, route := range r.routes {
		mw := strings.Join(route.middlewares, ", ")
		if mw == "" {
			mw = "-"
		}
		name := route.name
		if name == "" {
			name = "-"
		}
		sb.WriteString(fmt.Sprintf(format, route.method, route.path, name, mw))
	}
	return sb.String()
}

// PrintRoutes prints the route list to stdout.
func (r *Router) PrintRoutes() {
	fmt.Print(r.RouteList())
}

// matchDomain checks if the request host matches the route domain pattern.
func matchDomain(pattern, host string) bool {
	// Remove port from host if present
	if idx := strings.Index(host, ":"); idx != -1 {
		host = host[:idx]
	}

	// Simple wildcard matching for now
	if strings.Contains(pattern, "{") {
		// Pattern has parameters, use regex-like matching
		// For simplicity, just check if the static parts match
		patternParts := strings.Split(pattern, ".")
		hostParts := strings.Split(host, ".")

		if len(patternParts) != len(hostParts) {
			return false
		}

		for i, part := range patternParts {
			if !strings.HasPrefix(part, "{") && part != hostParts[i] {
				return false
			}
		}

		return true
	}

	return pattern == host
}

// matchPath checks if the request path matches the route pattern.
// Returns extracted parameters and match status.
func matchPath(pattern, path string, constraints map[string]*regexp.Regexp) (map[string]string, bool) {
	// Normalize paths
	pattern = strings.TrimRight(pattern, "/")
	path = strings.TrimRight(path, "/")

	if pattern == "" {
		pattern = "/"
	}
	if path == "" {
		path = "/"
	}

	// Split into segments
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")
	pathParts := strings.Split(strings.Trim(path, "/"), "/")

	// Length must match (unless pattern is just "/" or empty)
	if pattern == "/" && path == "/" {
		return map[string]string{}, true
	}

	if len(patternParts) != len(pathParts) {
		return nil, false
	}

	params := make(map[string]string)

	for i, part := range patternParts {
		// Check if this is a parameter
		if strings.HasPrefix(part, "{") && strings.HasSuffix(part, "}") {
			paramName := strings.Trim(part, "{}")

			// Check constraint if exists
			if regex, exists := constraints[paramName]; exists {
				if !regex.MatchString(pathParts[i]) {
					return nil, false
				}
			}

			params[paramName] = pathParts[i]
		} else {
			// Static segment must match exactly
			if part != pathParts[i] {
				return nil, false
			}
		}
	}

	return params, true
}

// DefaultRouter is the package-level facade variable.
var DefaultRouter = NewRouter()
