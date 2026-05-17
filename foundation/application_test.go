package foundation

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewApplication(t *testing.T) {
	app := NewApplication("/tmp/testapp")
	if app == nil {
		t.Fatal("app should not be nil")
	}
	if app.BasePath() != "/tmp/testapp" {
		t.Error("base path wrong")
	}
}

func TestVersion(t *testing.T) {
	app := NewApplication("/tmp/testapp")
	if app.Version() != Version {
		t.Error("version mismatch")
	}
}

func TestPaths(t *testing.T) {
	app := NewApplication("/tmp/testapp")

	if app.StoragePath() != "/tmp/testapp/storage" {
		t.Errorf("storage path: %s", app.StoragePath())
	}
	if app.StoragePath("logs") != "/tmp/testapp/storage/logs" {
		t.Errorf("storage subpath: %s", app.StoragePath("logs"))
	}
	if app.ConfigPath() != "/tmp/testapp/config" {
		t.Errorf("config path: %s", app.ConfigPath())
	}
	if app.DatabasePath("migrations") != "/tmp/testapp/database/migrations" {
		t.Errorf("database path: %s", app.DatabasePath("migrations"))
	}
}

func TestEnvironment(t *testing.T) {
	app := NewApplication("/tmp/testapp")
	os.Setenv("APP_ENV", "production")
	defer os.Unsetenv("APP_ENV")
	app.Bootstrap()
	if !app.IsProduction() {
		t.Error("should be production")
	}
}

func TestBootstrap(t *testing.T) {
	dir := t.TempDir()
	envFile := filepath.Join(dir, ".env")
	os.WriteFile(envFile, []byte("APP_NAME=TestApp\nAPP_ENV=testing\nAPP_DEBUG=true\n"), 0644)

	os.Unsetenv("APP_NAME")
	os.Unsetenv("APP_ENV")
	os.Unsetenv("APP_DEBUG")

	app := NewApplication(dir)
	app.Bootstrap()

	if app.Config().GetString("app.name") != "TestApp" {
		t.Error("app name not loaded")
	}
	if !app.IsTesting() {
		t.Error("should be testing env")
	}
}

type testProvider struct {
	registered bool
	booted     bool
}

func (p *testProvider) Register(app *Application) { p.registered = true }
func (p *testProvider) Boot(app *Application)     { p.booted = true }

func TestServiceProvider(t *testing.T) {
	app := NewApplication("/tmp/testapp")
	p := &testProvider{}
	app.Register(p)

	if !p.registered {
		t.Error("provider not registered")
	}
	if p.booted {
		t.Error("provider should not boot before app.Boot()")
	}

	app.Boot()
	if !p.booted {
		t.Error("provider not booted")
	}
}

func TestBootCalledOnceOnly(t *testing.T) {
	app := NewApplication("/tmp/testapp")
	app.Boot()
	app.Boot()
	if !app.IsBooted() {
		t.Error("should be booted")
	}
}

func TestLateProviderBootedImmediately(t *testing.T) {
	app := NewApplication("/tmp/testapp")
	app.Boot()

	p := &testProvider{}
	app.Register(p)
	if !p.booted {
		t.Error("late provider should boot immediately after app is booted")
	}
}
