package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
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

	files := map[string]string{
		"go.mod":                       goModTemplate(modulePath),
		"main.go":                      mainTemplate(modulePath),
		"ignite":                       binIgniteTemplate(),
		"ignite.bat":                   binIgniteBatTemplate(),
		".env":                         envTemplate(name),
		".env.example":                 envTemplate(name),
		".gitignore":                   gitignoreTemplate(),
		"bootstrap/app.go":             bootstrapTemplate(modulePath),
		"database/migrations/migrations.go": migrationsPackageTemplate(),
		"config/config.go":             configLoaderTemplate(),
		"config/app.go":                configAppTemplate(),
		"config/database.go":           configDatabaseTemplate(),
		"routes/web.go":                routesWebTemplate(modulePath),
		"routes/api.go":                routesAPITemplate(),
		"database/migrations/0001_01_01_000000_create_users_table.go": usersMigrationTemplate(),
		"app/projdb/projdb.go":         projdbTemplate(modulePath),
		"app/repositories/user_repo.go": userRepoTemplate(modulePath),
		"app/appauth/appauth.go":       appAuthTemplate(modulePath),
		"app/appcsrf/appcsrf.go":       appCsrfTemplate(),
		"app/Http/Controllers/AuthController.go": authControllerTemplate(modulePath),
		"resources/views/auth/register.ignite.html": registerViewTemplate(),
		"resources/views/auth/login.ignite.html":    loginViewTemplate(),
		"resources/views/dashboard.ignite.html":     dashboardViewTemplate(),
		"app/Http/Controllers/Controller.go": baseControllerTemplate(modulePath),
		"app/Http/Middleware/Authenticate.go": authenticateMiddlewareTemplate(modulePath),
		"app/Models/User.go":           userModelTemplate(modulePath),
		"app/Providers/AppServiceProvider.go": appServiceProviderTemplate(modulePath),
		"app/Providers/RouteServiceProvider.go": routeServiceProviderTemplate(modulePath),
		"app/Console/Kernel.go":        consoleKernelTemplate(modulePath),
		"app/Exceptions/Handler.go":    exceptionHandlerTemplate(modulePath),
		"database/seeders/DatabaseSeeder.go": databaseSeederTemplate(modulePath),
		"resources/views/layouts/app.ignite.html": layoutTemplate(name),
		"resources/views/welcome.ignite.html": welcomeViewTemplate(name),
		"public/index.go":              publicIndexTemplate(modulePath),
		"tests/Feature/.gitkeep":       "",
		"tests/Unit/.gitkeep":          "",
	}

	for path, content := range files {
		fullPath := filepath.Join(name, path)
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			return err
		}
	}

	// The project console wrapper must be executable.
	if err := os.Chmod(filepath.Join(name, "ignite"), 0755); err != nil {
		return err
	}

	// Create storage .gitignore files
	storageGitignore := "*\n!.gitignore\n"
	for _, dir := range []string{"storage/app", "storage/framework/cache", "storage/framework/sessions", "storage/framework/views", "storage/logs"} {
		os.WriteFile(filepath.Join(name, dir, ".gitignore"), []byte(storageGitignore), 0644)
	}

	return nil
}

func goModTemplate(modulePath string) string {
	return fmt.Sprintf(`module %s

go 1.22

require github.com/sazzadh88/ignite v0.1.0

replace github.com/sazzadh88/ignite => %s
`, modulePath, ignitePath())
}

func ignitePath() string {
	// Try to find the ignite source directory
	// Check common locations
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
	// Fallback: use GOPATH-based path or relative
	return "../framework"
}

// migrationsPackageTemplate is the base file for the database/migrations
// package so it compiles (and can be blank-imported) before any migrations
// exist. Generated migrations register themselves via init().
func migrationsPackageTemplate() string {
	return `// Package migrations holds the application's database migrations.
// Each migration file registers itself via schema.RegisterMigration in an
// init() function; main.go blank-imports this package so they load.
package migrations
`
}

func mainTemplate(modulePath string) string {
	return fmt.Sprintf(`package main

import (
	"log"

	"github.com/sazzadh88/ignite/foundation"
	"github.com/sazzadh88/ignite/routing"
	"%s/app/appcsrf"
	"%s/app/projdb"
	"%s/config"
	_ "%s/database/migrations"
	"%s/routes"
)

func main() {
	app := foundation.NewApplication(".")
	app.Bootstrap()

	// Apply project configuration on top of framework defaults.
	config.Load(app.Config())

	// Open the default database connection for repositories.
	if err := projdb.Init(app.Config()); err != nil {
		log.Println("database:", err)
	}

	// Register routes
	routes.RegisterWeb()
	routes.RegisterAPI()

	// Run the application (CSRF-protected router).
	app.Run(appcsrf.Protect(routing.DefaultRouter))
}
`, modulePath, modulePath, modulePath, modulePath, modulePath)
}

// binIgniteTemplate is the project console wrapper, shipped at the project
// root (the analog of Laravel's ./artisan). It runs the project's own main.go
// dispatch, so commands use the framework version pinned in this project's
// go.mod — no global install required.
func binIgniteTemplate() string {
	return `#!/usr/bin/env sh
# Ignite project console.
#
# Runs project commands (serve, key:generate, migrate, make:*, ...) using the
# framework version pinned in this project's go.mod. The global "ignite" binary
# is only for creating new projects.
#
# Usage: ./ignite <command> [arguments] [flags]

cd "$(dirname "$0")" || exit 1
exec go run . "$@"
`
}

func binIgniteBatTemplate() string {
	return "@echo off\r\n" +
		"REM Ignite project console. Runs project commands using the framework\r\n" +
		"REM version pinned in this project's go.mod.\r\n" +
		"cd /d \"%~dp0\"\r\n" +
		"go run . %*\r\n"
}

func envTemplate(name string) string {
	return fmt.Sprintf(`APP_NAME=%s
APP_ENV=local
APP_KEY=
APP_DEBUG=true
APP_URL=http://localhost
APP_PORT=8080

DB_CONNECTION=sqlite
# Uncomment and set these to use mysql or pgsql instead of sqlite.
# DB_HOST=127.0.0.1
# DB_PORT=3306
# DB_DATABASE=%s
# DB_USERNAME=root
# DB_PASSWORD=
# DB_SOCKET=    # socket-only MySQL, e.g. /tmp/mysql.sock

CACHE_DRIVER=file
SESSION_DRIVER=file
QUEUE_CONNECTION=sync

MAIL_MAILER=smtp
MAIL_HOST=127.0.0.1
MAIL_PORT=1025
MAIL_USERNAME=null
MAIL_PASSWORD=null
MAIL_ENCRYPTION=null
MAIL_FROM_ADDRESS="hello@example.com"
MAIL_FROM_NAME="${APP_NAME}"
`, name, name)
}

func gitignoreTemplate() string {
	return `/vendor
.env
storage/logs/*.log
storage/framework/cache/*
storage/framework/sessions/*
storage/framework/views/*
*.exe
*.test
*.out
`
}

