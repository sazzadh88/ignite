package middleware

import "net/http"

// Pipeline represents a middleware processing pipeline.
// It chains middleware together and executes them in order,
// allowing each middleware to decide whether to pass control to the next.
type Pipeline struct {
	request     *http.Request
	middlewares []Middleware
}

// NewPipeline creates a new middleware pipeline.
func NewPipeline() *Pipeline {
	return &Pipeline{
		middlewares: make([]Middleware, 0),
	}
}

// Send sets the HTTP request to be processed through the pipeline.
// Returns the pipeline for method chaining.
func (p *Pipeline) Send(r *http.Request) *Pipeline {
	p.request = r
	return p
}

// Through adds middleware to the pipeline.
// Middleware will be executed in the order they are added.
// Returns the pipeline for method chaining.
func (p *Pipeline) Through(middlewares []Middleware) *Pipeline {
	p.middlewares = append(p.middlewares, middlewares...)
	return p
}

// Then wraps the final handler with the middleware pipeline.
// It returns an http.HandlerFunc that executes all middleware in order
// before calling the final handler.
func (p *Pipeline) Then(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Use the request from Send if available, otherwise use the one passed in
		req := r
		if p.request != nil {
			req = p.request
		}

		// Build the middleware chain from the end backwards
		next := p.buildChain(handler)
		next(w, req)
	}
}

// ThenReturn returns an http.HandlerFunc that executes the middleware pipeline
// without a final handler. The last middleware in the chain is responsible
// for writing the response.
func (p *Pipeline) ThenReturn() http.HandlerFunc {
	return p.Then(func(w http.ResponseWriter, r *http.Request) {
		// No-op handler - middleware must handle the response
	})
}

// buildChain constructs the middleware chain by wrapping each middleware
// around the next one, starting from the final handler.
func (p *Pipeline) buildChain(finalHandler http.HandlerFunc) Next {
	// Start with the final handler
	next := Next(func(w http.ResponseWriter, r *http.Request) {
		finalHandler(w, r)
	})

	// Wrap each middleware around the next one, in reverse order
	for i := len(p.middlewares) - 1; i >= 0; i-- {
		middleware := p.middlewares[i]
		current := next
		next = func(w http.ResponseWriter, r *http.Request) {
			middleware.Handle(w, r, current)
		}
	}

	return next
}
