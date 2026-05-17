package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSetAndGet(t *testing.T) {
	r := New()
	r.Set("app.name", "Ignite")
	if r.Get("app.name") != "Ignite" {
		t.Error("set/get failed")
	}
}

func TestGetString(t *testing.T) {
	r := New()
	r.Set("app.name", "Ignite")
	if r.GetString("app.name") != "Ignite" {
		t.Error("GetString failed")
	}
	if r.GetString("missing", "default") != "default" {
		t.Error("GetString default failed")
	}
}

func TestGetInt(t *testing.T) {
	r := New()
	r.Set("app.port", 8080)
	if r.GetInt("app.port") != 8080 {
		t.Error("GetInt failed")
	}
	if r.GetInt("missing", 3000) != 3000 {
		t.Error("GetInt default failed")
	}
}

func TestGetBool(t *testing.T) {
	r := New()
	r.Set("app.debug", true)
	if !r.GetBool("app.debug") {
		t.Error("GetBool failed")
	}
	if r.GetBool("missing", false) {
		t.Error("GetBool default failed")
	}
}

func TestHas(t *testing.T) {
	r := New()
	r.Set("app.name", "Ignite")
	if !r.Has("app.name") {
		t.Error("Has should return true")
	}
	if r.Has("missing") {
		t.Error("Has should return false")
	}
}

func TestNestedDotNotation(t *testing.T) {
	r := New()
	r.Set("database.mysql.host", "localhost")
	r.Set("database.mysql.port", 3306)
	if r.GetString("database.mysql.host") != "localhost" {
		t.Error("nested dot notation failed")
	}
	if r.GetInt("database.mysql.port") != 3306 {
		t.Error("nested int failed")
	}
}

func TestLoadEnv(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	content := `APP_NAME=Ignite
APP_PORT=8080
APP_DEBUG=true
# comment line
APP_KEY="base64:abc123"
`
	os.WriteFile(envFile, []byte(content), 0644)

	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_PORT")
	os.Unsetenv("APP_DEBUG")
	os.Unsetenv("APP_KEY")

	err := LoadEnv(envFile)
	if err != nil {
		t.Fatal(err)
	}

	if os.Getenv("APP_NAME") != "Ignite" {
		t.Error("APP_NAME not loaded")
	}
	if os.Getenv("APP_PORT") != "8080" {
		t.Error("APP_PORT not loaded")
	}
	if os.Getenv("APP_KEY") != "base64:abc123" {
		t.Error("APP_KEY not loaded, got:", os.Getenv("APP_KEY"))
	}
}

func TestEnvDoesNotOverride(t *testing.T) {
	os.Setenv("TEST_VAR", "original")
	defer os.Unsetenv("TEST_VAR")

	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("TEST_VAR=overridden"), 0644)

	LoadEnv(envFile)
	if os.Getenv("TEST_VAR") != "original" {
		t.Error("env should not override existing vars")
	}
}

func TestEnvHelpers(t *testing.T) {
	os.Setenv("MY_STR", "hello")
	os.Setenv("MY_INT", "42")
	os.Setenv("MY_BOOL", "true")
	defer func() {
		os.Unsetenv("MY_STR")
		os.Unsetenv("MY_INT")
		os.Unsetenv("MY_BOOL")
	}()

	if Env("MY_STR") != "hello" {
		t.Error("Env failed")
	}
	if Env("MISSING", "default") != "default" {
		t.Error("Env default failed")
	}
	if EnvInt("MY_INT") != 42 {
		t.Error("EnvInt failed")
	}
	if !EnvBool("MY_BOOL") {
		t.Error("EnvBool failed")
	}
}

func TestAll(t *testing.T) {
	r := New()
	r.Set("a", 1)
	r.Set("b", 2)
	all := r.All()
	if len(all) != 2 {
		t.Errorf("expected 2 items, got %d", len(all))
	}
}