func bootstrapTemplate(modulePath string) string {
	_ = modulePath
	return `package bootstrap

import "github.com/sazzadh88/ignite/foundation"

// CreateApplication bootstraps and returns the application.
func CreateApplication() *foundation.Application {
	app := foundation.NewApplication(".")
	app.Bootstrap()
	return app
}
`
}

// configLoaderTemplate is the project's config entrypoint. The framework
// already maps .env into sensible defaults at bootstrap; Load() runs after
// that so a project can override or extend any configuration value. This is
// the analog of Laravel's config/*.php files.
func configLoaderTemplate() string {
	return `package config

import icfg "github.com/sazzadh88/ignite/config"

// Load applies this project's configuration on top of the framework
// defaults. It is called from main.go after app.Bootstrap().
func Load(c *icfg.Repository) {
	registerApp(c)
	registerDatabase(c)
}
`
}

func configAppTemplate() string {
	return `package config

import icfg "github.com/sazzadh88/ignite/config"

// registerApp defines application configuration. Values come from .env;
// the second argument to icfg.Env is the default. This block is the whole
// "app" config section (the analog of Laravel's config/app.php).
func registerApp(c *icfg.Repository) {
	c.Set("app", map[string]any{
		"name":     icfg.Env("APP_NAME", "Ignite"),
		"env":      icfg.Env("APP_ENV", "local"),
		"debug":    icfg.EnvBool("APP_DEBUG", true),
		"url":      icfg.Env("APP_URL", "http://localhost"),
		"port":     icfg.EnvInt("APP_PORT", 8080),
		"key":      icfg.Env("APP_KEY", ""),
		"timezone": icfg.Env("APP_TIMEZONE", "UTC"),
	})
}
`
}

func configDatabaseTemplate() string {
	return `package config

import icfg "github.com/sazzadh88/ignite/config"

// registerDatabase defines database connections (the analog of Laravel's
// config/database.php). "default" selects which connection under
// "connections" is used; switch it via DB_CONNECTION in .env.
func registerDatabase(c *icfg.Repository) {
	c.Set("database", map[string]any{
		"default": icfg.Env("DB_CONNECTION", "sqlite"),
		"connections": map[string]any{
			"sqlite": map[string]any{
				"driver":   "sqlite3",
				"database": icfg.Env("DB_DATABASE", "database/database.sqlite"),
			},
			"mysql": map[string]any{
				"driver":         "mysql",
				"host":           icfg.Env("DB_HOST", "127.0.0.1"),
				"port":           icfg.EnvInt("DB_PORT", 3306),
				"database":       icfg.Env("DB_DATABASE", "ignite"),
				"username":       icfg.Env("DB_USERNAME", "root"),
				"password":       icfg.Env("DB_PASSWORD", ""),
				"unix_socket":    icfg.Env("DB_SOCKET", ""),
				"charset":        icfg.Env("DB_CHARSET", "utf8mb4"),
				"prefix":         "",
				"prefix_indexes": true,
				"strict":         icfg.EnvBool("DB_STRICT", true),
				"engine":         icfg.Env("DB_ENGINE", ""),
				// TLS over TCP: prefer = encrypt if supported, else plaintext.
				"sslmode": icfg.Env("DB_SSLMODE", "prefer"),
			},
			"pgsql": map[string]any{
				"driver":         "postgres",
				"host":           icfg.Env("DB_HOST", "127.0.0.1"),
				"port":           icfg.EnvInt("DB_PORT", 5432),
				"database":       icfg.Env("DB_DATABASE", "ignite"),
				"username":       icfg.Env("DB_USERNAME", "root"),
				"password":       icfg.Env("DB_PASSWORD", ""),
				"charset":        icfg.Env("DB_CHARSET", "utf8"),
				"prefix":         "",
				"prefix_indexes": true,
				"search_path":    icfg.Env("DB_SEARCH_PATH", "public"),
				// "prefer": encrypted if the server supports SSL, else
				// plaintext. Set DB_SSLMODE=require to force encryption.
				"sslmode": icfg.Env("DB_SSLMODE", "prefer"),
			},
		},
	})
}
`
}

func routesWebTemplate(modulePath string) string {
	return modReplace(modulePath, `package routes

import (
	"net/http"
	"runtime"

	"github.com/sazzadh88/ignite/foundation"
	"github.com/sazzadh88/ignite/routing"
	"github.com/sazzadh88/ignite/view"

	"__MOD__/app/Http/Controllers"
	"__MOD__/app/appauth"
)

// RegisterWeb defines web routes. A working authentication flow ships
// out of the box: register, login, logout, and a protected dashboard.
func RegisterWeb() {
	r := routing.DefaultRouter
	ac := &Controllers.AuthController{}

	r.Get("/", routing.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		engine := view.NewEngine("resources/views")
		html, err := engine.Render("welcome", map[string]any{
			"frameworkVersion": foundation.Version,
			"goVersion":        runtime.Version(),
		})
		if err != nil {
			http.Error(w, "View error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(html))
	})).Name("home")

	r.Get("/register", routing.HandlerFunc(ac.ShowRegister)).Name("register")
	r.Post("/register", routing.HandlerFunc(ac.Register))
	r.Get("/login", routing.HandlerFunc(ac.ShowLogin)).Name("login")
	r.Post("/login", routing.HandlerFunc(ac.Login))
	r.Post("/logout", routing.HandlerFunc(ac.Logout)).Name("logout")
	r.Get("/dashboard", routing.HandlerFunc(appauth.RequireAuth(ac.Dashboard))).Name("dashboard")
}
`)
}

func routesAPITemplate() string {
	return `package routes

// API routes are loaded by the RouteServiceProvider.
// These routes are assigned the "api" middleware group with /api prefix.

func RegisterAPI() {
	// Route.Get("/users", controllers.UserController.Index)
	// Route.Post("/users", controllers.UserController.Store)
}
`
}

func baseControllerTemplate(modulePath string) string {
	_ = modulePath
	return `package Controllers

// Controller is the base controller that other controllers embed.
type Controller struct{}
`
}

func authenticateMiddlewareTemplate(modulePath string) string {
	_ = modulePath
	return `package Middleware

import "net/http"

// Authenticate verifies the user is authenticated.
type Authenticate struct{}

// Handle processes the request.
func (a *Authenticate) Handle(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// TODO: Check authentication
	next.ServeHTTP(w, r)
}
`
}

func userModelTemplate(modulePath string) string {
	_ = modulePath
	return `package Models

// User represents the application user. It implements auth.Authenticatable
// (GetAuthIdentifier / GetAuthPassword) so it works with the framework's
// auth guards and user providers.
type User struct {
	ID        int64
	Name      string
	Email     string
	Password  string
	CreatedAt string
	UpdatedAt string
}

// GetAuthIdentifier returns the unique identifier (primary key).
func (u *User) GetAuthIdentifier() any { return u.ID }

// GetAuthPassword returns the hashed password.
func (u *User) GetAuthPassword() string { return u.Password }

// Table returns the table name.
func (u *User) Table() string { return "users" }

// Fillable returns mass-assignable fields.
func (u *User) Fillable() []string { return []string{"name", "email", "password"} }

// Hidden returns fields hidden from serialization.
func (u *User) Hidden() []string { return []string{"password"} }
`
}

