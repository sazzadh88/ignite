// Package middleware provides HTTP middleware functionality for Ignite.
// It implements a Laravel-inspired middleware pipeline with support for
// chaining, termination, and stack management.
package middleware

import "net/http"

// Next is a function that passes control to the next middleware in the chain.
// Middleware can choose to call Next to continue processing, or not call it
// to short-circuit the pipeline.
type Next func(http.ResponseWriter, *http.Request)

// Middleware defines the interface for HTTP middleware handlers.
// Implementations receive the request, response writer, and a Next function
// to pass control to subsequent middleware.
type Middleware interface {
	Handle(http.ResponseWriter, *http.Request, Next)
}

// MiddlewareFunc is a function type that implements the Middleware interface.
// It allows plain functions to be used as middleware without defining a new type.
type MiddlewareFunc func(http.ResponseWriter, *http.Request, Next)

// Handle implements the Middleware interface for MiddlewareFunc.
func (m MiddlewareFunc) Handle(w http.ResponseWriter, r *http.Request, next Next) {
	m(w, r, next)
}
