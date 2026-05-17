package middleware

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"runtime/debug"
	"strings"
	"time"
)

// TrimStrings returns middleware that trims whitespace from string inputs.
// This is commonly used to normalize form data and query parameters.
func TrimStrings() Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		// Trim query parameters
		query := r.URL.Query()
		for key, values := range query {
			for i, value := range values {
				query[key][i] = strings.TrimSpace(value)
			}
		}
		r.URL.RawQuery = query.Encode()

		// Parse form if present
		if r.Form != nil {
			for key, values := range r.Form {
				for i, value := range values {
					r.Form[key][i] = strings.TrimSpace(value)
				}
			}
		}

		if r.PostForm != nil {
			for key, values := range r.PostForm {
				for i, value := range values {
					r.PostForm[key][i] = strings.TrimSpace(value)
				}
			}
		}

		next(w, r)
	})
}

// CORSConfig defines configuration for CORS middleware.
type CORSConfig struct {
	AllowedOrigins   []string
	AllowedMethods   []string
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a permissive CORS configuration.
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders: []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders: []string{},
		MaxAge:         3600,
	}
}

// CORS returns middleware that handles Cross-Origin Resource Sharing.
func CORS(config CORSConfig) Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		origin := r.Header.Get("Origin")

		// Check if origin is allowed
		allowedOrigin := ""
		if len(config.AllowedOrigins) > 0 && config.AllowedOrigins[0] == "*" {
			allowedOrigin = "*"
		} else {
			for _, o := range config.AllowedOrigins {
				if o == origin {
					allowedOrigin = origin
					break
				}
			}
		}

		if allowedOrigin != "" {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
		}

		if config.AllowCredentials {
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowedMethods) > 0 {
			w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowedMethods, ", "))
		}

		if len(config.AllowedHeaders) > 0 {
			w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowedHeaders, ", "))
		}

		if len(config.ExposedHeaders) > 0 {
			w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposedHeaders, ", "))
		}

		if config.MaxAge > 0 {
			w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
		}

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next(w, r)
	})
}

// Recovery returns middleware that recovers from panics and returns
// a 500 Internal Server Error response.
func Recovery() Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		defer func() {
			if err := recover(); err != nil {
				// Log the error and stack trace
				fmt.Printf("panic recovered: %v\n%s\n", err, debug.Stack())

				// Return 500 error
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Internal Server Error"))
			}
		}()

		next(w, r)
	})
}

// LoggerConfig defines configuration for the Logger middleware.
type LoggerConfig struct {
	// SkipPaths defines paths to skip logging
	SkipPaths []string
	// Output defines where to write logs (defaults to stdout)
	Output func(format string, v ...interface{})
}

// Logger returns middleware that logs HTTP requests.
func Logger(config LoggerConfig) Middleware {
	if config.Output == nil {
		config.Output = func(format string, v ...interface{}) {
			fmt.Printf(format, v...)
		}
	}

	skipPaths := make(map[string]bool)
	for _, path := range config.SkipPaths {
		skipPaths[path] = true
	}

	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		// Skip logging for configured paths
		if skipPaths[r.URL.Path] {
			next(w, r)
			return
		}

		start := time.Now()
		path := r.URL.Path
		query := r.URL.RawQuery

		// Create response recorder to capture status code
		recorder := newResponseRecorder(w)

		// Process request
		next(recorder, r)

		// Log request details
		latency := time.Since(start)
		statusCode := recorder.statusCode

		config.Output("[%s] %d %s %s %s %v\n",
			time.Now().Format("2006-01-02 15:04:05"),
			statusCode,
			r.Method,
			path,
			query,
			latency,
		)
	})
}

// RequestID returns middleware that adds a unique X-Request-ID header
// to each request. If the header already exists, it is preserved.
func RequestID() Middleware {
	return MiddlewareFunc(func(w http.ResponseWriter, r *http.Request, next Next) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			b := make([]byte, 16)
			rand.Read(b)
			requestID = fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
		}

		// Set request ID in context for handlers to use
		r.Header.Set("X-Request-ID", requestID)

		// Also set in response headers
		w.Header().Set("X-Request-ID", requestID)

		next(w, r)
	})
}
