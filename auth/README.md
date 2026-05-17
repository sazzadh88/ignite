# Auth Package

Laravel-inspired authentication system for GoFrame with support for multiple guards and user providers.

## Features

- **Multiple Guards**: Session-based and token-based authentication
- **User Providers**: Flexible user retrieval and validation
- **Middleware**: Ready-to-use authentication middleware
- **Password Reset**: Built-in password reset functionality
- **Facade Pattern**: Convenient package-level functions
- **Thread-Safe**: All components are safe for concurrent use
- **Zero Dependencies**: Only uses Go standard library

## Installation

The auth package is part of GoFrame:

```go
import "github.com/sazzad/goframe/auth"
```

## Quick Start

### 1. Define Your User Model

```go
type User struct {
    ID       int
    Email    string
    Password string
}

func (u *User) GetAuthIdentifier() any {
    return u.ID
}

func (u *User) GetAuthPassword() string {
    return u.Password
}
```

### 2. Create a User Provider

```go
provider := auth.NewCallbackProvider(
    func(id any) (auth.Authenticatable, error) {
        // Retrieve user from database by ID
        return db.FindUserByID(id)
    },
    func(credentials map[string]string) (auth.Authenticatable, error) {
        // Retrieve user by email/username
        return db.FindUserByEmail(credentials["email"])
    },
    func(user auth.Authenticatable, credentials map[string]string) bool {
        // Validate password (use bcrypt in production)
        return bcrypt.CompareHashAndPassword(
            []byte(user.GetAuthPassword()),
            []byte(credentials["password"]),
        ) == nil
    },
)
```

### 3. Setup Guard and Manager

```go
// Create session (implement auth.Session interface)
session := mySessionImplementation

// Create guard
guard := auth.NewSessionGuard(provider, session)

// Setup manager
manager := auth.GetManager()
manager.AddGuard("session", guard)
manager.SetDefaultGuard("session")
```

### 4. Use Authentication

```go
// Attempt login
if auth.Attempt(map[string]string{
    "email":    "user@example.com",
    "password": "secret",
}) {
    // User is authenticated
    user := auth.User()
    userID := auth.ID()
}

// Check authentication status
if auth.Check() {
    // User is authenticated
}

if auth.Guest() {
    // User is not authenticated
}

// Logout
auth.Logout()
```

## Session-Based Authentication

Session guards store authentication state in sessions (cookies).

```go
session := mySessionImplementation
guard := auth.NewSessionGuard(provider, session)

// Attempt login (validates and stores in session)
guard.Attempt(map[string]string{
    "email":    "user@example.com",
    "password": "secret",
})

// Login without validation
guard.Login(user)

// Login by user ID
guard.LoginUsingID(1)

// Logout
guard.Logout()

// Stateless authentication (one request)
guard.Once(credentials)
```

## Token-Based Authentication

Token guards use Bearer tokens for API authentication.

```go
storage := auth.NewMemoryTokenStorage()
guard := auth.NewTokenGuard(provider, storage)

// Create token for user
token, err := guard.CreateToken(user, "api-token", []string{"read", "write"})
if err != nil {
    log.Fatal(err)
}

// Use token.PlainTextToken in API requests
// Authorization: Bearer {token.PlainTextToken}

// Authenticate with token
err = guard.SetToken(plainTextToken)
if err == nil && guard.Check() {
    // User is authenticated
}

// Revoke token
guard.RevokeToken(token.ID)
```

## Middleware

### Require Authentication

```go
// Redirect to /login if not authenticated
authMiddleware := auth.Authenticate()

// Or abort with 401
authMiddleware := auth.Authenticate().SetAbortOnFail(true)

// Use in your routing
router.Use(authMiddleware)
```

### Guest Only

```go
// Redirect to /home if authenticated (for login/register pages)
guestMiddleware := auth.RedirectIfAuthenticated()

router.Use(guestMiddleware)
```

### Token Authentication

```go
tokenMiddleware := auth.NewTokenAuthMiddleware(tokenGuard)

// Extracts Bearer token from Authorization header
router.Use(tokenMiddleware)
```

## Password Reset

```go
storage := auth.NewMemoryPasswordResetStorage()
resetter := myPasswordResetter // Implement auth.PasswordResetter
broker := auth.NewPasswordBroker(provider, storage, resetter)

// Send reset link (generates token and stores it)
err := broker.SendResetLink("user@example.com")

// Reset password with token
err = broker.Reset(token, "user@example.com", "new_password")

// Set token lifetime (default 60 minutes)
broker.SetTokenLife(30 * time.Minute)
```

## Multiple Guards

```go
manager := auth.NewManager()

// Session guard
sessionGuard := auth.NewSessionGuard(provider, session)
manager.AddGuard("session", sessionGuard)

// Token guard
tokenGuard := auth.NewTokenGuard(provider, storage)
manager.AddGuard("api", tokenGuard)

// Set default
manager.SetDefaultGuard("session")

// Use specific guard
webGuard := manager.Guard("session")
apiGuard := manager.Guard("api")
```

## Custom Guards

Implement the `Guard` interface:

```go
type CustomGuard struct {
    // Your fields
}

func (g *CustomGuard) User() auth.Authenticatable {
    // Return current user
}

func (g *CustomGuard) ID() any {
    // Return user ID
}

func (g *CustomGuard) Check() bool {
    return g.User() != nil
}

func (g *CustomGuard) Guest() bool {
    return !g.Check()
}

func (g *CustomGuard) Validate(credentials map[string]string) bool {
    // Validate credentials without logging in
}
```

## Testing

The package includes comprehensive tests:

```bash
go test ./auth/... -v
```

Run examples:

```bash
go test ./auth/... -run Example -v
```

## Production Considerations

### Security

1. **Password Hashing**: Use bcrypt or argon2 for password hashing
2. **Token Security**: Use crypto/sha256 for token hashing
3. **HTTPS**: Always use HTTPS in production
4. **Session Security**: Set secure cookie flags (HttpOnly, Secure, SameSite)
5. **Rate Limiting**: Implement rate limiting for login attempts
6. **Token Expiration**: Set expiration times for tokens

### Example with bcrypt

```go
import "golang.org/x/crypto/bcrypt"

provider := auth.NewCallbackProvider(
    retrieveByID,
    retrieveByCredentials,
    func(user auth.Authenticatable, credentials map[string]string) bool {
        password := credentials["password"]
        err := bcrypt.CompareHashAndPassword(
            []byte(user.GetAuthPassword()),
            []byte(password),
        )
        return err == nil
    },
)
```

### Database Integration

For production, implement `UserProvider` with database queries:

```go
type DatabaseProvider struct {
    db *sql.DB
}

func (p *DatabaseProvider) RetrieveByID(id any) (auth.Authenticatable, error) {
    var user User
    err := p.db.QueryRow("SELECT * FROM users WHERE id = ?", id).Scan(&user)
    return &user, err
}

func (p *DatabaseProvider) RetrieveByCredentials(credentials map[string]string) (auth.Authenticatable, error) {
    email := credentials["email"]
    var user User
    err := p.db.QueryRow("SELECT * FROM users WHERE email = ?", email).Scan(&user)
    return &user, err
}

func (p *DatabaseProvider) ValidateCredentials(user auth.Authenticatable, credentials map[string]string) bool {
    password := credentials["password"]
    return bcrypt.CompareHashAndPassword(
        []byte(user.GetAuthPassword()),
        []byte(password),
    ) == nil
}
```

## API Reference

See [GoDoc](https://pkg.go.dev/github.com/sazzad/goframe/auth) for complete API documentation.

## License

Part of GoFrame framework.