// modReplace substitutes the module import path into a template body. Used
// instead of fmt.Sprintf so bodies can contain literal %v/%s (e.g. in
// fmt.Sprintf calls within generated code) without escaping.
func modReplace(modulePath, body string) string {
	return strings.ReplaceAll(body, "__MOD__", modulePath)
}

func usersMigrationTemplate() string {
	return `package migrations

import "github.com/sazzadh88/ignite/schema"

func init() {
	schema.RegisterMigration("0001_01_01_000000_create_users_table", &CreateUsersTable{})
}

// CreateUsersTable creates the users table.
type CreateUsersTable struct{}

// Up runs the migration.
func (m *CreateUsersTable) Up(s *schema.Schema) error {
	return s.Create("users", func(t *schema.Blueprint) {
		t.ID()
		t.String("name")
		t.String("email").Unique()
		t.String("password")
		t.Timestamps()
	})
}

// Down reverses the migration.
func (m *CreateUsersTable) Down(s *schema.Schema) error {
	return s.DropIfExists("users")
}
`
}

func projdbTemplate(modulePath string) string {
	_ = modulePath
	return `// Package projdb exposes the application's default database connection,
// built from config at boot. The framework ORM does not persist yet, so
// repositories use this *database.Connection directly.
package projdb

import (
	"fmt"

	"github.com/sazzadh88/ignite/config"
	"github.com/sazzadh88/ignite/database"
)

var conn *database.Connection

// Init opens the default database connection from configuration.
// Call this from main.go after config.Load().
func Init(c *config.Repository) error {
	dbCfg, ok := c.Get("database").(map[string]any)
	if !ok {
		return fmt.Errorf("database configuration not found")
	}
	mgr := database.NewManager(dbCfg)
	cn, err := mgr.Default()
	if err != nil {
		return fmt.Errorf("database connect: %w", err)
	}
	conn = cn
	return nil
}

// Conn returns the application database connection.
func Conn() *database.Connection { return conn }
`
}

func userRepoTemplate(modulePath string) string {
	return modReplace(modulePath, `// Package repositories holds database access for models. The framework
// ORM does not persist yet, so these use the raw *database.Connection.
package repositories

import (
	"database/sql"
	"fmt"

	"__MOD__/app/Models"
	"__MOD__/app/projdb"
)

// EmailExists reports whether a user with the given email already exists.
func EmailExists(email string) (bool, error) {
	rows, err := projdb.Conn().Query("SELECT id FROM users WHERE email = ? LIMIT 1", email)
	if err != nil {
		return false, err
	}
	return len(rows) > 0, nil
}

// Create inserts a new user and returns the stored record.
func Create(name, email, passwordHash string) (*Models.User, error) {
	_, err := projdb.Conn().Exec(
		"INSERT INTO users (name, email, password, created_at, updated_at) VALUES (?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)",
		name, email, passwordHash,
	)
	if err != nil {
		return nil, err
	}
	return FindByEmail(email)
}

// FindByEmail returns the user with the given email, or sql.ErrNoRows.
func FindByEmail(email string) (*Models.User, error) {
	rows, err := projdb.Conn().Query(
		"SELECT id, name, email, password, created_at, updated_at FROM users WHERE email = ? LIMIT 1",
		email,
	)
	if err != nil {
		return nil, err
	}
	return firstUser(rows)
}

// FindByID returns the user with the given id. id may arrive as int64 or
// float64 (the latter after a JSON round-trip through the session cookie).
func FindByID(id any) (*Models.User, error) {
	rows, err := projdb.Conn().Query(
		"SELECT id, name, email, password, created_at, updated_at FROM users WHERE id = ? LIMIT 1",
		toInt64(id),
	)
	if err != nil {
		return nil, err
	}
	return firstUser(rows)
}

func firstUser(rows []map[string]any) (*Models.User, error) {
	if len(rows) == 0 {
		return nil, sql.ErrNoRows
	}
	r := rows[0]
	return &Models.User{
		ID:        toInt64(r["id"]),
		Name:      str(r["name"]),
		Email:     str(r["email"]),
		Password:  str(r["password"]),
		CreatedAt: str(r["created_at"]),
		UpdatedAt: str(r["updated_at"]),
	}, nil
}

func str(v any) string {
	switch s := v.(type) {
	case nil:
		return ""
	case string:
		return s
	case []byte:
		return string(s)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int64:
		return n
	case int:
		return int64(n)
	case float64:
		return int64(n)
	case string:
		var i int64
		fmt.Sscan(n, &i)
		return i
	}
	return 0
}
`)
}

func appAuthTemplate(modulePath string) string {
	return modReplace(modulePath, `// Package appauth wires the framework auth package to HTTP using an
// encrypted session cookie. The framework does not HTTP-wire auth itself,
// and its router does not run middleware, so route protection is a handler
// wrapper.
package appauth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sazzadh88/ignite/auth"
	"github.com/sazzadh88/ignite/encryption"
	"github.com/sazzadh88/ignite/hashing"

	"__MOD__/app/Models"
	"__MOD__/app/repositories"
)

const cookieName = "ignite_session"

// cookieSession is an auth.Session backed by an encrypted cookie.
type cookieSession struct {
	w      http.ResponseWriter
	r      *http.Request
	data   map[string]any
	loaded bool
}

func (s *cookieSession) load() {
	if s.loaded {
		return
	}
	s.loaded = true
	s.data = map[string]any{}
	if s.r == nil || encryption.Crypt == nil {
		return
	}
	c, err := s.r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return
	}
	plain, err := encryption.Crypt.DecryptString(c.Value)
	if err != nil {
		return
	}
	_ = json.Unmarshal([]byte(plain), &s.data)
}

func (s *cookieSession) save() {
	if s.w == nil || encryption.Crypt == nil {
		return
	}
	b, _ := json.Marshal(s.data)
	enc, err := encryption.Crypt.EncryptString(string(b))
	if err != nil {
		return
	}
	http.SetCookie(s.w, &http.Cookie{
		Name:     cookieName,
		Value:    enc,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

func (s *cookieSession) Get(key string) (any, bool) {
	s.load()
	v, ok := s.data[key]
	return v, ok
}

func (s *cookieSession) Put(key, value any) {
	s.load()
	s.data[fmt.Sprint(key)] = value
	s.save()
}

func (s *cookieSession) Forget(key string) {
	s.load()
	delete(s.data, key)
	s.save()
}

func (s *cookieSession) Flush() {
	s.load()
	s.data = map[string]any{}
	s.save()
}

func provider() auth.UserProvider {
	return auth.NewCallbackProvider(
		func(id any) (auth.Authenticatable, error) {
			u, err := repositories.FindByID(id)
			if err != nil {
				return nil, err
			}
			return u, nil
		},
		func(creds map[string]string) (auth.Authenticatable, error) {
			u, err := repositories.FindByEmail(creds["email"])
			if err != nil {
				return nil, err
			}
			return u, nil
		},
		func(user auth.Authenticatable, creds map[string]string) bool {
			return hashing.Hash.Check(creds["password"], user.GetAuthPassword())
		},
	)
}

// Guard builds a request-scoped session guard.
func Guard(w http.ResponseWriter, r *http.Request) *auth.SessionGuard {
	return auth.NewSessionGuard(provider(), &cookieSession{w: w, r: r})
}

// Attempt validates credentials and, on success, logs the user in.
func Attempt(w http.ResponseWriter, r *http.Request, email, password string) bool {
	return Guard(w, r).Attempt(map[string]string{"email": email, "password": password})
}

// LoginUser logs a user in directly (used after registration).
func LoginUser(w http.ResponseWriter, r *http.Request, u *Models.User) {
	Guard(w, r).Login(u)
}

// Logout clears the session cookie.
func Logout(w http.ResponseWriter, r *http.Request) {
	Guard(w, r).Logout()
}

// Current returns the authenticated user, or nil for a guest.
func Current(w http.ResponseWriter, r *http.Request) *Models.User {
	au := Guard(w, r).User()
	if au == nil {
		return nil
	}
	if u, ok := au.(*Models.User); ok {
		return u
	}
	return nil
}

// RequireAuth wraps a handler, redirecting guests to /login.
func RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !Guard(w, r).Check() {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		next(w, r)
	}
}
`)
}

