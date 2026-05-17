package http

import (
	"encoding/json"
	"net/http"
	"sync"
)

// Context ties together Request, ResponseWriter, and route parameters.
// It provides a convenient API for handling HTTP requests similar to other frameworks.
type Context struct {
	Request  *Request
	Writer   http.ResponseWriter
	params   map[string]string
	values   map[string]any
	mu       sync.RWMutex
}

// NewContext creates a new Context.
func NewContext(w http.ResponseWriter, r *http.Request) *Context {
	return &Context{
		Request: NewRequest(r),
		Writer:  w,
		params:  make(map[string]string),
		values:  make(map[string]any),
	}
}

// Param retrieves a route parameter by key.
func (c *Context) Param(key string) string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.params[key]
}

// SetParam sets a route parameter.
func (c *Context) SetParam(key, value string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.params[key] = value
	c.Request.SetRouteParam(key, value)
}

// JSON sends a JSON response.
func (c *Context) JSON(status int, data any) {
	c.Writer.Header().Set("Content-Type", "application/json")
	c.Writer.WriteHeader(status)
	json.NewEncoder(c.Writer).Encode(data)
}

// String sends a plain text response.
func (c *Context) String(status int, text string) {
	c.Writer.Header().Set("Content-Type", "text/plain; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(text))
}

// HTML sends an HTML response.
func (c *Context) HTML(status int, html string) {
	c.Writer.Header().Set("Content-Type", "text/html; charset=utf-8")
	c.Writer.WriteHeader(status)
	c.Writer.Write([]byte(html))
}

// Redirect sends a redirect response.
func (c *Context) Redirect(url string, status ...int) {
	statusCode := http.StatusFound
	if len(status) > 0 {
		statusCode = status[0]
	}
	http.Redirect(c.Writer, c.Request.Raw(), url, statusCode)
}

// Back redirects to the referer with fallback.
func (c *Context) Back(fallback ...string) {
	referer := c.Request.Referer()
	if referer == "" && len(fallback) > 0 {
		referer = fallback[0]
	}
	if referer == "" {
		referer = "/"
	}
	c.Redirect(referer)
}

// Set stores a value in the context.
func (c *Context) Set(key string, value any) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.values[key] = value
}

// Get retrieves a value from the context.
func (c *Context) Get(key string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, exists := c.values[key]
	return val, exists
}

// MustGet retrieves a value from the context or panics if not found.
func (c *Context) MustGet(key string) any {
	val, exists := c.Get(key)
	if !exists {
		panic("key \"" + key + "\" does not exist in context")
	}
	return val
}

// Status sets the response status code.
func (c *Context) Status(code int) *Context {
	c.Writer.WriteHeader(code)
	return c
}

// Header sets a response header.
func (c *Context) Header(key, value string) {
	c.Writer.Header().Set(key, value)
}

// Cookie sets a cookie.
func (c *Context) Cookie(cookie *http.Cookie) {
	http.SetCookie(c.Writer, cookie)
}

// Abort stops the request chain execution.
// This is typically used in middleware to halt further processing.
func (c *Context) Abort() {
	// This will be implemented when middleware chain is added
	c.Set("__aborted", true)
}

// IsAborted checks if the request chain has been aborted.
func (c *Context) IsAborted() bool {
	val, exists := c.Get("__aborted")
	if !exists {
		return false
	}
	aborted, ok := val.(bool)
	return ok && aborted
}

// Error sends an error response with the given status and message.
func (c *Context) Error(status int, message string) {
	c.JSON(status, map[string]string{"error": message})
}

// NoContent sends a 204 No Content response.
func (c *Context) NoContent() {
	c.Writer.WriteHeader(http.StatusNoContent)
}

// SetContentType sets the Content-Type header.
func (c *Context) SetContentType(contentType string) {
	c.Header("Content-Type", contentType)
}

// Send sends a Response object.
func (c *Context) Send(response *Response) {
	response.Send(c.Writer)
}
