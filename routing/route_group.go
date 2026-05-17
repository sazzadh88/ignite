package routing

// RouteGroup represents a group of routes with shared attributes.
type RouteGroup struct {
	router      *Router
	prefix      string
	middlewares []string
	domain      string
	namespace   string
	routes      []*Route
}

// newRouteGroup creates a new RouteGroup instance.
func newRouteGroup(router *Router) *RouteGroup {
	return &RouteGroup{
		router:      router,
		middlewares: []string{},
	}
}

// Prefix sets the URI prefix for the group.
// If routes were already added (chained after Group), applies retroactively.
func (g *RouteGroup) Prefix(prefix string) *RouteGroup {
	g.prefix = prefix
	for _, route := range g.routes {
		route.path = prefix + route.path
	}
	return g
}

// Middleware adds middleware to the group.
// If routes were already added (chained after Group), applies retroactively.
func (g *RouteGroup) Middleware(middleware ...string) *RouteGroup {
	g.middlewares = append(g.middlewares, middleware...)
	for _, route := range g.routes {
		route.middlewares = append(middleware, route.middlewares...)
	}
	return g
}

// Domain restricts the group to a specific domain.
func (g *RouteGroup) Domain(domain string) *RouteGroup {
	g.domain = domain
	return g
}

// Namespace sets the controller namespace (future use).
func (g *RouteGroup) Namespace(namespace string) *RouteGroup {
	g.namespace = namespace
	return g
}

// applyToRoute applies group attributes to a route.
func (g *RouteGroup) applyToRoute(route *Route) {
	// Apply prefix
	if g.prefix != "" {
		route.path = g.prefix + route.path
	}

	// Apply middlewares
	if len(g.middlewares) > 0 {
		route.middlewares = append(g.middlewares, route.middlewares...)
	}

	// Apply domain
	if g.domain != "" && route.domain == "" {
		route.domain = g.domain
	}
}
