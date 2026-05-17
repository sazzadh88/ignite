package http

import (
	"io"
	"net/http"
	"strconv"
	"strings"
)

// Request wraps *http.Request with Laravel-inspired helper methods.
type Request struct {
	raw         *http.Request
	routeParams map[string]string
	inputCache  map[string]any
}

// NewRequest creates a new Request wrapper from *http.Request.
func NewRequest(r *http.Request) *Request {
	return &Request{
		raw:         r,
		routeParams: make(map[string]string),
		inputCache:  make(map[string]any),
	}
}

// Raw returns the underlying *http.Request.
func (r *Request) Raw() *http.Request {
	return r.raw
}

// SetRouteParam sets a route parameter.
func (r *Request) SetRouteParam(key, value string) {
	r.routeParams[key] = value
}

// Input retrieves an input value from query, form, or JSON body.
func (r *Request) Input(key string) any {
	if val, exists := r.inputCache[key]; exists {
		return val
	}

	// Try query params first
	if values := r.raw.URL.Query(); values.Has(key) {
		return values.Get(key)
	}

	// Try form data
	r.raw.ParseForm()
	r.raw.ParseMultipartForm(32 << 20)

	if r.raw.Form != nil && len(r.raw.Form[key]) > 0 {
		return r.raw.Form.Get(key)
	}

	return nil
}

