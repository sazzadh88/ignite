package httpclient

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"sync"
)

// FakeClient is a fake HTTP client for testing.
type FakeClient struct {
	responses     []fakeResponse
	currentIndex  int
	sentRequests  []*http.Request
	mu            sync.Mutex
	defaultStatus int
}

// fakeResponse represents a fake HTTP response.
type fakeResponse struct {
	body   any
	status int
}

// Fake creates a new fake HTTP client.
func Fake(responses ...fakeResponse) *FakeClient {
	return &FakeClient{
		responses:     responses,
		sentRequests:  make([]*http.Request, 0),
		defaultStatus: 200,
	}
}

// FakeResponse creates a fake response with body and status.
func FakeResponse(body any, status int) fakeResponse {
	return fakeResponse{
		body:   body,
		status: status,
	}
}

// Sequence creates a sequential fake response that returns responses in order.
func Sequence(responses ...fakeResponse) *FakeClient {
	return Fake(responses...)
}

// Get performs a fake GET request.
func (f *FakeClient) Get(url string) (*Response, error) {
	return f.send("GET", url, nil)
}

// Post performs a fake POST request.
func (f *FakeClient) Post(url string, data map[string]any) (*Response, error) {
	return f.send("POST", url, data)
}

// Put performs a fake PUT request.
func (f *FakeClient) Put(url string, data map[string]any) (*Response, error) {
	return f.send("PUT", url, data)
}

// Patch performs a fake PATCH request.
func (f *FakeClient) Patch(url string, data map[string]any) (*Response, error) {
	return f.send("PATCH", url, data)
}

// Delete performs a fake DELETE request.
func (f *FakeClient) Delete(url string) (*Response, error) {
	return f.send("DELETE", url, nil)
}

// send records the request and returns a fake response.
func (f *FakeClient) send(method, url string, data map[string]any) (*Response, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	// Build and record the request
	var body io.Reader
	if data != nil {
		jsonBytes, _ := json.Marshal(data)
		body = bytes.NewReader(jsonBytes)
	}

	req, _ := http.NewRequest(method, url, body)
	f.sentRequests = append(f.sentRequests, req)

	// Get the response
	resp := f.getNextResponse()
	return NewResponse(resp), nil
}

// getNextResponse returns the next fake response.
func (f *FakeClient) getNextResponse() *http.Response {
	if len(f.responses) == 0 {
		return f.buildResponse(nil, f.defaultStatus)
	}

	if f.currentIndex >= len(f.responses) {
		// Return last response if we've exhausted the sequence
		return f.buildResponse(f.responses[len(f.responses)-1].body, f.responses[len(f.responses)-1].status)
	}

	fake := f.responses[f.currentIndex]
	f.currentIndex++

	return f.buildResponse(fake.body, fake.status)
}

// buildResponse builds an *http.Response from body and status.
func (f *FakeClient) buildResponse(body any, status int) *http.Response {
	var bodyReader io.ReadCloser

	if body != nil {
		var bodyBytes []byte
		switch v := body.(type) {
		case string:
			bodyBytes = []byte(v)
		case []byte:
			bodyBytes = v
		default:
			bodyBytes, _ = json.Marshal(v)
		}
		bodyReader = io.NopCloser(bytes.NewReader(bodyBytes))
	} else {
		bodyReader = io.NopCloser(bytes.NewReader([]byte{}))
	}

	return &http.Response{
		StatusCode: status,
		Body:       bodyReader,
		Header:     make(http.Header),
	}
}

// AssertSent asserts that a request matching the function was sent.
func (f *FakeClient) AssertSent(fn func(req *http.Request) bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, req := range f.sentRequests {
		if fn(req) {
			return true
		}
	}
	return false
}

// AssertNotSent asserts that no request matching the function was sent.
func (f *FakeClient) AssertNotSent(fn func(req *http.Request) bool) bool {
	return !f.AssertSent(fn)
}

// AssertSentCount asserts that exactly count requests were sent.
func (f *FakeClient) AssertSentCount(count int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.sentRequests) == count
}

// AssertNothingSent asserts that no requests were sent.
func (f *FakeClient) AssertNothingSent() bool {
	return f.AssertSentCount(0)
}

// SentRequests returns all sent requests for inspection.
func (f *FakeClient) SentRequests() []*http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.sentRequests
}
