# HTTP Package

Laravel-inspired HTTP request and response abstractions for GoFrame.

## Overview

This package provides a clean, intuitive API for handling HTTP interactions in Go, mirroring Laravel's Request and Response objects.

## Features

### Request
- **Input Retrieval**: `Input()`, `InputString()`, `InputInt()` - Get data from query, form, or JSON
- **Filtering**: `Only()`, `Except()` - Select or exclude specific keys
- **Validation Helpers**: `Has()`, `Filled()`, `Missing()` - Check input presence
- **Headers & Metadata**: `Header()`, `IP()`, `Method()`, `Path()`, `UserAgent()`
- **File Uploads**: `File()`, `Files()` - Handle multipart file uploads
- **Authentication**: `BearerToken()` - Extract JWT/Bearer tokens
- **Content Negotiation**: `WantsJSON()`, `Ajax()`, `Secure()`

### Response
- **JSON Responses**: `JSON(data, status)`
- **Redirects**: `Redirect(url)`, `RedirectBack()`
- **File Downloads**: `Download(path)`, `File(path)`
- **Status Codes**: `NoContent()`, `Status(code)`
- **Chainable Headers**: `Header(key, val)`, `Cookie(cookie)`

### Context
- **Unified API**: Combines Request and ResponseWriter
- **Route Parameters**: `Param(key)`, `SetParam(key, val)`
- **Quick Responses**: `JSON()`, `String()`, `HTML()`, `Redirect()`
- **State Management**: `Set()`, `Get()`, `MustGet()`
- **Middleware Support**: `Abort()`, `IsAborted()`

### UploadedFile
- **Metadata**: `GetClientOriginalName()`, `GetMimeType()`, `GetSize()`
- **Validation**: `IsValid()`, `IsImage()`
- **Storage**: `Store(path)`, `StoreAs(path, name)`

## Usage Examples

### Basic Request Handling

```go
// Create request wrapper
req := http.NewRequest(r)

// Get input values
name := req.InputString("name")
age := req.InputInt("age", 18) // with default
email := req.Query("email")

// Check input
if req.Has("email") && req.Filled("email") {
    // Process email
}

// Get filtered input
credentials := req.Only("email", "password")
safeData := req.Except("password", "token")
```

### Response Building

```go
// JSON response
resp := http.JSON(map[string]any{
    "user": user,
    "token": token,
}, http.StatusCreated)

resp.Header("X-API-Version", "v1").
    Cookie(&http.Cookie{Name: "session", Value: "abc123"}).
    Send(w)

// File download
http.Download("/path/to/file.pdf", "invoice.pdf").Send(w)

// Redirect
http.Redirect("/dashboard", http.StatusSeeOther).Send(w)
```

### Context API

```go
ctx := http.NewContext(w, r)
ctx.SetParam("id", "123")

// Quick JSON response
ctx.JSON(http.StatusOK, map[string]string{
    "id": ctx.Param("id"),
    "status": "active",
})

// Store data in context
ctx.Set("user", currentUser)
user := ctx.MustGet("user")

// Redirect
ctx.Redirect("/home")
ctx.Back("/fallback") // Redirect to referer
```

### File Uploads

```go
file := req.File("avatar")
if file != nil && file.IsValid() {
    if file.IsImage() {
        path, err := file.StoreAs("/storage/avatars", "user-123.jpg")
        if err != nil {
            // Handle error
        }
    }
}

// Multiple files
files := req.Files("documents")
for _, file := range files {
    file.Store("/storage/documents")
}
```

## Testing

The package includes comprehensive tests covering all major functionality:

```bash
go test ./http/... -v
go test ./http/... -cover
```

Current test coverage: **89.4%**

## Design Decisions

1. **Zero External Dependencies**: Uses only Go standard library
2. **Laravel Compatibility**: API mirrors Laravel's Request/Response as closely as possible
3. **Type Safety**: All public types have full GoDoc comments
4. **Fluent Interface**: Chainable methods for response building
5. **Graceful Degradation**: Methods handle missing data gracefully with defaults

## Integration

This package is part of GoFrame (github.com/sazzad/goframe) and integrates seamlessly with:

- **Routing**: Route parameters via `Context.Param()`
- **Middleware**: Abort chain with `Context.Abort()`
- **Service Container**: Access services via `Context.Get/Set`
- **Validation**: Placeholder `Validate()` for future validation package

## License

Part of the GoFrame framework.
