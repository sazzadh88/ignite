package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// NewProject scaffolds a new Ignite application.
func NewProject(name string) error {
	if _, err := os.Stat(name); err == nil {
		return fmt.Errorf("directory '%s' already exists", name)
	}

	fmt.Printf("  Creating application '%s'...\n", name)

	dirs := []string{
		"app/Console",
		"app/Exceptions",
		"app/Http/Controllers",
		"app/Http/Middleware",
		"app/Http/Requests",
		"app/Models",
		"app/Policies",
		"app/Providers",
		"app/Jobs",
		"app/appauth",
		"app/appcsrf",
		"app/projdb",
		"app/repositories",
		"bootstrap",
		"config",
		"database/migrations",
		"database/seeders",
		"database/factories",
		"public",
		"resources/views",
		"resources/views/layouts",
		"resources/views/auth",
		"resources/lang",
		"routes",
		"storage/app",
		"storage/framework/cache",
		"storage/framework/sessions",
		"storage/framework/views",
		"storage/logs",
		"tests/Feature",
		"tests/Unit",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join(name, dir), 0755); err != nil {
			return err
		}
	}

	modulePath := "github.com/user/" + name
	replacements := map[string]string{
		"__MOD__":          modulePath,
		"__APP_NAME__":     name,
		"__IGNITE_PATH__":  ignitePath(),
	}

	// The appcsrf stub has `__ERROR_PAGE_TPL__` inside backticks; replace with raw HTML.
	errorPageHTML := stub("project/views/error-page.html.stub", nil)
	csrfReplacements := map[string]string{
		"__MOD__":            modulePath,
		"__ERROR_PAGE_TPL__": errorPageHTML,
	}

	files := map[string]string{
		"go.mod":                       stub("project/go.mod.stub", replacements),
		"main.go":                      stub("project/main.go.stub", replacements),
		"ignite":                       stub("project/ignite.stub", nil),
		"ignite.bat":                   stub("project/ignite.bat.stub", nil),
		".env":                         stub("project/env.stub", replacements),
		".env.example":                 stub("project/env.stub", replacements),
		".gitignore":                   stub("project/gitignore.stub", nil),
		"bootstrap/app.go":             stub("project/bootstrap-app.go.stub", replacements),
		"database/migrations/migrations.go": stub("project/migrations-package.go.stub", nil),
		"config/config.go":             stub("project/config-loader.go.stub", nil),
		"config/app.go":                stub("project/config-app.go.stub", nil),
		"config/database.go":           stub("project/config-database.go.stub", nil),
		"routes/web.go":                stub("project/routes-web.go.stub", replacements),
		"routes/api.go":                stub("project/routes-api.go.stub", nil),
		"database/migrations/0001_01_01_000000_create_users_table.go": stub("project/migration-users.go.stub", nil),
		"app/projdb/projdb.go":         stub("project/projdb.go.stub", nil),
		"app/repositories/user_repo.go": stub("project/user-repo.go.stub", replacements),
		"app/appauth/appauth.go":       stub("project/appauth.go.stub", replacements),
		"app/appcsrf/appcsrf.go":       stub("project/appcsrf.go.stub", csrfReplacements),
		"app/Http/Controllers/AuthController.go": stub("project/auth-controller.go.stub", replacements),
		"resources/views/auth/register.ignite.html":        stub("project/views/register.ignite.html.stub", nil),
		"resources/views/auth/login.ignite.html":           stub("project/views/login.ignite.html.stub", nil),
		"resources/views/auth/profile.ignite.html":         stub("project/views/profile.ignite.html.stub", nil),
		"resources/views/auth/change-password.ignite.html": stub("project/views/change-password.ignite.html.stub", nil),
		"resources/views/dashboard.ignite.html":            stub("project/views/dashboard.ignite.html.stub", nil),
		"app/Http/Controllers/Controller.go": stub("project/controller-base.go.stub", nil),
		"app/Http/Middleware/Authenticate.go": stub("project/middleware-authenticate.go.stub", replacements),
		"app/Models/User.go":           stub("project/model-user.go.stub", nil),
		"app/Providers/AppServiceProvider.go": stub("project/provider-app.go.stub", replacements),
		"app/Providers/RouteServiceProvider.go": stub("project/provider-route.go.stub", replacements),
		"app/Console/Kernel.go":        stub("project/console-kernel.go.stub", replacements),
		"app/Exceptions/Handler.go":    stub("project/exception-handler.go.stub", nil),
		"database/seeders/DatabaseSeeder.go": stub("project/database-seeder.go.stub", nil),
		"resources/views/layouts/app.ignite.html": stub("project/views/layout.ignite.html.stub", nil),
		"resources/views/welcome.ignite.html": stub("project/views/welcome.ignite.html.stub", replacements),
		"public/index.go":              stub("project/public-index.go.stub", nil),
		"tests/Feature/.gitkeep":       "",
		"tests/Unit/.gitkeep":          "",
	}

	for path, content := range files {
		fullPath := filepath.Join(name, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	if err := os.Chmod(filepath.Join(name, "ignite"), 0755); err != nil {
		return err
	}

	storageGitignore := "*\n!.gitignore\n"
	for _, dir := range []string{"storage/app", "storage/framework/cache", "storage/framework/sessions", "storage/framework/views", "storage/logs"} {
		os.WriteFile(filepath.Join(name, dir, ".gitignore"), []byte(storageGitignore), 0644)
	}

	return nil
}

func ignitePath() string {
	home, _ := os.UserHomeDir()
	candidates := []string{
		filepath.Join(home, "Projects/Go/framework"),
		filepath.Join(home, "go/src/github.com/sazzadh88/ignite"),
	}
	for _, p := range candidates {
		if _, err := os.Stat(filepath.Join(p, "go.mod")); err == nil {
			return p
		}
	}
	return "../framework"
}

// dashboardLayoutWrap renders a page with the full dashboard layout.
// Used by the view stubs that need sidebar + topbar chrome.
func dashboardLayoutWrap(title, activePage, content string) string {
	type navItem struct{ href, label, id string }
	items := []navItem{
		{"/dashboard", "Dashboard", "dashboard"},
		{"/profile", "Profile", "profile"},
		{"/change-password", "Change Password", "change-password"},
	}
	var sidebarItems string
	for _, item := range items {
		activeClass := ""
		if item.id == activePage {
			activeClass = " active"
		}
		sidebarItems += fmt.Sprintf(`      <a href="%s" class="nav-item%s">%s</a>`+"\n", item.href, activeClass, item.label)
	}

	return strings.ReplaceAll(
		strings.ReplaceAll(
			strings.ReplaceAll(
				dashboardLayoutShell,
				"__TITLE__", title),
			"__SIDEBAR_ITEMS__", sidebarItems),
		"__CONTENT__", content)
}

// dashboardLayoutShell is loaded once from the embedded stub if needed.
// For now the stubs contain the fully-assembled views, so this is only
// used as a reference for future dynamic assembly.
const dashboardLayoutShell = "" // unused — views are pre-assembled in stubs