// InputString retrieves an input value as a string.
func (r *Request) InputString(key string) string {
	val := r.Input(key)
	if val == nil {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// InputInt retrieves an input value as an integer with a default.
func (r *Request) InputInt(key string, defaultVal int) int {
	val := r.InputString(key)
	if val == "" {
		return defaultVal
	}
	if intVal, err := strconv.Atoi(val); err == nil {
		return intVal
	}
	return defaultVal
}

// All returns all input values from query and form data.
func (r *Request) All() map[string]any {
	result := make(map[string]any)

	// Parse forms
	r.raw.ParseForm()
	r.raw.ParseMultipartForm(32 << 20)

	// Add query params
	for key, values := range r.raw.URL.Query() {
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = values
		}
	}

	// Add form values (overwrite query if present)
	for key, values := range r.raw.Form {
		if len(values) == 1 {
			result[key] = values[0]
		} else {
			result[key] = values
		}
	}

	return result
}

// Only returns only the specified keys from input.
func (r *Request) Only(keys ...string) map[string]any {
	all := r.All()
	result := make(map[string]any)

	for _, key := range keys {
		if val, exists := all[key]; exists {
			result[key] = val
		}
	}

	return result
}

// Except returns all input except the specified keys.
func (r *Request) Except(keys ...string) map[string]any {
	all := r.All()
	exclude := make(map[string]bool)

	for _, key := range keys {
		exclude[key] = true
	}

	result := make(map[string]any)
	for key, val := range all {
		if !exclude[key] {
			result[key] = val
		}
	}

	return result
}

// Has checks if the request contains a given input key.
func (r *Request) Has(key string) bool {
	return r.Input(key) != nil
}

// Filled checks if the request contains a non-empty value for the key.
func (r *Request) Filled(key string) bool {
	val := r.InputString(key)
	return val != ""
}

// Missing checks if the request is missing a given input key.
func (r *Request) Missing(key string) bool {
	return !r.Has(key)
}

// Query retrieves a query string value with an optional default.
func (r *Request) Query(key string, defaultVal ...string) string {
	val := r.raw.URL.Query().Get(key)
	if val == "" && len(defaultVal) > 0 {
		return defaultVal[0]
	}
	return val
}

// RouteParam retrieves a URL path parameter.
func (r *Request) RouteParam(key string) string {
	return r.routeParams[key]
}

// Header retrieves a header value.
func (r *Request) Header(key string) string {
	return r.raw.Header.Get(key)
}

// IP returns the client IP address.
func (r *Request) IP() string {
	// Check X-Forwarded-For header first
	if ip := r.raw.Header.Get("X-Forwarded-For"); ip != "" {
		parts := strings.Split(ip, ",")
		return strings.TrimSpace(parts[0])
	}

	// Check X-Real-IP header
	if ip := r.raw.Header.Get("X-Real-IP"); ip != "" {
		return ip
	}

	// Fall back to RemoteAddr
	ip := r.raw.RemoteAddr
	if idx := strings.LastIndex(ip, ":"); idx != -1 {
		ip = ip[:idx]
	}
	return ip
}

// Method returns the HTTP method.
func (r *Request) Method() string {
	return r.raw.Method
}

// Path returns the request path.
func (r *Request) Path() string {
	return r.raw.URL.Path
}

// URL returns the request URL without query string.
func (r *Request) URL() string {
	return r.raw.URL.Scheme + "://" + r.raw.Host + r.raw.URL.Path
}

// FullURL returns the full request URL including query string.
func (r *Request) FullURL() string {
	fullURL := r.URL()
	if r.raw.URL.RawQuery != "" {
		fullURL += "?" + r.raw.URL.RawQuery
	}
	return fullURL
}

// WantsJSON checks if the request expects JSON response.
func (r *Request) WantsJSON() bool {
	accept := r.Header("Accept")
	return strings.Contains(accept, "application/json") ||
		strings.Contains(accept, "text/json")
}

// Ajax checks if the request is an AJAX request.
func (r *Request) Ajax() bool {
	return r.Header("X-Requested-With") == "XMLHttpRequest"
}

// Secure checks if the request is over HTTPS.
func (r *Request) Secure() bool {
	return r.raw.TLS != nil || r.raw.URL.Scheme == "https"
}

// File retrieves an uploaded file by key.
func (r *Request) File(key string) *UploadedFile {
	if r.raw.MultipartForm == nil {
		r.raw.ParseMultipartForm(32 << 20)
	}

	if r.raw.MultipartForm == nil {
		return nil
	}

	if files, exists := r.raw.MultipartForm.File[key]; exists && len(files) > 0 {
		return NewUploadedFile(files[0])
	}

	return nil
}

// BearerToken retrieves the bearer token from the Authorization header.
func (r *Request) BearerToken() string {
	auth := r.Header("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}

// Merge merges the given data into the request input.
func (r *Request) Merge(data map[string]any) {
	for key, val := range data {
		r.inputCache[key] = val
	}
}

// MergeIfMissing merges data only if keys are missing.
func (r *Request) MergeIfMissing(data map[string]any) {
	for key, val := range data {
		if !r.Has(key) {
			r.inputCache[key] = val
		}
	}
}

// Validate is a placeholder for request validation.
// To be implemented with a validation package.
func (r *Request) Validate(rules map[string]string) error {
	// Placeholder - to be implemented
	return nil
}

// Body returns the raw request body reader.
func (r *Request) Body() io.ReadCloser {
	return r.raw.Body
}

// ContentType returns the Content-Type header.
func (r *Request) ContentType() string {
	return r.Header("Content-Type")
}

// UserAgent returns the User-Agent header.
func (r *Request) UserAgent() string {
	return r.Header("User-Agent")
}

// Referer returns the Referer header.
func (r *Request) Referer() string {
	return r.Header("Referer")
}

// Cookie retrieves a cookie by name.
func (r *Request) Cookie(name string) (*http.Cookie, error) {
	return r.raw.Cookie(name)
}

// Cookies returns all cookies.
func (r *Request) Cookies() []*http.Cookie {
	return r.raw.Cookies()
}

// Files returns all uploaded files for a given key.
func (r *Request) Files(key string) []*UploadedFile {
	if r.raw.MultipartForm == nil {
		r.raw.ParseMultipartForm(32 << 20)
	}

	if r.raw.MultipartForm == nil {
		return nil
	}

	var result []*UploadedFile
	if files, exists := r.raw.MultipartForm.File[key]; exists {
		for _, fileHeader := range files {
			result = append(result, NewUploadedFile(fileHeader))
		}
	}

	return result
}