func appCsrfTemplate() string {
	return `// Package appcsrf adds Laravel-style CSRF protection. The framework
// router does not run middleware, so this wraps the router: it issues a
// per-session token in an encrypted cookie, exposes it to views as
// .CSRFToken (the @csrf blade directive), and rejects unsafe requests
// whose submitted token is missing or wrong.
package appcsrf

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"

	"github.com/sazzadh88/ignite/encryption"
	"github.com/sazzadh88/ignite/routing"
)

const cookieName = "ignite_csrf"

type ctxKey struct{}

// Guard wraps the router with CSRF protection. It embeds *routing.Router so
// it still satisfies the framework RouterInterface (PrintRoutes/RouteList).
type Guard struct {
	*routing.Router
}

// Protect wraps the application router with CSRF protection.
func Protect(r *routing.Router) *Guard { return &Guard{r} }

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func decode(v string) string {
	if encryption.Crypt == nil {
		return v
	}
	s, err := encryption.Crypt.DecryptString(v)
	if err != nil {
		return ""
	}
	return s
}

func encode(t string) string {
	if encryption.Crypt == nil {
		return t
	}
	s, err := encryption.Crypt.EncryptString(t)
	if err != nil {
		return t
	}
	return s
}

// ServeHTTP issues/validates the CSRF token, then delegates to the router.
func (g *Guard) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	token := ""
	if c, err := r.Cookie(cookieName); err == nil && c.Value != "" {
		token = decode(c.Value)
	}
	if token == "" {
		token = newToken()
		http.SetCookie(w, &http.Cookie{
			Name:     cookieName,
			Value:    encode(token),
			Path:     "/",
			HttpOnly: true,
			SameSite: http.SameSiteLaxMode,
		})
	}

	switch r.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		sent := r.Header.Get("X-CSRF-Token")
		if sent == "" {
			r.ParseForm()
			sent = r.PostFormValue("_token")
		}
		if sent == "" || sent != token {
			w.WriteHeader(419)
			w.Write([]byte("419 Page Expired - CSRF token mismatch."))
			return
		}
	}

	r = r.WithContext(context.WithValue(r.Context(), ctxKey{}, token))
	g.Router.ServeHTTP(w, r)
}

// Token returns the current request's CSRF token. Pass it to views as
// the "CSRFToken" key so the @csrf directive renders correctly.
func Token(r *http.Request) string {
	if v, ok := r.Context().Value(ctxKey{}).(string); ok {
		return v
	}
	return ""
}
`
}

func authControllerTemplate(modulePath string) string {
	return modReplace(modulePath, `package Controllers

import (
	"net/http"
	"strings"

	"github.com/sazzadh88/ignite/hashing"
	"github.com/sazzadh88/ignite/view"

	"__MOD__/app/appauth"
	"__MOD__/app/appcsrf"
	"__MOD__/app/repositories"
)

// AuthController handles registration, login and logout.
type AuthController struct{}

func (c *AuthController) render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	data["CSRFToken"] = appcsrf.Token(r)
	engine := view.NewEngine("resources/views")
	html, err := engine.Render(name, data)
	if err != nil {
		http.Error(w, "View error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

// ShowRegister renders the registration form.
func (c *AuthController) ShowRegister(w http.ResponseWriter, r *http.Request) {
	c.render(w, r, "auth.register", map[string]any{"Error": "", "Name": "", "Email": ""})
}

// Register creates a new user and logs them in.
func (c *AuthController) Register(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	fail := func(msg string) {
		c.render(w, r, "auth.register", map[string]any{"Error": msg, "Name": name, "Email": email})
	}

	if name == "" || email == "" || len(password) < 6 {
		fail("All fields are required and password must be at least 6 characters.")
		return
	}
	if exists, _ := repositories.EmailExists(email); exists {
		fail("That email is already registered.")
		return
	}
	hash, err := hashing.Hash.Make(password)
	if err != nil {
		fail("Could not hash password.")
		return
	}
	u, err := repositories.Create(name, email, hash)
	if err != nil {
		fail("Could not create account: " + err.Error())
		return
	}
	appauth.LoginUser(w, r, u)
	http.Redirect(w, r, "/dashboard", http.StatusFound)
}

// ShowLogin renders the login form.
func (c *AuthController) ShowLogin(w http.ResponseWriter, r *http.Request) {
	c.render(w, r, "auth.login", map[string]any{"Error": "", "Email": ""})
}

// Login authenticates a user.
func (c *AuthController) Login(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	email := strings.TrimSpace(r.FormValue("email"))
	password := r.FormValue("password")

	if appauth.Attempt(w, r, email, password) {
		http.Redirect(w, r, "/dashboard", http.StatusFound)
		return
	}
	c.render(w, r, "auth.login", map[string]any{"Error": "Invalid credentials.", "Email": email})
}

// Logout ends the session.
func (c *AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	appauth.Logout(w, r)
	http.Redirect(w, r, "/login", http.StatusFound)
}

// Dashboard is a protected page.
func (c *AuthController) Dashboard(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"Name": "", "Email": ""}
	if u := appauth.Current(w, r); u != nil {
		data["Name"] = u.Name
		data["Email"] = u.Email
	}
	c.render(w, r, "dashboard", data)
}
`)
}

