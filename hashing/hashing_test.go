package hashing

import (
	"strings"
	"testing"
)

func TestSHA256Hasher_Make(t *testing.T) {
	hasher := NewSHA256Hasher(10000)

	password := "secret123"
	hash1, err := hasher.Make(password)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	hash2, err := hasher.Make(password)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	// Hashes should be different due to random salt
	if hash1 == hash2 {
		t.Error("Make() should produce different hashes for same input")
	}

	// Check format
	if !strings.HasPrefix(hash1, "$ignite$") {
		t.Errorf("Hash format incorrect, got: %s", hash1)
	}

	parts := strings.Split(hash1, "$")
	if len(parts) != 5 {
		t.Errorf("Hash should have 5 parts, got %d", len(parts))
	}
}

func TestSHA256Hasher_Check(t *testing.T) {
	hasher := NewSHA256Hasher(10000)

	password := "mypassword"
	hash, err := hasher.Make(password)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	tests := []struct {
		name     string
		value    string
		hash     string
		expected bool
	}{
		{
			name:     "correct password",
			value:    password,
			hash:     hash,
			expected: true,
		},
		{
			name:     "wrong password",
			value:    "wrongpassword",
			hash:     hash,
			expected: false,
		},
		{
			name:     "empty password",
			value:    "",
			hash:     hash,
			expected: false,
		},
		{
			name:     "invalid hash format",
			value:    password,
			hash:     "invalid",
			expected: false,
		},
		{
			name:     "malformed hash",
			value:    password,
			hash:     "$ignite$10000$salt",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasher.Check(tt.value, tt.hash)
			if result != tt.expected {
				t.Errorf("Check() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSHA256Hasher_NeedsRehash(t *testing.T) {
	hasher := NewSHA256Hasher(10000)

	tests := []struct {
		name     string
		hash     string
		expected bool
	}{
		{
			name:     "current iterations",
			hash:     "$ignite$10000$c2FsdA$aGFzaA",
			expected: false,
		},
		{
			name:     "old iterations",
			hash:     "$ignite$5000$c2FsdA$aGFzaA",
			expected: true,
		},
		{
			name:     "higher iterations",
			hash:     "$ignite$20000$c2FsdA$aGFzaA",
			expected: false,
		},
		{
			name:     "invalid format",
			hash:     "invalid",
			expected: true,
		},
		{
			name:     "wrong prefix",
			hash:     "$bcrypt$10$salt$hash",
			expected: true,
		},
		{
			name:     "malformed iterations",
			hash:     "$ignite$abc$salt$hash",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasher.NeedsRehash(tt.hash)
			if result != tt.expected {
				t.Errorf("NeedsRehash() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestSHA256Hasher_MinimumIterations(t *testing.T) {
	hasher := NewSHA256Hasher(500) // below minimum
	if hasher.Iterations < 1000 {
		t.Error("Iterations should be enforced to minimum 1000")
	}
}

func TestManager_DefaultDriver(t *testing.T) {
	manager := NewManager()

	password := "test123"
	hash, err := manager.Make(password)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if !manager.Check(password, hash) {
		t.Error("Check() should validate correct password")
	}

	if manager.Check("wrong", hash) {
		t.Error("Check() should reject wrong password")
	}
}

func TestManager_Register(t *testing.T) {
	manager := NewManager()

	// Register custom hasher
	customHasher := NewSHA256Hasher(15000)
	manager.Register("custom", customHasher)

	driver := manager.Driver("custom")
	if driver == nil {
		t.Fatal("Driver() should return registered hasher")
	}

	// Verify it's the custom hasher
	password := "test"
	hash, _ := driver.Make(password)
	if !strings.Contains(hash, "$15000$") {
		t.Error("Custom hasher should use 15000 iterations")
	}
}

func TestManager_SetDefault(t *testing.T) {
	manager := NewManager()

	customHasher := NewSHA256Hasher(20000)
	manager.Register("custom", customHasher)

	err := manager.SetDefault("custom")
	if err != nil {
		t.Fatalf("SetDefault() error = %v", err)
	}

	// Default driver should now be custom
	password := "test"
	hash, _ := manager.Make(password)
	if !strings.Contains(hash, "$20000$") {
		t.Error("Default hasher should now be custom with 20000 iterations")
	}

	// Try to set non-existent driver
	err = manager.SetDefault("nonexistent")
	if err == nil {
		t.Error("SetDefault() should error for non-existent driver")
	}
}

func TestManager_Driver_Fallback(t *testing.T) {
	manager := NewManager()

	// Non-existent driver should return default
	driver := manager.Driver("nonexistent")
	if driver == nil {
		t.Fatal("Driver() should return default for non-existent driver")
	}

	password := "test"
	hash, err := driver.Make(password)
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if !driver.Check(password, hash) {
		t.Error("Fallback driver should work correctly")
	}
}

func TestHash_GlobalFacade(t *testing.T) {
	password := "globaltest"

	hash, err := Hash.Make(password)
	if err != nil {
		t.Fatalf("Hash.Make() error = %v", err)
	}

	if !Hash.Check(password, hash) {
		t.Error("Hash.Check() should validate correct password")
	}

	if Hash.Check("wrong", hash) {
		t.Error("Hash.Check() should reject wrong password")
	}
}

func TestSHA256Hasher_EmptyPassword(t *testing.T) {
	hasher := NewSHA256Hasher(10000)

	hash, err := hasher.Make("")
	if err != nil {
		t.Fatalf("Make() error = %v", err)
	}

	if !hasher.Check("", hash) {
		t.Error("Should be able to hash and verify empty string")
	}

	if hasher.Check("notempty", hash) {
		t.Error("Non-empty string should not match empty password hash")
	}
}

func BenchmarkSHA256Hasher_Make(b *testing.B) {
	hasher := NewSHA256Hasher(10000)
	password := "benchmarkpassword"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := hasher.Make(password)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSHA256Hasher_Check(b *testing.B) {
	hasher := NewSHA256Hasher(10000)
	password := "benchmarkpassword"
	hash, _ := hasher.Make(password)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		hasher.Check(password, hash)
	}
}
