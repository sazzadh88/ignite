package http_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"

	gohttp "github.com/sazzad/ignite/http"
)

// Example_basicUsage demonstrates basic request and response handling
func Example_basicUsage() {
	// Create a test request
	r := httptest.NewRequest("GET", "/users?name=John&age=25", nil)
	w := httptest.NewRecorder()

	// Wrap with Ignite Request
	req := gohttp.NewRequest(r)

	// Access input
	name := req.InputString("name")
	age := req.InputInt("age", 0)

	fmt.Printf("Name: %s, Age: %d\n", name, age)

	// Create JSON response
	resp := gohttp.JSON(map[string]any{
		"name": name,
		"age":  age,
	}, http.StatusOK)

	// Send response
	resp.Send(w)

	// Output:
	// Name: John, Age: 25
}

// Example_contextAPI demonstrates the Context API
func Example_contextAPI() {
	// Create test request
	r := httptest.NewRequest("GET", "/users/123", nil)
	w := httptest.NewRecorder()

	// Create context
	ctx := gohttp.NewContext(w, r)
	ctx.SetParam("id", "123")

	// Use context to respond
	userID := ctx.Param("id")
	fmt.Printf("User ID: %s\n", userID)

	// Output:
	// User ID: 123
}

// Example_inputFiltering demonstrates input filtering
func Example_inputFiltering() {
	r := httptest.NewRequest("GET", "/users?name=John&email=john@example.com&password=secret&admin=true", nil)
	req := gohttp.NewRequest(r)

	// Get only safe fields
	safeInput := req.Only("name", "email")
	fmt.Printf("Safe input: name=%v, email=%v\n", safeInput["name"], safeInput["email"])

	// Exclude sensitive fields
	publicData := req.Except("password", "admin")
	fmt.Printf("Public data contains password: %v\n", publicData["password"] != nil)

	// Output:
	// Safe input: name=John, email=john@example.com
	// Public data contains password: false
}

// Example_inputValidation demonstrates checking input presence
func Example_inputValidation() {
	r := httptest.NewRequest("GET", "/users?name=John&empty=", nil)
	req := gohttp.NewRequest(r)

	fmt.Printf("Has 'name': %v\n", req.Has("name"))
	fmt.Printf("Filled 'name': %v\n", req.Filled("name"))
	fmt.Printf("Has 'empty': %v\n", req.Has("empty"))
	fmt.Printf("Filled 'empty': %v\n", req.Filled("empty"))
	fmt.Printf("Missing 'email': %v\n", req.Missing("email"))

	// Output:
	// Has 'name': true
	// Filled 'name': true
	// Has 'empty': true
	// Filled 'empty': false
	// Missing 'email': true
}

// Example_responseChaining demonstrates fluent response building
func Example_responseChaining() {
	w := httptest.NewRecorder()

	// Build response with chaining
	resp := gohttp.JSON(map[string]string{"status": "ok"}, http.StatusOK).
		Header("X-API-Version", "v1").
		Header("X-Request-ID", "abc123")

	resp.Send(w)

	result := w.Result()
	fmt.Printf("Status: %d\n", result.StatusCode)
	fmt.Printf("X-API-Version: %s\n", result.Header.Get("X-API-Version"))
	fmt.Printf("Content-Type: %s\n", result.Header.Get("Content-Type"))

	// Output:
	// Status: 200
	// X-API-Version: v1
	// Content-Type: application/json
}

// Example_contextState demonstrates context state management
func Example_contextState() {
	r := httptest.NewRequest("GET", "/profile", nil)
	w := httptest.NewRecorder()

	ctx := gohttp.NewContext(w, r)

	// Store user in context
	ctx.Set("user", map[string]any{
		"id":   123,
		"name": "John Doe",
	})

	// Retrieve user
	if user, exists := ctx.Get("user"); exists {
		userMap := user.(map[string]any)
		fmt.Printf("User: %s (ID: %d)\n", userMap["name"], userMap["id"])
	}

	// Output:
	// User: John Doe (ID: 123)
}