// authLayoutWrap renders an auth page (register/login/dashboard) as a
// glass card on the same gradient backdrop as the welcome screen.
func authLayoutWrap(title, card string) string {
	const styles = `    body {
        background:
            radial-gradient(900px circle at 12% -10%, rgba(99,102,241,0.18), transparent 45%),
            radial-gradient(800px circle at 90% 0%, rgba(236,72,153,0.14), transparent 42%),
            #0b1020;
        overflow-x: hidden;
    }
    .glow { position: fixed; border-radius: 50%; filter: blur(90px);
        opacity: 0.5; z-index: 0; pointer-events: none;
        animation: float 14s ease-in-out infinite; }
    .glow.a { width: 380px; height: 380px; top: -120px; left: -90px;
        background: radial-gradient(circle, #6366f1, transparent 70%); }
    .glow.b { width: 440px; height: 440px; bottom: -160px; right: -120px;
        background: radial-gradient(circle, #ec4899, transparent 70%);
        animation-delay: -7s; }
    @keyframes float { 0%,100% { transform: translateY(0) scale(1); }
        50% { transform: translateY(-24px) scale(1.05); } }
    .auth-wrap { position: relative; z-index: 1; min-height: 100vh;
        display: flex; align-items: center; justify-content: center; padding: 2rem 1.25rem; }
    .auth-card { width: min(420px, 94vw);
        background: rgba(15,23,42,0.72); border: 1px solid rgba(148,163,184,0.16);
        border-radius: 18px; padding: 2.4rem 2.2rem; backdrop-filter: blur(14px);
        box-shadow: 0 30px 70px -28px rgba(0,0,0,0.75);
        animation: rise .5s ease both; }
    @keyframes rise { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: none; } }
    .brand { font-size: .82rem; font-weight: 800; letter-spacing: .22em;
        background: linear-gradient(110deg, #818cf8, #c084fc 50%, #f472b6);
        -webkit-background-clip: text; background-clip: text;
        -webkit-text-fill-color: transparent; margin-bottom: 1.4rem; }
    .auth-card h1 { font-size: 1.55rem; color: #f1f5f9; margin-bottom: .4rem; letter-spacing: -0.02em; }
    .sub { color: #94a3b8; font-size: .92rem; margin-bottom: 1.7rem; line-height: 1.55; }
    .err { background: rgba(239,68,68,0.12); border: 1px solid rgba(239,68,68,0.35);
        color: #fca5a5; padding: .7rem .9rem; border-radius: 10px; font-size: .85rem; margin-bottom: 1.1rem; }
    .field { margin-bottom: 1rem; }
    .field input { width: 100%; padding: .82rem .95rem;
        background: rgba(148,163,184,0.06); border: 1px solid rgba(148,163,184,0.18);
        border-radius: 10px; color: #e2e8f0; font-size: .95rem; outline: none;
        transition: border-color .15s, box-shadow .15s; }
    .field input::placeholder { color: #64748b; }
    .field input:focus { border-color: #6366f1; box-shadow: 0 0 0 3px rgba(99,102,241,0.22); }
    .btn-primary { width: 100%; padding: .85rem; border: 0; border-radius: 10px;
        font-size: .98rem; font-weight: 650; color: #fff; cursor: pointer;
        background: linear-gradient(120deg, #6366f1, #a855f7);
        box-shadow: 0 12px 30px -10px rgba(124,58,237,0.65);
        transition: transform .15s, box-shadow .2s; }
    .btn-primary:hover { transform: translateY(-1px); box-shadow: 0 18px 40px -10px rgba(124,58,237,0.8); }
    .alt { margin-top: 1.4rem; text-align: center; color: #94a3b8; font-size: .88rem; }
    .alt a { color: #818cf8; text-decoration: none; }
    .alt a:hover { color: #a5b4fc; }
    .ok { display: inline-flex; align-items: center; gap: .5rem; margin-bottom: 1rem;
        padding: .35rem .8rem; border-radius: 999px; font-size: .78rem; color: #6ee7b7;
        background: rgba(16,185,129,0.12); border: 1px solid rgba(16,185,129,0.3); }
    .ok .dot { width: 7px; height: 7px; border-radius: 50%; background: #34d399;
        box-shadow: 0 0 10px #34d399; }
    .meta { color: #cbd5e1; font-weight: 600; }`
	return "@extends(\"layouts.app\")\n\n@section(\"title\")\n" + title +
		"\n@endsection\n\n@section(\"styles\")\n" + styles +
		"\n@endsection\n\n@section(\"content\")\n" +
		"<div class=\"glow a\"></div>\n<div class=\"glow b\"></div>\n" +
		"<div class=\"auth-wrap\">\n  <div class=\"auth-card\">\n" + card +
		"\n  </div>\n</div>\n@endsection\n"
}

func registerViewTemplate() string {
	return authLayoutWrap("Register · Ignite", `    <div class="brand">IGNITE</div>
    <h1>Create your account</h1>
    <p class="sub">Start building with your new Ignite application.</p>
    @if(.Error)
    <div class="err">{{ .Error }}</div>
    @endif
    <form method="POST" action="/register">
      @csrf
      <div class="field"><input name="name" placeholder="Full name" value="{{ .Name }}"></div>
      <div class="field"><input name="email" type="email" placeholder="Email address" value="{{ .Email }}"></div>
      <div class="field"><input name="password" type="password" placeholder="Password (min 6 characters)"></div>
      <button class="btn-primary" type="submit">Create account</button>
    </form>
    <p class="alt">Already have an account? <a href="/login">Log in</a></p>`)
}

func loginViewTemplate() string {
	return authLayoutWrap("Log in · Ignite", `    <div class="brand">IGNITE</div>
    <h1>Welcome back</h1>
    <p class="sub">Log in to continue to your dashboard.</p>
    @if(.Error)
    <div class="err">{{ .Error }}</div>
    @endif
    <form method="POST" action="/login">
      @csrf
      <div class="field"><input name="email" type="email" placeholder="Email address" value="{{ .Email }}"></div>
      <div class="field"><input name="password" type="password" placeholder="Password"></div>
      <button class="btn-primary" type="submit">Log in</button>
    </form>
    <p class="alt">No account yet? <a href="/register">Register</a></p>`)
}

func dashboardViewTemplate() string {
	return authLayoutWrap("Dashboard · Ignite", `    <div class="brand">IGNITE</div>
    <div class="ok"><span class="dot"></span> Authenticated</div>
    <h1>Welcome, <span class="meta">{{ .Name }}</span></h1>
    <p class="sub">You are signed in as {{ .Email }}. This page is protected by the auth guard.</p>
    <form method="POST" action="/logout">
      @csrf
      <button class="btn-primary" type="submit">Log out</button>
    </form>
    <p class="alt"><a href="/">&larr; Back to home</a></p>`)
}

