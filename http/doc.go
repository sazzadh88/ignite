// Package http provides Laravel-inspired HTTP request and response abstractions for Ignite.
//
// This package mirrors Laravel's Request and Response objects, providing a clean and intuitive
// API for handling HTTP interactions in Go web applications.
//
// # Request
//
// The Request type wraps Go's *http.Request with convenient methods for accessing input data:
//
//	req := http.NewRequest(r)
//	name := req.InputString("name")
//	age := req.InputInt("age", 18)
//	email := req.Query("email", "default@example.com")
//
// # Response
//
// The Response type provides a fluent interface for building HTTP responses:
//
//	resp := http.JSON(map[string]string{"status": "ok"}, http.StatusOK)
//	resp.Header("X-Custom", "Value").Send(w)
//
// # Context
//
// The Context type ties together Request, ResponseWriter, and route parameters:
//
//	ctx := http.NewContext(w, r)
//	ctx.SetParam("id", "123")
//	ctx.JSON(http.StatusOK, map[string]string{"id": ctx.Param("id")})
//
// # File Uploads
//
// Handle file uploads with the UploadedFile type:
//
//	file := req.File("avatar")
//	if file != nil && file.IsValid() {
//	    path, err := file.Store("/storage/uploads")
//	}
package http
