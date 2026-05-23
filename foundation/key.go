package foundation

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sazzadh88/ignite/encryption"
)

// missingAppKeyError is the message shown, Laravel-style, when the
// application key is absent or invalid and is actually required.
const missingAppKeyError = "No application encryption key has been specified. Run 'go run main.go key:generate' to generate one."

// GenerateAppKey generates a fresh 32-byte key, base64-encodes it with the
// "base64:" prefix (matching Laravel's APP_KEY format), writes it to the
// .env file at basePath, and returns the generated value.
func GenerateAppKey(basePath string) (string, error) {
	keyBytes, err := encryption.GenerateKey()
	if err != nil {
		return "", err
	}

	key := "base64:" + base64.StdEncoding.EncodeToString(keyBytes)

	envPath := filepath.Join(basePath, ".env")
	if err := setEnvValue(envPath, "APP_KEY", key); err != nil {
		return "", err
	}

	return key, nil
}

// parseAppKey decodes a raw APP_KEY value into 32 bytes suitable for AES-256.
// It accepts both the "base64:" prefixed form and a raw 32-byte string.
func parseAppKey(raw string) ([]byte, error) {
	if raw == "" {
		return nil, fmt.Errorf("application key is empty")
	}

	if strings.HasPrefix(raw, "base64:") {
		decoded, err := base64.StdEncoding.DecodeString(strings.TrimPrefix(raw, "base64:"))
		if err != nil {
			return nil, fmt.Errorf("application key is not valid base64: %w", err)
		}
		if len(decoded) != 32 {
			return nil, fmt.Errorf("application key must decode to 32 bytes, got %d", len(decoded))
		}
		return decoded, nil
	}

	if len(raw) != 32 {
		return nil, fmt.Errorf("application key must be 32 bytes, got %d", len(raw))
	}
	return []byte(raw), nil
}

// setEnvValue replaces the value of key in the .env file at path, or appends
// the key if it is not already present. Other lines are preserved verbatim.
func setEnvValue(path, key, value string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("no .env file found at %s", path)
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	replaced := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, key+"=") {
			lines[i] = key + "=" + value
			replaced = true
			break
		}
	}

	if !replaced {
		lines = append(lines, key+"="+value)
	}

	return os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
}
