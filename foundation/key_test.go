package foundation

import (
	"encoding/base64"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseAppKey(t *testing.T) {
	valid := "base64:" + base64.StdEncoding.EncodeToString([]byte("01234567890123456789012345678901"))
	if b, err := parseAppKey(valid); err != nil || len(b) != 32 {
		t.Fatalf("valid base64 key: got len=%d err=%v", len(b), err)
	}

	if _, err := parseAppKey(""); err == nil {
		t.Error("empty key should error")
	}

	// The placeholder shipped in scaffolded .env decodes to 24 bytes.
	if _, err := parseAppKey("base64:1234567890abcdef1234567890abcdef"); err == nil {
		t.Error("short base64 key should error")
	}

	if _, err := parseAppKey("base64:not valid base64!!"); err == nil {
		t.Error("invalid base64 should error")
	}

	if b, err := parseAppKey("01234567890123456789012345678901"); err != nil || len(b) != 32 {
		t.Errorf("raw 32-byte key should be accepted: err=%v", err)
	}

	if _, err := parseAppKey("tooshort"); err == nil {
		t.Error("raw key of wrong length should error")
	}
}

func TestGenerateAppKey(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	original := "APP_NAME=test\nAPP_KEY=\nAPP_DEBUG=true\n"
	if err := os.WriteFile(envPath, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}

	key, err := GenerateAppKey(dir)
	if err != nil {
		t.Fatalf("GenerateAppKey: %v", err)
	}

	if !strings.HasPrefix(key, "base64:") {
		t.Errorf("key should be base64-prefixed, got %q", key)
	}
	if b, err := parseAppKey(key); err != nil || len(b) != 32 {
		t.Errorf("generated key must parse to 32 bytes: err=%v", err)
	}

	data, _ := os.ReadFile(envPath)
	content := string(data)
	if !strings.Contains(content, "APP_KEY="+key) {
		t.Errorf("APP_KEY line not written: %s", content)
	}
	if !strings.Contains(content, "APP_NAME=test") || !strings.Contains(content, "APP_DEBUG=true") {
		t.Errorf("other .env lines must be preserved: %s", content)
	}
	if strings.Count(content, "APP_KEY=") != 1 {
		t.Errorf("APP_KEY should be replaced, not duplicated: %s", content)
	}
}

func TestSetEnvValueAppendsWhenMissing(t *testing.T) {
	dir := t.TempDir()
	envPath := filepath.Join(dir, ".env")
	if err := os.WriteFile(envPath, []byte("APP_NAME=test\n"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := setEnvValue(envPath, "APP_KEY", "base64:xyz"); err != nil {
		t.Fatal(err)
	}

	data, _ := os.ReadFile(envPath)
	if !strings.Contains(string(data), "APP_KEY=base64:xyz") {
		t.Errorf("APP_KEY should be appended: %s", string(data))
	}
}

func TestSetEnvValueMissingFile(t *testing.T) {
	if err := setEnvValue(filepath.Join(t.TempDir(), ".env"), "APP_KEY", "x"); err == nil {
		t.Error("missing .env file should error")
	}
}
