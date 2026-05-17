// Package encryption provides AES-256-GCM authenticated encryption utilities.
package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

// Encrypter provides authenticated encryption using AES-256-GCM.
type Encrypter struct {
	key []byte
}

// NewEncrypter creates a new encrypter with the given key.
// The key must be exactly 32 bytes for AES-256.
func NewEncrypter(key []byte) (*Encrypter, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("key must be exactly 32 bytes, got %d", len(key))
	}

	return &Encrypter{
		key: key,
	}, nil
}

// Encrypt encrypts any value by JSON marshaling it first.
// Returns a base64-encoded string containing the nonce, ciphertext, and authentication tag.
func (e *Encrypter) Encrypt(value any) (string, error) {
	// Marshal value to JSON
	data, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("failed to marshal value: %w", err)
	}

	return e.EncryptString(string(data))
}

// EncryptString encrypts a plain string.
// Returns a base64-encoded string containing the nonce, ciphertext, and authentication tag.
// Format: base64(nonce + ciphertext + tag)
func (e *Encrypter) EncryptString(value string) (string, error) {
	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nonce, nonce, []byte(value), nil)

	// Encode as base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decrypts a base64-encoded payload and unmarshals it.
func (e *Encrypter) Decrypt(payload string) (any, error) {
	// Decrypt to string
	plaintext, err := e.DecryptString(payload)
	if err != nil {
		return nil, err
	}

	// Unmarshal JSON
	var result any
	if err := json.Unmarshal([]byte(plaintext), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal value: %w", err)
	}

	return result, nil
}

// DecryptString decrypts a base64-encoded payload to a plain string.
func (e *Encrypter) DecryptString(payload string) (string, error) {
	// Decode base64
	data, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Create AES cipher
	block, err := aes.NewCipher(e.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher: %w", err)
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Check minimum length
	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	// Extract nonce and ciphertext
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Decrypt and verify
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("decryption failed: %w", err)
	}

	return string(plaintext), nil
}

// GenerateKey generates a cryptographically secure random 32-byte key suitable for AES-256.
func GenerateKey() ([]byte, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}
	return key, nil
}

// Crypt is the package-level facade for encryption.
// It should be initialized with a key (typically from APP_KEY environment variable).
var Crypt *Encrypter
