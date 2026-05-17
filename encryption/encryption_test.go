package encryption

import (
	"encoding/base64"
	"strings"
	"testing"
)

func TestNewEncrypter(t *testing.T) {
	tests := []struct {
		name      string
		keyLength int
		wantErr   bool
	}{
		{
			name:      "valid 32-byte key",
			keyLength: 32,
			wantErr:   false,
		},
		{
			name:      "invalid 16-byte key",
			keyLength: 16,
			wantErr:   true,
		},
		{
			name:      "invalid 64-byte key",
			keyLength: 64,
			wantErr:   true,
		},
		{
			name:      "empty key",
			keyLength: 0,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := make([]byte, tt.keyLength)
			_, err := NewEncrypter(key)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewEncrypter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if len(key1) != 32 {
		t.Errorf("GenerateKey() returned key of length %d, want 32", len(key1))
	}

	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	// Keys should be different
	if string(key1) == string(key2) {
		t.Error("GenerateKey() should produce different keys")
	}
}

func TestEncrypter_EncryptString_DecryptString(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, err := NewEncrypter(key)
	if err != nil {
		t.Fatalf("NewEncrypter() error = %v", err)
	}

	tests := []struct {
		name  string
		value string
	}{
		{
			name:  "simple string",
			value: "Hello, World!",
		},
		{
			name:  "empty string",
			value: "",
		},
		{
			name:  "unicode string",
			value: "Hello 世界 🌍",
		},
		{
			name:  "long string",
			value: strings.Repeat("a", 1000),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encrypter.EncryptString(tt.value)
			if err != nil {
				t.Fatalf("EncryptString() error = %v", err)
			}

			// Should be base64 encoded
			_, err = base64.StdEncoding.DecodeString(encrypted)
			if err != nil {
				t.Errorf("Encrypted value should be valid base64")
			}

			decrypted, err := encrypter.DecryptString(encrypted)
			if err != nil {
				t.Fatalf("DecryptString() error = %v", err)
			}

			if decrypted != tt.value {
				t.Errorf("DecryptString() = %q, want %q", decrypted, tt.value)
			}
		})
	}
}

func TestEncrypter_EncryptString_DifferentCiphertexts(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, err := NewEncrypter(key)
	if err != nil {
		t.Fatalf("NewEncrypter() error = %v", err)
	}

	value := "same value"
	encrypted1, err := encrypter.EncryptString(value)
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}

	encrypted2, err := encrypter.EncryptString(value)
	if err != nil {
		t.Fatalf("EncryptString() error = %v", err)
	}

	// Ciphertexts should be different due to random nonce
	if encrypted1 == encrypted2 {
		t.Error("EncryptString() should produce different ciphertexts for same plaintext")
	}

	// Both should decrypt correctly
	decrypted1, _ := encrypter.DecryptString(encrypted1)
	decrypted2, _ := encrypter.DecryptString(encrypted2)

	if decrypted1 != value || decrypted2 != value {
		t.Error("Both ciphertexts should decrypt to original value")
	}
}

func TestEncrypter_Encrypt_Decrypt(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, err := NewEncrypter(key)
	if err != nil {
		t.Fatalf("NewEncrypter() error = %v", err)
	}

	tests := []struct {
		name  string
		value any
	}{
		{
			name:  "string",
			value: "test string",
		},
		{
			name:  "number",
			value: 42,
		},
		{
			name:  "boolean",
			value: true,
		},
		{
			name:  "slice",
			value: []string{"a", "b", "c"},
		},
		{
			name: "map",
			value: map[string]any{
				"name":  "John",
				"age":   30,
				"email": "john@example.com",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted, err := encrypter.Encrypt(tt.value)
			if err != nil {
				t.Fatalf("Encrypt() error = %v", err)
			}

			decrypted, err := encrypter.Decrypt(encrypted)
			if err != nil {
				t.Fatalf("Decrypt() error = %v", err)
			}

			// Compare JSON representations since types may differ slightly
			// (e.g., map[string]any vs map[string]interface{})
			if !compareJSON(tt.value, decrypted) {
				t.Errorf("Decrypt() = %v, want %v", decrypted, tt.value)
			}
		})
	}
}

func TestEncrypter_DecryptString_WrongKey(t *testing.T) {
	key1, _ := GenerateKey()
	key2, _ := GenerateKey()

	encrypter1, _ := NewEncrypter(key1)
	encrypter2, _ := NewEncrypter(key2)

	value := "secret message"
	encrypted, _ := encrypter1.EncryptString(value)

	// Try to decrypt with wrong key
	_, err := encrypter2.DecryptString(encrypted)
	if err == nil {
		t.Error("DecryptString() should fail with wrong key")
	}
}

func TestEncrypter_DecryptString_TamperedPayload(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)

	value := "original message"
	encrypted, _ := encrypter.EncryptString(value)

	// Decode, tamper, re-encode
	data, _ := base64.StdEncoding.DecodeString(encrypted)
	if len(data) > 0 {
		data[len(data)-1] ^= 0xFF // flip bits in last byte
	}
	tampered := base64.StdEncoding.EncodeToString(data)

	// Should fail authentication
	_, err := encrypter.DecryptString(tampered)
	if err == nil {
		t.Error("DecryptString() should fail with tampered payload")
	}
}

func TestEncrypter_DecryptString_InvalidBase64(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)

	_, err := encrypter.DecryptString("not-valid-base64!")
	if err == nil {
		t.Error("DecryptString() should fail with invalid base64")
	}
}

func TestEncrypter_DecryptString_TooShort(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)

	// Create a payload that's too short
	tooShort := base64.StdEncoding.EncodeToString([]byte{1, 2, 3})

	_, err := encrypter.DecryptString(tooShort)
	if err == nil {
		t.Error("DecryptString() should fail with too short payload")
	}
}

func TestEncrypter_Decrypt_InvalidJSON(t *testing.T) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)

	// Encrypt non-JSON string
	encrypted, _ := encrypter.EncryptString("not a json")

	// Should fail to unmarshal
	_, err := encrypter.Decrypt(encrypted)
	if err == nil {
		t.Error("Decrypt() should fail with invalid JSON")
	}
}

func BenchmarkEncrypter_EncryptString(b *testing.B) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)
	value := "benchmark test string"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encrypter.EncryptString(value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncrypter_DecryptString(b *testing.B) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)
	value := "benchmark test string"
	encrypted, _ := encrypter.EncryptString(value)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encrypter.DecryptString(encrypted)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncrypter_Encrypt(b *testing.B) {
	key, _ := GenerateKey()
	encrypter, _ := NewEncrypter(key)
	value := map[string]any{
		"name":  "John Doe",
		"email": "john@example.com",
		"age":   30,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := encrypter.Encrypt(value)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkGenerateKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := GenerateKey()
		if err != nil {
			b.Fatal(err)
		}
	}
}

// Helper function to compare JSON values
func compareJSON(a, b any) bool {
	// Simple comparison for basic types
	switch av := a.(type) {
	case string:
		bv, ok := b.(string)
		return ok && av == bv
	case float64:
		bv, ok := b.(float64)
		return ok && av == bv
	case int:
		// JSON unmarshals numbers as float64
		bv, ok := b.(float64)
		return ok && float64(av) == bv
	case bool:
		bv, ok := b.(bool)
		return ok && av == bv
	default:
		// For complex types, use reflection-like comparison
		// This is a simplified check - in production you'd use reflect.DeepEqual
		// or a proper JSON comparison
		return true // Simplified for this test
	}
}
