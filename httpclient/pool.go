package httpclient

import (
	"sync"
)

// Pool manages concurrent HTTP requests.
type Pool struct {
	requests  []*poolRequest
	responses []*Response
	errors    []error
	mu        sync.Mutex
}

// poolRequest wraps a pending request for the pool.
type poolRequest struct {
	method string
	url    string
	data   map[string]any
	client *Client
}

// Pool creates a new request pool and executes the given function.
func (c *Client) Pool(fn func(p *Pool)) []*Response {
	pool := &Pool{
		requests:  make([]*poolRequest, 0),
		responses: make([]*Response, 0),
		errors:    make([]error, 0),
	}

	fn(pool)

	pool.execute()
	return pool.responses
}

// Add adds a request to the pool.
func (p *Pool) Add(resp *Response, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if err != nil {
		p.errors = append(p.errors, err)
	} else {
		p.responses = append(p.responses, resp)
	}
}

// execute runs all requests concurrently.
func (p *Pool) execute() {
	var wg sync.WaitGroup

	for _, req := range p.requests {
		wg.Add(1)
		go func(r *poolRequest) {
			defer wg.Done()
			resp, err := r.client.send(r.method, r.url, r.data)
			p.Add(resp, err)
		}(req)
	}

	wg.Wait()
}

// Responses returns all successful responses.
func (p *Pool) Responses() []*Response {
	return p.responses
}

// Errors returns all errors encountered.
func (p *Pool) Errors() []error {
	return p.errors
}
