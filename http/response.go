package http

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// Response builder for HTTP responses.
type Response struct {
	statusCode int
	headers    http.Header
	cookies    []*http.Cookie
	body       []byte
}

// NewResponse creates a new Response builder.
func NewResponse() *Response {
	return &Response{
		statusCode: http.StatusOK,
		headers:    make(http.Header),
	}
}

// Make creates a response with the given body and status code.
func Make(body []byte, status ...int) *Response {
	r := NewResponse()
	r.body = body

	if len(status) > 0 {
		r.statusCode = status[0]
	}

	return r
}

// JSON creates a JSON response.
func JSON(data any, status ...int) *Response {
	r := NewResponse()

	if len(status) > 0 {
		r.statusCode = status[0]
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		r.statusCode = http.StatusInternalServerError
		r.body = []byte(`{"error":"Failed to encode JSON"}`)
	} else {
		r.body = jsonData
	}

	r.headers.Set("Content-Type", "application/json")
	return r
}

// Redirect creates a redirect response.
func Redirect(url string, status ...int) *Response {
	r := NewResponse()
	r.statusCode = http.StatusFound

	if len(status) > 0 {
		r.statusCode = status[0]
	}

	r.headers.Set("Location", url)
	return r
}

// RedirectBack creates a redirect response to the referer.
func RedirectBack(fallback string) *Response {
	// This will be enhanced when used with Context that has access to request
	return Redirect(fallback)
}

// Download creates a file download response.
func Download(filePath string, filename ...string) *Response {
	r := NewResponse()

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		r.statusCode = http.StatusNotFound
		r.body = []byte("File not found")
		return r
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		r.statusCode = http.StatusInternalServerError
		r.body = []byte("Failed to read file")
		return r
	}

	r.body = content

	// Set filename
	downloadName := filepath.Base(filePath)
	if len(filename) > 0 {
		downloadName = filename[0]
	}

	r.headers.Set("Content-Disposition", "attachment; filename=\""+downloadName+"\"")
	r.headers.Set("Content-Type", "application/octet-stream")

	return r
}

// File creates a file response for inline display.
func File(filePath string) *Response {
	r := NewResponse()

	// Open file
	file, err := os.Open(filePath)
	if err != nil {
		r.statusCode = http.StatusNotFound
		r.body = []byte("File not found")
		return r
	}
	defer file.Close()

	// Read file content
	content, err := io.ReadAll(file)
	if err != nil {
		r.statusCode = http.StatusInternalServerError
		r.body = []byte("Failed to read file")
		return r
	}

	r.body = content

	// Try to detect content type
	contentType := http.DetectContentType(content)
	r.headers.Set("Content-Type", contentType)

	return r
}

// NoContent creates a 204 No Content response.
func NoContent() *Response {
	r := NewResponse()
	r.statusCode = http.StatusNoContent
	r.body = nil
	return r
}

// Header sets a response header (chainable).
func (r *Response) Header(key, value string) *Response {
	r.headers.Set(key, value)
	return r
}

// Cookie adds a cookie to the response (chainable).
func (r *Response) Cookie(cookie *http.Cookie) *Response {
	r.cookies = append(r.cookies, cookie)
	return r
}

// Status sets the response status code (chainable).
func (r *Response) Status(code int) *Response {
	r.statusCode = code
	return r
}

// Send writes the response to http.ResponseWriter.
func (r *Response) Send(w http.ResponseWriter) {
	// Write headers
	for key, values := range r.headers {
		for _, value := range values {
			w.Header().Add(key, value)
		}
	}

	// Set cookies
	for _, cookie := range r.cookies {
		http.SetCookie(w, cookie)
	}

	// Write status code
	w.WriteHeader(r.statusCode)

	// Write body
	if r.body != nil {
		w.Write(r.body)
	}
}

// GetStatusCode returns the status code.
func (r *Response) GetStatusCode() int {
	return r.statusCode
}

// GetHeaders returns the headers.
func (r *Response) GetHeaders() http.Header {
	return r.headers
}

// GetBody returns the response body.
func (r *Response) GetBody() []byte {
	return r.body
}

// SetBody sets the response body.
func (r *Response) SetBody(body []byte) *Response {
	r.body = body
	return r
}
