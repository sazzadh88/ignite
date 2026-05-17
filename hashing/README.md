# Hashing Package

Secure password hashing and verification for GoFrame.

## Features

- **Zero Dependencies**: Uses only Go standard library (crypto/*, encoding/*)
- **Secure by Default**: HMAC-SHA256 with salt and configurable iterations
- **Laravel-like API**: Familiar interface for Laravel developers
- **Pluggable**: Register custom hashers via the Hasher interface
- **Constant-time Comparison**: Prevents timing attacks

## Quick Start

```go
import "github.com/sazzad/goframe/hashing"

// Hash a password
hash, err := hashing.Hash.Make("secret123")

// Verify password
if hashing.Hash.Check("secret123", hash) {
    // Password is correct
}

// Check if rehashing needed (e.g., iterations increased)
if hashing.Hash.NeedsRehash(hash) {
    newHash, _ := hashing.Hash.Make("secret123")
    // Update stored hash
}
```

## Hash Format

Hashes use the format: `$goframe$iterations$salt$hash`

Example: `$goframe$10000$randomBase64Salt$derivedBase64Hash`

## Custom Hashers

```go
// Create hasher with custom iterations
strongHasher := hashing.NewSHA256Hasher(50000)

// Register as driver
hashing.Hash.Register("strong", strongHasher)

// Use specific driver
hash, _ := hashing.Hash.Driver("strong").Make("password")

// Set as default
hashing.Hash.SetDefault("strong")
```

## Security

- Minimum 10,000 iterations (enforced)
- 16-byte random salt per hash
- HMAC-SHA256 key derivation (PBKDF2-like)
- Constant-time comparison prevents timing attacks
- Random nonce ensures different hashes for same input

## API Reference

### Hasher Interface

```go
type Hasher interface {
    Make(value string) (string, error)
    Check(value, hash string) bool
    NeedsRehash(hash string) bool
}
```

### Manager

```go
func NewManager() *Manager
func (m *Manager) Register(name string, hasher Hasher)
func (m *Manager) Driver(name string) Hasher
func (m *Manager) SetDefault(name string) error
func (m *Manager) Make(value string) (string, error)
func (m *Manager) Check(value, hash string) bool
func (m *Manager) NeedsRehash(hash string) bool
```

### SHA256Hasher

```go
func NewSHA256Hasher(iterations int) *SHA256Hasher
```

## Performance

Benchmarks on Apple M4:
- Make: ~0.87ms per operation (10,000 iterations)
- Check: ~0.87ms per operation

Adjust iterations based on your security/performance requirements.

## Extending

Implement the `Hasher` interface for custom algorithms:

```go
type MyHasher struct{}

func (h *MyHasher) Make(value string) (string, error) {
    // Your implementation
}

func (h *MyHasher) Check(value, hash string) bool {
    // Your implementation
}

func (h *MyHasher) NeedsRehash(hash string) bool {
    // Your implementation
}

// Register
hashing.Hash.Register("custom", &MyHasher{})
```

## Notes

- **No bcrypt/argon2**: Zero dependencies means stdlib only. The SHA256-HMAC hasher is secure for most use cases.
- **Pluggable Design**: Add bcrypt/argon2 via the Hasher interface if needed.
- **Production Ready**: Constant-time comparison, secure randomness, configurable work factor.
