package scaffold

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewProject(t *testing.T) {
	dir := t.TempDir()
	appName := "testapp"
	appPath := filepath.Join(dir, appName)

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := NewProject(appName)
	if err != nil {
		t.Fatalf("NewProject failed: %v", err)
	}

	// Check directories exist
	expectedDirs := []string{
		"app/Http/Controllers",
		"app/Http/Middleware",
		"app/Models",
		"app/Providers",
		"bootstrap",
		"config",
		"database/migrations",
		"database/seeders",
		"routes",
		"resources/views",
		"storage/logs",
		"tests/Feature",
		"tests/Unit",
	}
	for _, d := range expectedDirs {
		if _, err := os.Stat(filepath.Join(appPath, d)); os.IsNotExist(err) {
			t.Errorf("directory missing: %s", d)
		}
	}

	// Check key files exist
	expectedFiles := []string{
		"go.mod",
		"main.go",
		".env",
		".env.example",
		".gitignore",
		"bootstrap/app.go",
		"config/app.go",
		"routes/web.go",
		"routes/api.go",
		"app/Models/User.go",
		"app/Providers/AppServiceProvider.go",
		"resources/views/welcome.ignite.html",
	}
	for _, f := range expectedFiles {
		if _, err := os.Stat(filepath.Join(appPath, f)); os.IsNotExist(err) {
			t.Errorf("file missing: %s", f)
		}
	}
}

func TestNewProjectAlreadyExists(t *testing.T) {
	dir := t.TempDir()
	appName := "existingapp"

	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	os.MkdirAll(appName, 0755)

	err := NewProject(appName)
	if err == nil {
		t.Error("should error when directory exists")
	}
}

func TestMakeController(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeController("UserController")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Controllers/UserController.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("controller not created")
	}
}

func TestMakeModel(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeModel("Post")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Models/Post.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("model not created")
	}
}

func TestMakeMiddleware(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeMiddleware("RateLimit")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Middleware/RateLimit.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("middleware not created")
	}
}

func TestMakeMigration(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeMigration("create_posts_table")
	if err != nil {
		t.Fatal(err)
	}

	files, _ := os.ReadDir("database/migrations")
	if len(files) == 0 {
		t.Error("migration not created")
	}
}

func TestMakeSeeder(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeSeeder("UserSeeder")
	if err != nil {
		t.Fatal(err)
	}

	path := "database/seeders/UserSeeder.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("seeder not created")
	}
}