func appServiceProviderTemplate(modulePath string) string {
	_ = modulePath
	return `package Providers

import "github.com/sazzadh88/ignite/foundation"

// AppServiceProvider registers application services.
type AppServiceProvider struct{}

// Register is called during registration phase.
func (p *AppServiceProvider) Register(app *foundation.Application) {
	// Register application bindings
}

// Boot is called after all providers are registered.
func (p *AppServiceProvider) Boot(app *foundation.Application) {
	// Boot application services
}
`
}

func routeServiceProviderTemplate(modulePath string) string {
	_ = modulePath
	return `package Providers

import "github.com/sazzadh88/ignite/foundation"

// RouteServiceProvider loads application routes.
type RouteServiceProvider struct{}

// Register is called during registration phase.
func (p *RouteServiceProvider) Register(app *foundation.Application) {}

// Boot loads the routes.
func (p *RouteServiceProvider) Boot(app *foundation.Application) {
	// Load web and API routes
}
`
}

func consoleKernelTemplate(modulePath string) string {
	_ = modulePath
	return `package Console

// Kernel registers console commands and scheduled tasks.
type Kernel struct{}

// Commands returns the list of registered commands.
func (k *Kernel) Commands() []string {
	return []string{}
}

// Schedule defines the application's command schedule.
func (k *Kernel) Schedule() {
	// schedule.Command("inspire").Hourly()
}
`
}

func exceptionHandlerTemplate(modulePath string) string {
	_ = modulePath
	return `package Exceptions

import "net/http"

// Handler is the application exception handler.
type Handler struct{}

// Report reports an exception.
func (h *Handler) Report(err error) {
	// Log the exception
}

// Render renders an exception into an HTTP response.
func (h *Handler) Render(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
`
}

func databaseSeederTemplate(modulePath string) string {
	_ = modulePath
	return `package seeders

// DatabaseSeeder seeds the database with initial data.
type DatabaseSeeder struct{}

// Run executes the seeder.
func (s *DatabaseSeeder) Run() {
	// Call other seeders
	// UserSeeder{}.Run()
}
`
}

func layoutTemplate(name string) string {
	_ = name
	return `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>@yield("title")</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            min-height: 100vh;
            background: #0f172a;
            color: #e2e8f0;
        }
        @yield("styles")
    </style>
</head>
<body>
    @yield("content")
</body>
</html>
`
}

func welcomeViewTemplate(name string) string {
	return strings.ReplaceAll(`@extends("layouts.app")

@section("title")
Ignite — Web artisan meets Go performance
@endsection

@section("styles")
    body {
        background:
            radial-gradient(900px circle at 12% -10%, rgba(99,102,241,0.18), transparent 45%),
            radial-gradient(800px circle at 90% 0%, rgba(236,72,153,0.14), transparent 42%),
            #0b1020;
        overflow-x: hidden;
    }
    .glow {
        position: fixed; border-radius: 50%; filter: blur(90px);
        opacity: 0.55; z-index: 0; pointer-events: none;
        animation: float 14s ease-in-out infinite;
    }
    .glow.a { width: 420px; height: 420px; top: -120px; left: -80px;
        background: radial-gradient(circle, #6366f1, transparent 70%); }
    .glow.b { width: 480px; height: 480px; bottom: -160px; right: -120px;
        background: radial-gradient(circle, #ec4899, transparent 70%);
        animation-delay: -7s; }
    @keyframes float {
        0%,100% { transform: translateY(0) scale(1); }
        50%     { transform: translateY(-26px) scale(1.06); }
    }
    .wrap {
        position: relative; z-index: 1;
        max-width: 1040px; margin: 0 auto;
        padding: 7rem 1.5rem 4rem; text-align: center;
    }
    .badge {
        display: inline-flex; align-items: center; gap: .5rem;
        padding: .4rem .9rem; border-radius: 999px;
        background: rgba(148,163,184,0.08);
        border: 1px solid rgba(148,163,184,0.18);
        color: #cbd5e1; font-size: .8rem; letter-spacing: .02em;
        backdrop-filter: blur(8px);
        animation: rise .6s ease both;
    }
    .badge .dot { width: 7px; height: 7px; border-radius: 50%;
        background: #34d399; box-shadow: 0 0 10px #34d399; }
    .logo {
        font-size: clamp(3rem, 9vw, 6rem); font-weight: 850;
        line-height: 1.05; margin: 1.4rem 0 .4rem; letter-spacing: -0.03em;
        background: linear-gradient(110deg, #818cf8, #c084fc 45%, #f472b6);
        background-size: 220% 220%;
        -webkit-background-clip: text; background-clip: text;
        -webkit-text-fill-color: transparent;
        animation: rise .6s .05s ease both, shimmer 7s ease-in-out infinite;
    }
    @keyframes shimmer {
        0%,100% { background-position: 0% 50%; }
        50%     { background-position: 100% 50%; }
    }
    .tag { font-size: clamp(1.05rem, 2.4vw, 1.4rem); color: #94a3b8;
        max-width: 620px; margin: 0 auto 2.4rem; line-height: 1.6;
        animation: rise .6s .12s ease both; }
    .tag b { color: #e2e8f0; font-weight: 600; }
    .cta { display: flex; gap: 1rem; justify-content: center; flex-wrap: wrap;
        animation: rise .6s .18s ease both; }
    .btn { padding: .85rem 1.8rem; border-radius: 12px; font-weight: 650;
        font-size: .98rem; text-decoration: none; transition: transform .15s, box-shadow .2s, background .2s; }
    .btn.primary { color: #fff;
        background: linear-gradient(120deg, #6366f1, #a855f7);
        box-shadow: 0 10px 30px -8px rgba(124,58,237,0.65); }
    .btn.primary:hover { transform: translateY(-2px);
        box-shadow: 0 16px 40px -8px rgba(124,58,237,0.8); }
    .btn.ghost { color: #e2e8f0;
        background: rgba(148,163,184,0.06);
        border: 1px solid rgba(148,163,184,0.22); backdrop-filter: blur(8px); }
    .btn.ghost:hover { transform: translateY(-2px); background: rgba(148,163,184,0.12); }
    .term {
        margin: 3rem auto 0; max-width: 560px; text-align: left;
        background: rgba(15,23,42,0.7); border: 1px solid rgba(148,163,184,0.16);
        border-radius: 14px; overflow: hidden; backdrop-filter: blur(10px);
        box-shadow: 0 24px 60px -24px rgba(0,0,0,0.7);
        animation: rise .6s .24s ease both;
    }
    .term .bar { display: flex; gap: .45rem; padding: .8rem 1rem;
        border-bottom: 1px solid rgba(148,163,184,0.12); }
    .term .bar i { width: 11px; height: 11px; border-radius: 50%; display: block; }
    .term .bar i:nth-child(1){ background:#ef4444; }
    .term .bar i:nth-child(2){ background:#f59e0b; }
    .term .bar i:nth-child(3){ background:#22c55e; }
    .term pre { margin: 0; padding: 1.1rem 1.3rem; font-size: .9rem; line-height: 1.85;
        font-family: ui-monospace, SFMono-Regular, Menlo, monospace; color: #cbd5e1; }
    .term .c { color: #64748b; }
    .term .g { color: #34d399; }
    .grid { display: grid; gap: 1.1rem; margin: 4rem 0 0;
        grid-template-columns: repeat(auto-fit, minmax(230px, 1fr)); }
    .feat { position: relative; text-align: left; padding: 1.5rem;
        background: rgba(148,163,184,0.05);
        border: 1px solid rgba(148,163,184,0.14);
        border-radius: 14px; overflow: hidden;
        transition: transform .18s, border-color .2s, background .2s; }
    .feat::before { content: ""; position: absolute; inset: 0 0 auto 0; height: 2px;
        background: linear-gradient(90deg, #6366f1, #ec4899);
        transform: scaleX(0); transform-origin: left; transition: transform .25s; }
    .feat:hover { transform: translateY(-4px); background: rgba(148,163,184,0.09);
        border-color: rgba(129,140,248,0.5); }
    .feat:hover::before { transform: scaleX(1); }
    .feat h3 { font-size: 1.02rem; color: #f1f5f9; margin-bottom: .45rem; }
    .feat p { font-size: .88rem; color: #94a3b8; line-height: 1.55; }
    .foot { margin-top: 4.5rem; padding-top: 2rem;
        border-top: 1px solid rgba(148,163,184,0.12); }
    .foot .links { display: flex; gap: 2rem; justify-content: center; flex-wrap: wrap; }
    .foot a { color: #818cf8; text-decoration: none; font-size: .9rem; }
    .foot a:hover { color: #a5b4fc; }
    .foot .ver { margin-top: 1.2rem; color: #475569; font-size: .8rem; }
    @keyframes rise { from { opacity: 0; transform: translateY(16px); } to { opacity: 1; transform: none; } }
    @media (max-width: 640px) { .wrap { padding-top: 4.5rem; } }
@endsection

@section("content")
    <div class="glow a"></div>
    <div class="glow b"></div>
    <div class="wrap">
        <span class="badge"><span class="dot"></span> Ignite v{{ .frameworkVersion }} &middot; {{ .goVersion }}</span>

        <h1 class="logo">Ignite</h1>
        <p class="tag">Your application is <b>live</b>. A full authentication flow ships out of the box &mdash; <b>register, log in, and a protected dashboard</b>, ready to build on.</p>

        <div class="cta">
            <a class="btn primary" href="/register">Get started &rarr;</a>
            <a class="btn ghost" href="/login">Log in</a>
        </div>

        <div class="term">
            <div class="bar"><i></i><i></i><i></i></div>
            <pre><span class="c"># scaffold, migrate, and serve</span>
ignite new myapp
<span class="g">./ignite</span> key:generate
<span class="g">./ignite</span> migrate
<span class="g">./ignite</span> serve</pre>
        </div>

        <div class="grid">
            <div class="feat"><h3>Routing</h3><p>Expressive routes with groups, named routes, and resource controllers.</p></div>
            <div class="feat"><h3>Authentication</h3><p>Register / login / logout and a protected dashboard, generated for you.</p></div>
            <div class="feat"><h3>Migrations</h3><p>Self-registering migrations across SQLite, MySQL, and PostgreSQL.</p></div>
            <div class="feat"><h3>Config</h3><p>.env mapped into a typed config repository, Laravel-style.</p></div>
            <div class="feat"><h3>Queue &amp; Jobs</h3><p>Background processing with retries, chains, and batches.</p></div>
            <div class="feat"><h3>Project Console</h3><p>Run everything through the project-local <code>./ignite</code> CLI.</p></div>
        </div>

        <div class="foot">
            <div class="links">
                <a href="https://github.com/sazzadh88/ignite">Documentation</a>
                <a href="https://github.com/sazzadh88/ignite">GitHub</a>
                <a href="https://github.com/sazzadh88/ignite">Ecosystem</a>
            </div>
            <p class="ver">Ignite v{{ .frameworkVersion }} &middot; {{ .goVersion }}</p>
        </div>
    </div>
@endsection
`, "__APP__", name)
}

