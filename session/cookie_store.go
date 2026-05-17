package session

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

// CookieStore implements session storage using encrypted cookies.
type CookieStore struct {
	encryptionKey []byte
	signingKey    []byte
}

// NewCookieStore creates a new cookie-based session store.
// encryptionKey must be 16, 24, or 32 bytes for AES-128, AES-192, or AES-256.
func NewCookieStore(encryptionKey, signingKey []byte) (*CookieStore, error) {
	if len(encryptionKey) != 16 && len(encryptionKey) != 24 && len(encryptionKey) != 32 {
		return nil, fmt.Errorf("encryption key must be 16, 24, or 32 bytes")
	}

	if len(signingKey) == 0 {
		return nil, fmt.Errorf("signing key cannot be empty")
	}

	return &CookieStore{
		encryptionKey: encryptionKey,
		signingKey:    signingKey,
	}, nil
}

// Read decodes and decrypts session data from a cookie value.
func (c *CookieStore) Read(cookieValue string) (map[string]any, error) {
	if cookieValue == "" {
		return make(map[string]any), nil
	}

	// Decode base64
	data, err := base64.URLEncoding.DecodeString(cookieValue)
	if err != nil {
		return make(map[string]any), nil
	}

	// Verify HMAC
	if len(data) < sha256.Size {
		return make(map[string]any), nil
	}

	signature := data[:sha256.Size]
	payload := data[sha256.Size:]

	expectedMAC := c.sign(payload)
	if !hmac.Equal(signature, expectedMAC) {
		return make(map[string]any), nil
	}

	// Decrypt
	decrypted, err := c.decrypt(payload)
	if err != nil {
		return make(map[string]any), nil
	}

	// Unmarshal
	var wrapper struct {
		Data      map[string]any `json:"data"`
		ExpiresAt int64          `json:"expires_at"`
	}

	if err := json.Unmarshal(decrypted, &wrapper); err != nil {
		return make(map[string]any), nil
	}

	// Check expiration
	if wrapper.ExpiresAt > 0 && time.Now().Unix() > wrapper.ExpiresAt {
		return make(map[string]any), nil
	}

	return wrapper.Data, nil
}

// Write encrypts and encodes session data for storage in a cookie.
func (c *CookieStore) Write(id string, data map[string]any, ttl time.Duration) error {
	wrapper := struct {
		Data      map[string]any `json:"data"`
		ExpiresAt int64          `json:"expires_at"`
	}{
		Data:      data,
		ExpiresAt: time.Now().Add(ttl).Unix(),
	}

	jsonData, err := json.Marshal(wrapper)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	// Encrypt
	encrypted, err := c.encrypt(jsonData)
	if err != nil {
		return fmt.Errorf("failed to encrypt session data: %w", err)
	}

	// Sign
	signature := c.sign(encrypted)

	// Combine signature and encrypted data
	combined := append(signature, encrypted...)

	// Encode to base64
	_ = base64.URLEncoding.EncodeToString(combined)

	return nil
}

// Destroy is a no-op for cookie store (handled by expiring the cookie).
func (c *CookieStore) Destroy(id string) error {
	return nil
}

// GC is a no-op for cookie store (cookies expire naturally).
func (c *CookieStore) GC(maxLifetime time.Duration) error {
	return nil
}

// encrypt encrypts data using AES-GCM.
func (c *CookieStore) encrypt(plaintext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)
	return ciphertext, nil
}

// decrypt decrypts data using AES-GCM.
func (c *CookieStore) decrypt(ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.encryptionKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	return plaintext, nil
}

// sign creates an HMAC signature for the given data.
func (c *CookieStore) sign(data []byte) []byte {
	mac := hmac.New(sha256.New, c.signingKey)
	mac.Write(data)
	return mac.Sum(nil)
}
