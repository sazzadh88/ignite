package hashing

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
)

// SHA256Hasher implements password hashing using HMAC-SHA256 with salt and iterations.
// Format: $ignite$iterations$salt$hash
type SHA256Hasher struct {
	Iterations int
}

// NewSHA256Hasher creates a new SHA256 hasher with the specified iteration count.
func NewSHA256Hasher(iterations int) *SHA256Hasher {
	if iterations < 1000 {
		iterations = 10000 // minimum secure iterations
	}
	return &SHA256Hasher{
		Iterations: iterations,
	}
}

// Make generates a hash for the given value.
func (h *SHA256Hasher) Make(value string) (string, error) {
	// Generate random salt
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive hash using PBKDF2-like approach with HMAC-SHA256
	hash := h.deriveKey([]byte(value), salt, h.Iterations)

	// Encode as: $ignite$iterations$salt$hash
	saltEncoded := base64.RawStdEncoding.EncodeToString(salt)
	hashEncoded := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$ignite$%d$%s$%s", h.Iterations, saltEncoded, hashEncoded), nil
}

// Check validates a plain-text value against its hash.
func (h *SHA256Hasher) Check(value, hash string) bool {
	// Parse the hash format
	parts := strings.Split(hash, "$")
	if len(parts) != 5 || parts[0] != "" || parts[1] != "ignite" {
		return false
	}

	iterations, err := strconv.Atoi(parts[2])
	if err != nil {
		return false
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false
	}

	// Derive the key with the same parameters
	actualHash := h.deriveKey([]byte(value), salt, iterations)

	// Constant-time comparison to prevent timing attacks
	return subtle.ConstantTimeCompare(expectedHash, actualHash) == 1
}

// NeedsRehash checks if the hash needs to be regenerated.
func (h *SHA256Hasher) NeedsRehash(hash string) bool {
	parts := strings.Split(hash, "$")
	if len(parts) != 5 || parts[0] != "" || parts[1] != "ignite" {
		return true
	}

	iterations, err := strconv.Atoi(parts[2])
	if err != nil {
		return true
	}

	return iterations < h.Iterations
}

// deriveKey performs PBKDF2-like key derivation using HMAC-SHA256.
func (h *SHA256Hasher) deriveKey(password, salt []byte, iterations int) []byte {
	mac := hmac.New(sha256.New, password)
	mac.Write(salt)
	dk := mac.Sum(nil)

	// Iterate to strengthen the hash
	for i := 1; i < iterations; i++ {
		mac.Reset()
		mac.Write(dk)
		dk = mac.Sum(nil)
	}

	return dk
}