func publicIndexTemplate(modulePath string) string {
	_ = modulePath
	return `package public

// This file serves as the entry point for the public directory.
// Static assets are served from this directory.
`
}

// MakeController generates a controller file.
// Supports path prefixes (e.g., "Api/UserController") and optional API mode.
// When api is true, only generates Index/Store/Show/Update/Destroy methods.
func MakeController(name string, api ...bool) error {
	isAPI := false
	if len(api) > 0 {
		isAPI = api[0]
	}

	// Split name into path and controller name
	parts := strings.Split(name, "/")
	controllerName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Http/Controllers", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	var content string
	if isAPI {
		content = fmt.Sprintf(`package Controllers

import "net/http"

// %s handles HTTP requests.
type %s struct {
	Controller
}

// Index handles GET requests.
func (c *%s) Index(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Store handles POST requests.
func (c *%s) Store(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Show handles GET requests for a single resource.
func (c *%s) Show(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Update handles PUT/PATCH requests.
func (c *%s) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Destroy handles DELETE requests.
func (c *%s) Destroy(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}
`, controllerName, controllerName, controllerName, controllerName, controllerName, controllerName, controllerName)
	} else {
		content = fmt.Sprintf(`package Controllers

import "net/http"

// %s handles HTTP requests.
type %s struct {
	Controller
}

// Index handles GET requests.
func (c *%s) Index(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Create handles GET requests for the create form.
func (c *%s) Create(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Store handles POST requests.
func (c *%s) Store(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Show handles GET requests for a single resource.
func (c *%s) Show(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Edit handles GET requests for the edit form.
func (c *%s) Edit(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Update handles PUT/PATCH requests.
func (c *%s) Update(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}

// Destroy handles DELETE requests.
func (c *%s) Destroy(w http.ResponseWriter, r *http.Request) {
	// TODO: implement
}
`, controllerName, controllerName, controllerName, controllerName, controllerName, controllerName, controllerName, controllerName, controllerName)
	}

	return os.WriteFile(filepath.Join(dir, controllerName+".go"), []byte(content), 0644)
}

// MakeModel generates a model file.
func MakeModel(name string) error {
	dir := "app/Models"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tableName := strings.ToLower(name) + "s"
	content := fmt.Sprintf(`package Models

// %s represents the %s model.
type %s struct {
	ID        uint64
	CreatedAt string
	UpdatedAt string
}

// Table returns the table name.
func (m *%s) Table() string {
	return "%s"
}

// Fillable returns mass-assignable fields.
func (m *%s) Fillable() []string {
	return []string{}
}
`, name, name, name, name, tableName, name)

	return os.WriteFile(filepath.Join(dir, name+".go"), []byte(content), 0644)
}

// MakeMiddleware generates a middleware file.
func MakeMiddleware(name string) error {
	dir := "app/Http/Middleware"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := fmt.Sprintf(`package Middleware

import "net/http"

// %s middleware.
type %s struct{}

// Handle processes the request.
func (m *%s) Handle(w http.ResponseWriter, r *http.Request, next http.Handler) {
	// TODO: implement middleware logic
	next.ServeHTTP(w, r)
}
`, name, name, name)

	return os.WriteFile(filepath.Join(dir, name+".go"), []byte(content), 0644)
}

