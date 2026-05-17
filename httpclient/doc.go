// Package httpclient provides a Laravel-inspired HTTP client with a fluent API.
//
// The package offers a simple, expressive interface for making HTTP requests
// with support for authentication, retries, timeouts, and testing.
//
// Basic Usage:
//
//	resp, err := httpclient.Http.Get("https://api.example.com/users")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	fmt.Println(resp.Body())
//
// Fluent Configuration:
//
//	client := httpclient.New().
//	    WithToken("secret-token").
//	    Timeout(30 * time.Second).
//	    Retry(3, 100 * time.Millisecond).
//	    AcceptJSON()
//
//	resp, err := client.Get("https://api.example.com/data")
//
// JSON Requests:
//
//	resp, err := httpclient.New().
//	    AsJSON().
//	    Post("https://api.example.com/users", map[string]any{
//	        "name": "John Doe",
//	        "email": "john@example.com",
//	    })
//
// Response Helpers:
//
//	if resp.Successful() {
//	    data, err := resp.JSON()
//	    // Process JSON response
//	}
//
//	if resp.NotFound() {
//	    // Handle 404
//	}
//
// Testing with Fakes:
//
//	fake := httpclient.Fake(
//	    httpclient.FakeResponse(map[string]string{"status": "ok"}, 200),
//	)
//
//	resp, _ := fake.Get("https://api.example.com/test")
//	fake.AssertSent(func(req *http.Request) bool {
//	    return req.Method == "GET"
//	})
package httpclient
