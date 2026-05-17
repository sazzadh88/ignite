package routing

import (
	"net/http"
	"regexp"
)

// HandlerFunc defines the handler function type.
// TODO: Replace with context-aware handler once Context/Response types are implemented.
type HandlerFunc http.HandlerFunc

// Route represents a single HTTP route.
type Route struct {
	method      string
	path        string
	handler     HandlerFunc
	name        string
	middlewares []string
	constraints map[string]*regexp.Regexp
	domain      string
	router      *Router
}

// newRoute creates a new Route instance.
func newRoute(method, path string, handler HandlerFunc) *Route {
	return &Route{
		method:      method,
		path:        path,
		handler:     handler,
		middlewares: []string{},
		constraints: make(map[string]*regexp.Regexp),
	}
}

// Name sets the route name for URL generation.
func (r *Route) Name(name string) *Route {
	r.name = name
	// Register named route if router is set
	if r.router != nil {
		r.router.registerNamedRoute(name, r)
	}
	return r
}

// Middleware adds middleware to the route.
func (r *Route) Middleware(middleware ...string) *Route {
	r.middlewares = append(r.middlewares, middleware...)
	return r
}

// Where adds regex constraints for route parameters.
// Example: route.Where("id", `\d+`) restricts id to numeric values.
func (r *Route) Where(param string, pattern string) *Route {
	// Anchor the pattern to match the entire parameter value
	anchoredPattern := "^" + pattern + "$"
	regex, err := regexp.Compile(anchoredPattern)
	if err != nil {
		panic("routing: invalid regex pattern for parameter '" + param + "': " + err.Error())
	}
	r.constraints[param] = regex
	return r
}

// Domain restricts the route to a specific domain.
func (r *Route) Domain(domain string) *Route {
	r.domain = domain
	return r
}

// Method returns the HTTP method.
func (r *Route) Method() string {
	return r.method
}

// Path returns the route path.
func (r *Route) Path() string {
	return r.path
}

// Handler returns the route handler.
func (r *Route) Handler() HandlerFunc {
	return r.handler
}

// GetName returns the route name.
func (r *Route) GetName() string {
	return r.name
}

// Middlewares returns the route middlewares.
func (r *Route) Middlewares() []string {
	return r.middlewares
}

// GetDomain returns the route domain.
func (r *Route) GetDomain() string {
	return r.domain
}

// Constraints returns the route parameter constraints.
func (r *Route) Constraints() map[string]*regexp.Regexp {
	return r.constraints
}