// MakeMigration generates a migration file.
func MakeMigration(name string) error {
	dir := "database/migrations"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	timestamp := time.Now().Format("2006_01_02_150405")
	lower := strings.ToLower(name)
	migrationName := fmt.Sprintf("%s_%s", timestamp, lower)
	filename := migrationName + ".go"
	structName := toPascal(name)
	table := tableFromMigrationName(lower)

	content := fmt.Sprintf(`package migrations

import "github.com/sazzadh88/ignite/schema"

func init() {
	schema.RegisterMigration("%s", &%s{})
}

// %s migration.
type %s struct{}

// Up runs the migration.
func (m *%s) Up(s *schema.Schema) error {
	return s.Create("%s", func(t *schema.Blueprint) {
		t.ID()
		t.Timestamps()
	})
}

// Down reverses the migration.
func (m *%s) Down(s *schema.Schema) error {
	return s.DropIfExists("%s")
}
`, migrationName, structName, structName, structName, structName, table, structName, table)

	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
}

// tableFromMigrationName extracts the target table from a conventional
// migration name, e.g. "create_posts_table" -> "posts", "create_posts" ->
// "posts". Non-conventional names fall back to the raw name.
func tableFromMigrationName(lower string) string {
	if m := regexp.MustCompile(`^create_(.+?)_table$`).FindStringSubmatch(lower); m != nil {
		return m[1]
	}
	if m := regexp.MustCompile(`^create_(.+)$`).FindStringSubmatch(lower); m != nil {
		return m[1]
	}
	return lower
}

// MakeSeeder generates a seeder file.
func MakeSeeder(name string) error {
	dir := "database/seeders"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := fmt.Sprintf(`package seeders

// %s seeds the database.
type %s struct{}

// Run executes the seeder.
func (s *%s) Run() {
	// TODO: implement seeding logic
}
`, name, name, name)

	return os.WriteFile(filepath.Join(dir, name+".go"), []byte(content), 0644)
}

// MakeRequest generates a form request validation file.
func MakeRequest(name string) error {
	// Split name into path and request name
	parts := strings.Split(name, "/")
	requestName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Http/Requests", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`package Requests

// %sRequest handles validation for %s operations.
type %sRequest struct{}

// Rules returns validation rules.
func (r *%sRequest) Rules() map[string]string {
	return map[string]string{}
}

// Authorize checks if the user is authorized to make this request.
func (r *%sRequest) Authorize() bool {
	return true
}

// Messages returns custom validation messages.
func (r *%sRequest) Messages() map[string]string {
	return nil
}
`, requestName, requestName, requestName, requestName, requestName, requestName)

	return os.WriteFile(filepath.Join(dir, requestName+".go"), []byte(content), 0644)
}

// MakeCommand generates a console command file.
func MakeCommand(name string) error {
	// Split name into path and command name
	parts := strings.Split(name, "/")
	commandName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Console/Commands", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	signature := strings.ToLower(strings.ReplaceAll(commandName, "Command", ""))
	if !strings.Contains(signature, ":") {
		signature = "command:" + signature
	}

	content := fmt.Sprintf(`package Commands

// %s is a console command.
type %s struct{}

// Signature returns the command signature.
func (c *%s) Signature() string {
	return "%s"
}

// Description returns the command description.
func (c *%s) Description() string {
	return "Command description"
}

// Handle executes the command.
func (c *%s) Handle() error {
	// TODO: implement
	return nil
}
`, commandName, commandName, commandName, signature, commandName, commandName)

	return os.WriteFile(filepath.Join(dir, commandName+".go"), []byte(content), 0644)
}

// MakeEvent generates an event file.
func MakeEvent(name string) error {
	// Split name into path and event name
	parts := strings.Split(name, "/")
	eventName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Events", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`package Events

// %s represents an event.
type %s struct {
	// Add event data fields
}
`, eventName, eventName)

	return os.WriteFile(filepath.Join(dir, eventName+".go"), []byte(content), 0644)
}

// MakeListener generates an event listener file.
func MakeListener(name string) error {
	// Split name into path and listener name
	parts := strings.Split(name, "/")
	listenerName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Listeners", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`package Listeners

// %s handles events.
type %s struct{}

// Handle processes the event.
func (l *%s) Handle(event any) error {
	// TODO: implement
	return nil
}

// ShouldQueue determines if the listener should be queued.
func (l *%s) ShouldQueue() bool {
	return false
}
`, listenerName, listenerName, listenerName, listenerName)

	return os.WriteFile(filepath.Join(dir, listenerName+".go"), []byte(content), 0644)
}

// MakeJob generates a queue job file.
func MakeJob(name string) error {
	// Split name into path and job name
	parts := strings.Split(name, "/")
	jobName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Jobs", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`package Jobs

import "time"

// %s is a queue job.
type %s struct{}

// Handle executes the job.
func (j *%s) Handle() error {
	// TODO: implement
	return nil
}

// Queue returns the queue name.
func (j *%s) Queue() string {
	return "default"
}

// Tries returns the number of attempts.
func (j *%s) Tries() int {
	return 3
}

// Timeout returns the job timeout.
func (j *%s) Timeout() time.Duration {
	return 30 * time.Second
}
`, jobName, jobName, jobName, jobName, jobName, jobName)

	return os.WriteFile(filepath.Join(dir, jobName+".go"), []byte(content), 0644)
}

// MakePolicy generates a policy file for authorization.
func MakePolicy(name string) error {
	// Split name into path and policy name
	parts := strings.Split(name, "/")
	policyName := parts[len(parts)-1]
	var subPath string
	if len(parts) > 1 {
		subPath = filepath.Join(parts[:len(parts)-1]...)
	}

	dir := filepath.Join("app/Policies", subPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	content := fmt.Sprintf(`package Policies

// %s handles authorization for resources.
type %s struct{}

// ViewAny determines if the user can view any models.
func (p *%s) ViewAny(user any) bool {
	return true
}

// View determines if the user can view the model.
func (p *%s) View(user any, model any) bool {
	return true
}

// Create determines if the user can create models.
func (p *%s) Create(user any) bool {
	return true
}

// Update determines if the user can update the model.
func (p *%s) Update(user any, model any) bool {
	return true
}

// Delete determines if the user can delete the model.
func (p *%s) Delete(user any, model any) bool {
	return true
}
`, policyName, policyName, policyName, policyName, policyName, policyName, policyName)

	return os.WriteFile(filepath.Join(dir, policyName+".go"), []byte(content), 0644)
}

func toPascal(s string) string {
	parts := strings.Split(s, "_")
	for i, part := range parts {
		if len(part) > 0 {
			parts[i] = strings.ToUpper(part[:1]) + part[1:]
		}
	}
	return strings.Join(parts, "")
}
