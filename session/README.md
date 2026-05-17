# Session Package

A Laravel-inspired session management package for GoFrame with zero external dependencies.

## Features

- Multiple storage drivers (memory, file, cookie)
- Flash data support
- Session regeneration and invalidation
- CSRF token generation
- Thread-safe operations
- Middleware support
- Type-safe getters (GetString, GetInt)
- Array operations (Push)
- Counter operations (Increment, Decrement)
- Garbage collection

## Usage

### Basic Setup

```go
import "github.com/sazzad/goframe/session"

// Create session manager
config := session.DefaultConfig()
config.Driver = "file"
config.Files = "/tmp/sessions"

manager, err := session.NewManager(config)
if err != nil {
    panic(err)
}
```

### Middleware Integration

```go
// Use with HTTP router
handler := session.StartSession(manager)(yourHandler)
http.Handle("/", handler)
```

### Working with Sessions

```go
func handler(w http.ResponseWriter, r *http.Request) {
    // Get session from request
    sess := session.FromRequest(r)
    
    // Store data
    sess.Put("user_id", 123)
    sess.Put("username", "john")
    
    // Retrieve data
    userID := sess.GetInt("user_id")
    username := sess.GetString("username", "guest")
    
    // Check existence
    if sess.Has("user_id") {
        // User is logged in
    }
    
    // Remove data
    sess.Forget("temp_data")
    
    // Clear all data
    sess.Flush()
}
```

### Flash Data

Flash data is available only for the next request:

```go
// Set flash data
sess.Flash("message", "Record saved successfully!")

// In the next request:
message := sess.GetString("message") // "Record saved successfully!"

// After that request, the flash data is gone
```

### Session Regeneration

```go
// Regenerate session ID (useful after login)
newID := sess.Regenerate()

// Invalidate session (clear data and regenerate)
newID := sess.Invalidate()
```

### CSRF Protection

```go
// Get CSRF token
token := sess.Token()

// Use in forms
fmt.Fprintf(w, `<input type="hidden" name="_token" value="%s">`, token)
```

### Counter Operations

```go
// Increment page views
sess.Increment("page_views")
sess.Increment("page_views", 5) // Increment by 5

// Decrement
sess.Decrement("counter")
sess.Decrement("counter", 3) // Decrement by 3
```

### Array Operations

```go
// Build arrays
sess.Push("cart", "item1")
sess.Push("cart", "item2")

cart := sess.Get("cart").([]any)
```

## Storage Drivers

### Memory Store (Default)

In-memory storage, ideal for development:

```go
config := session.DefaultConfig()
config.Driver = "memory"
```

### File Store

Stores sessions as JSON files:

```go
config := session.DefaultConfig()
config.Driver = "file"
config.Files = "/var/sessions"
```

### Cookie Store

Stores encrypted session data in cookies:

```go
config := session.DefaultConfig()
config.Driver = "cookie"
config.EncryptionKey = []byte("32-byte-encryption-key-here!")
config.SigningKey = []byte("signing-key")
```

## Configuration

```go
type SessionConfig struct {
    Driver        string            // "file", "memory", or "cookie"
    CookieName    string            // Cookie name (default: "session")
    Lifetime      int               // Session lifetime in minutes
    Path          string            // Cookie path
    Domain        string            // Cookie domain
    Secure        bool              // HTTPS only
    HTTPOnly      bool              // HTTP only (no JavaScript access)
    SameSite      http.SameSite     // SameSite attribute
    Files         string            // Directory for file-based sessions
    EncryptionKey []byte            // For cookie encryption
    SigningKey    []byte            // For cookie signing
}
```

## Garbage Collection

Run periodic garbage collection to remove expired sessions:

```go
// Run GC every hour
ticker := time.NewTicker(time.Hour)
go func() {
    for range ticker.C {
        manager.GC()
    }
}()
```

## Testing

Run tests:

```bash
go test ./session/
```

All tests include:
- Session data operations
- Flash data lifecycle
- Store implementations
- Middleware integration
- Thread safety
