package httpclient

import (
	"context"
	"net"
	"net/http"
	"time"
)

// Client wraps *http.Client with a fluent API for making HTTP requests.
type Client struct {
	httpClient    *http.Client
	headers       map[string]string
	cookies       map[string]string
	bearerToken   string
	basicAuthUser string
	basicAuthPass string
	retryTimes    int
	retrySleep    time.Duration
}

// New creates a new HTTP client.
func New() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		headers: make(map[string]string),
		cookies: make(map[string]string),
	}
}

// Http is a global instance for facade-style usage.
var Http = New()

// Get performs a GET request.
func (c *Client) Get(url string) (*Response, error) {
	return c.send("GET", url, nil)
}

// Post performs a POST request.
func (c *Client) Post(url string, data map[string]any) (*Response, error) {
	return c.send("POST", url, data)
}

// Put performs a PUT request.
func (c *Client) Put(url string, data map[string]any) (*Response, error) {
	return c.send("PUT", url, data)
}

// Patch performs a PATCH request.
func (c *Client) Patch(url string, data map[string]any) (*Response, error) {
	return c.send("PATCH", url, data)
}

// Delete performs a DELETE request.
func (c *Client) Delete(url string) (*Response, error) {
	return c.send("DELETE", url, nil)
}

// send creates and sends an HTTP request.
func (c *Client) send(method, url string, data map[string]any) (*Response, error) {
	pr := newPendingRequest(c, method, url, data)
	return pr.send()
}

// WithToken sets the Bearer token for authentication.
func (c *Client) WithToken(token string) *Client {
	c.bearerToken = token
	return c
}

// WithBasicAuth sets basic authentication credentials.
func (c *Client) WithBasicAuth(user, pass string) *Client {
	c.basicAuthUser = user
	c.basicAuthPass = pass
	return c
}

// WithHeaders sets multiple headers at once.
func (c *Client) WithHeaders(headers map[string]string) *Client {
	for k, v := range headers {
		c.headers[k] = v
	}
	return c
}

// WithHeader sets a single header.
func (c *Client) WithHeader(key, value string) *Client {
	c.headers[key] = value
	return c
}

// WithCookies sets multiple cookies at once.
func (c *Client) WithCookies(cookies map[string]string) *Client {
	for k, v := range cookies {
		c.cookies[k] = v
	}
	return c
}

// WithoutRedirecting disables automatic redirect following.
func (c *Client) WithoutRedirecting() *Client {
	c.httpClient.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return c
}

// Timeout sets the request timeout.
func (c *Client) Timeout(d time.Duration) *Client {
	c.httpClient.Timeout = d
	return c
}

// ConnectTimeout sets the connection timeout.
func (c *Client) ConnectTimeout(d time.Duration) *Client {
	transport := c.getTransport()
	transport.DialContext = (&timeoutDialer{timeout: d}).DialContext
	return c
}

// Retry sets retry configuration for failed requests.
func (c *Client) Retry(times int, sleep time.Duration) *Client {
	c.retryTimes = times
	c.retrySleep = sleep
	return c
}

// AcceptJSON sets the Accept header to application/json.
func (c *Client) AcceptJSON() *Client {
	return c.WithHeader("Accept", "application/json")
}

// ContentType sets the Content-Type header.
func (c *Client) ContentType(ct string) *Client {
	return c.WithHeader("Content-Type", ct)
}

// AsJSON configures the client to send request body as JSON.
func (c *Client) AsJSON() *Client {
	c.headers["_as_json"] = "true"
	return c.WithHeader("Content-Type", "application/json")
}

// AsForm configures the client to send request body as form-urlencoded.
func (c *Client) AsForm() *Client {
	c.headers["_as_form"] = "true"
	return c.WithHeader("Content-Type", "application/x-www-form-urlencoded")
}

// getTransport returns the http.Transport, creating one if needed.
func (c *Client) getTransport() *http.Transport {
	if c.httpClient.Transport == nil {
		c.httpClient.Transport = &http.Transport{}
	}
	return c.httpClient.Transport.(*http.Transport)
}

// timeoutDialer wraps net.Dialer with a custom timeout.
type timeoutDialer struct {
	timeout time.Duration
}

// DialContext implements the dial function with timeout.
func (d *timeoutDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: d.timeout,
	}
	return dialer.DialContext(ctx, network, addr)
}
