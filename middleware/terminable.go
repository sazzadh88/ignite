package middleware

import "net/http"

// Terminable defines an interface for middleware that needs to perform
// cleanup or additional processing after the response has been sent.
// This is useful for tasks like logging, metrics, or resource cleanup
// that should not delay the response to the client.
type Terminable interface {
	Middleware
	// Terminate is called after the response has been sent to the client.
	// It receives the original request and response for inspection.
	Terminate(*http.Request, http.ResponseWriter)
}

// responseRecorder wraps http.ResponseWriter to capture response details
// for terminable middleware.
type responseRecorder struct {
	http.ResponseWriter
	statusCode int
	written    bool
}

// newResponseRecorder creates a new response recorder.
func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

// WriteHeader captures the status code and delegates to the wrapped writer.
func (r *responseRecorder) WriteHeader(statusCode int) {
	if !r.written {
		r.statusCode = statusCode
		r.written = true
	}
	r.ResponseWriter.WriteHeader(statusCode)
}

// Write delegates to the wrapped writer and marks response as written.
func (r *responseRecorder) Write(b []byte) (int, error) {
	if !r.written {
		r.written = true
	}
	return r.ResponseWriter.Write(b)
}

// WrapTerminable wraps a handler to support terminable middleware.
// It collects all terminable middleware and calls their Terminate methods
// after the handler completes.
func WrapTerminable(handler http.HandlerFunc, middlewares []Middleware) http.HandlerFunc {
	// Extract terminable middleware
	terminables := make([]Terminable, 0)
	for _, m := range middlewares {
		if t, ok := m.(Terminable); ok {
			terminables = append(terminables, t)
		}
	}

	if len(terminables) == 0 {
		return handler
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// Record response for terminable middleware
		recorder := newResponseRecorder(w)

		// Execute the handler
		handler(recorder, r)

		// Call Terminate on all terminable middleware
		for _, t := range terminables {
			t.Terminate(r, recorder)
		}
	}
}
