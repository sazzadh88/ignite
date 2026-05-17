package httpclient

import (
	"encoding/json"
	"io"
	"net/http"
)

// Response wraps *http.Response with Laravel-inspired helper methods.
type Response struct {
	raw        *http.Response
	bodyBytes  []byte
	bodyString string
	bodyRead   bool
}

// NewResponse creates a new Response wrapper from *http.Response.
func NewResponse(r *http.Response) *Response {
	return &Response{
		raw: r,
	}
}

// readBody reads and caches the response body.
func (r *Response) readBody() error {
	if r.bodyRead {
		return nil
	}

	defer r.raw.Body.Close()
	bytes, err := io.ReadAll(r.raw.Body)
	if err != nil {
		return err
	}

	r.bodyBytes = bytes
	r.bodyString = string(bytes)
	r.bodyRead = true
	return nil
}

// Body returns the response body as a string.
func (r *Response) Body() string {
	r.readBody()
	return r.bodyString
}

// Bytes returns the response body as bytes.
func (r *Response) Bytes() []byte {
	r.readBody()
	return r.bodyBytes
}

// JSON parses the response body as JSON into a map.
func (r *Response) JSON() (map[string]any, error) {
	var result map[string]any
	if err := r.JSONInto(&result); err != nil {
		return nil, err
	}
	return result, nil
}

// JSONInto parses the response body as JSON into the target.
func (r *Response) JSONInto(target any) error {
	if err := r.readBody(); err != nil {
		return err
	}
	return json.Unmarshal(r.bodyBytes, target)
}

// Status returns the HTTP status code.
func (r *Response) Status() int {
	return r.raw.StatusCode
}

// Header returns the value of a response header.
func (r *Response) Header(key string) string {
	return r.raw.Header.Get(key)
}

// Headers returns all response headers.
func (r *Response) Headers() map[string][]string {
	return r.raw.Header
}

// Cookies returns all response cookies.
func (r *Response) Cookies() []*http.Cookie {
	return r.raw.Cookies()
}

// Successful checks if the response status is 2xx.
func (r *Response) Successful() bool {
	return r.raw.StatusCode >= 200 && r.raw.StatusCode < 300
}

// Failed checks if the response status is not 2xx.
func (r *Response) Failed() bool {
	return !r.Successful()
}

// ServerError checks if the response status is 5xx.
func (r *Response) ServerError() bool {
	return r.raw.StatusCode >= 500 && r.raw.StatusCode < 600
}

// ClientError checks if the response status is 4xx.
func (r *Response) ClientError() bool {
	return r.raw.StatusCode >= 400 && r.raw.StatusCode < 500
}

// Redirect checks if the response status is 3xx.
func (r *Response) Redirect() bool {
	return r.raw.StatusCode >= 300 && r.raw.StatusCode < 400
}

// Unauthorized checks if the response status is 401.
func (r *Response) Unauthorized() bool {
	return r.raw.StatusCode == 401
}

// Forbidden checks if the response status is 403.
func (r *Response) Forbidden() bool {
	return r.raw.StatusCode == 403
}

// NotFound checks if the response status is 404.
func (r *Response) NotFound() bool {
	return r.raw.StatusCode == 404
}

// Ok checks if the response status is 200.
func (r *Response) Ok() bool {
	return r.raw.StatusCode == 200
}

// Raw returns the underlying *http.Response.
func (r *Response) Raw() *http.Response {
	return r.raw
}
