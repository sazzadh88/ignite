# Encryption Package

Authenticated encryption using AES-256-GCM for GoFrame.

## Features

- **Zero Dependencies**: Uses only Go standard library (crypto/aes, crypto/cipher)
- **Authenticated Encryption**: AES-256-GCM prevents tampering
- **Laravel-like API**: Familiar interface for Laravel developers
- **JSON Support**: Encrypt/decrypt any Go value via JSON serialization
- **Random Nonces**: Different ciphertext each time for same input

## Quick Start

```go
import "github.com/sazzad/goframe/encryption"

// Generate a key (do this once, store securely)
key, err := encryption.GenerateKey()

// Create encrypter
encrypter, err := encryption.NewEncrypter(key)

// Encrypt a string
encrypted, err := encrypter.EncryptString("Hello, World!")

// Decrypt
decrypted, err := encrypter.DecryptString(encrypted)
```

## Encrypt Complex Data

```go
// Encrypt any value (uses JSON)
data := map[string]any{
    "name":  "John Doe",
    "email": "john@example.com",
    "age":   30,
}

encrypted, err := encrypter.Encrypt(data)

// Decrypt
decrypted, err := encrypter.Decrypt(encrypted)
result := decrypted.(map[string]any)
```

## Key Management

```go
// Generate a key
key, _ := encryption.GenerateKey()

// Store as base64 (e.g., APP_KEY environment variable)
encoded := base64.StdEncoding.EncodeToString(key)

// Load key
decoded, _ := base64.StdEncoding.DecodeString(encoded)
encrypter, _ := encryption.NewEncrypter(decoded)
```

## Payload Format

Encrypted payloads are base64-encoded and contain:
- 12-byte nonce (GCM standard)
- Ciphertext
- 16-byte authentication tag

Format: `base64(nonce + ciphertext + tag)`

## Security

- **AES-256-GCM**: Industry-standard authenticated encryption
- **Random Nonces**: 12 bytes of cryptographic randomness per encryption
- **Authentication**: GCM mode prevents tampering and forgery
- **32-byte Keys**: Full AES-256 security

## API Reference

### Encrypter

```go
func NewEncrypter(key []byte) (*Encrypter, error)
func (e *Encrypter) Encrypt(value any) (string, error)
func (e *Encrypter) EncryptString(value string) (string, error)
func (e *Encrypter) Decrypt(payload string) (any, error)
func (e *Encrypter) DecryptString(payload string) (string, error)
```

### Key Generation

```go
func GenerateKey() ([]byte, error)
```

### Global Facade

```go
var Crypt *Encrypter
```

Initialize in your application:

```go
func init() {
    key, _ := base64.StdEncoding.DecodeString(os.Getenv("APP_KEY"))
    encryption.Crypt, _ = encryption.NewEncrypter(key)
}
```

## Performance

Benchmarks on Apple M4:
- EncryptString: ~0.54µs per operation
- DecryptString: ~0.28µs per operation
- Encrypt (with JSON): ~0.86µs per operation
- GenerateKey: ~0.19µs per operation

## Error Handling

```go
encrypted, err := encrypter.EncryptString("data")
if err != nil {
    // Handle encryption error
}

decrypted, err := encrypter.DecryptString(encrypted)
if err != nil {
    // Possible causes:
    // - Invalid base64
    // - Wrong key
    // - Tampered payload
    // - Corrupted data
}
```

## Examples

### Cookie Encryption

```go
// Encrypt cookie value
encryptedCookie, _ := encrypter.EncryptString(cookieValue)
http.SetCookie(w, &http.Cookie{
    Name:  "session",
    Value: encryptedCookie,
})

// Decrypt cookie
cookie, _ := r.Cookie("session")
decrypted, _ := encrypter.DecryptString(cookie.Value)
```

### Database Encryption

```go
// Encrypt before storing
encryptedSSN, _ := encrypter.EncryptString(user.SSN)
db.Exec("INSERT INTO users (ssn) VALUES (?)", encryptedSSN)

// Decrypt after retrieving
var encryptedSSN string
db.QueryRow("SELECT ssn FROM users WHERE id = ?", id).Scan(&encryptedSSN)
ssn, _ := encrypter.DecryptString(encryptedSSN)
```

### API Token Encryption

```go
// Encrypt sensitive tokens
token := "sk_live_1234567890abcdef"
encrypted, _ := encrypter.EncryptString(token)

// Store encrypted token
// Later, decrypt when needed
decrypted, _ := encrypter.DecryptString(encrypted)
```

## Notes

- **Key Rotation**: Generate new keys periodically. Re-encrypt data with new key.
- **Key Storage**: Never commit keys to version control. Use environment variables or key management systems.
- **32-byte Requirement**: AES-256 requires exactly 32 bytes. Use `GenerateKey()` or ensure your key is 32 bytes.
- **Nonce Uniqueness**: Random nonces ensure different ciphertexts for same plaintext. Never reuse nonces with same key.
