package httpclient

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// pendingRequest builds an HTTP request from client configuration.
type pendingRequest struct {
	client      *Client
	method      string
	url         string
	data        map[string]any
	headers     map[string]string
	contentType string
	asJSON      bool
	asForm      bool
}

// newPendingRequest creates a new pending request.
func newPendingRequest(client *Client, method, urlStr string, data map[string]any) *pendingRequest {
	return &pendingRequest{
		client:  client,
		method:  method,
		url:     urlStr,
		data:    data,
		headers: make(map[string]string),
	}
}

// build constructs the *http.Request.
func (p *pendingRequest) build() (*http.Request, error) {
	var body io.Reader
	var err error

	// Merge client headers
	for k, v := range p.client.headers {
		if _, exists := p.headers[k]; !exists {
			p.headers[k] = v
		}
	}

	// Check for AsJSON/AsForm flags in headers
	if p.client.headers["_as_json"] == "true" {
		p.asJSON = true
	}
	if p.client.headers["_as_form"] == "true" {
		p.asForm = true
	}

	// Get content type from headers if set
	if ct, ok := p.client.headers["Content-Type"]; ok {
		p.contentType = ct
	}

	// Build request body based on content type
	if p.data != nil && len(p.data) > 0 {
		if p.asJSON || p.contentType == "application/json" {
			body, err = p.buildJSONBody()
			if err != nil {
				return nil, err
			}
			if p.contentType == "" {
				p.contentType = "application/json"
			}
		} else if p.asForm || p.method == "POST" || p.method == "PUT" || p.method == "PATCH" {
			body = p.buildFormBody()
			if p.contentType == "" {
				p.contentType = "application/x-www-form-urlencoded"
			}
		}
	}

	req, err := http.NewRequest(p.method, p.url, body)
	if err != nil {
		return nil, err
	}

	// Set headers
	if p.contentType != "" {
		req.Header.Set("Content-Type", p.contentType)
	}
	for k, v := range p.headers {
		// Skip internal flags
		if k == "_as_json" || k == "_as_form" {
			continue
		}
		req.Header.Set(k, v)
	}

	// Set auth
	if p.client.bearerToken != "" {
		req.Header.Set("Authorization", "Bearer "+p.client.bearerToken)
	}
	if p.client.basicAuthUser != "" {
		req.SetBasicAuth(p.client.basicAuthUser, p.client.basicAuthPass)
	}

	// Set cookies
	for name, value := range p.client.cookies {
		req.AddCookie(&http.Cookie{Name: name, Value: value})
	}

	return req, nil
}

// buildJSONBody builds JSON request body.
func (p *pendingRequest) buildJSONBody() (io.Reader, error) {
	jsonBytes, err := json.Marshal(p.data)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(jsonBytes), nil
}

// buildFormBody builds form-urlencoded request body.
func (p *pendingRequest) buildFormBody() io.Reader {
	values := url.Values{}
	for k, v := range p.data {
		values.Set(k, fmt.Sprintf("%v", v))
	}
	return strings.NewReader(values.Encode())
}

// send executes the request with retry logic.
func (p *pendingRequest) send() (*Response, error) {
	var lastErr error
	var lastResp *Response
	attempts := 1
	if p.client.retryTimes > 0 {
		attempts = p.client.retryTimes
	}

	for i := 0; i < attempts; i++ {
		if i > 0 && p.client.retrySleep > 0 {
			time.Sleep(p.client.retrySleep)
		}

		req, err := p.build()
		if err != nil {
			return nil, err
		}

		resp, err := p.client.httpClient.Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		response := NewResponse(resp)

		// If successful or client error, return immediately (don't retry)
		if response.Successful() || response.ClientError() {
			return response, nil
		}

		// Store for potential retry
		lastResp = response
		lastErr = fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	// Return last response if available, otherwise error
	if lastResp != nil {
		return lastResp, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, fmt.Errorf("request failed after %d attempts", attempts)
}
