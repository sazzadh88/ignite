package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
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
		"bootstrap",
		"config",
		"database/migrations",
		"database/seeders",
		"database/factories",
		"public",
		"resources/views",
		"resources/views/layouts",
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
		"config/app.go":                configAppTemplate(),
		"config/database.go":           configDatabaseTemplate(),
		"routes/web.go":                routesWebTemplate(),
		"routes/api.go":                routesAPITemplate(),
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

func mainTemplate(modulePath string) string {
	return fmt.Sprintf(`package main

import (
	"github.com/sazzadh88/ignite/foundation"
	"github.com/sazzadh88/ignite/routing"
	"%s/routes"
)

func main() {
	app := foundation.NewApplication(".")
	app.Bootstrap()

	// Register routes
	routes.RegisterWeb()
	routes.RegisterAPI()

	// Run the application
	app.Run(routing.DefaultRouter)
}
`, modulePath)
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

DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=%s
DB_USERNAME=root
DB_PASSWORD=

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

func configAppTemplate() string {
	return `package config

// App configuration
var App = map[string]any{
	"name":  "Ignite",
	"env":   "local",
	"debug": true,
	"url":   "http://localhost",
	"port":  8080,
}
`
}

func configDatabaseTemplate() string {
	return `package config

// Database configuration
var Database = map[string]any{
	"default": "mysql",
	"connections": map[string]any{
		"mysql": map[string]any{
			"driver":   "mysql",
			"host":     "127.0.0.1",
			"port":     3306,
			"database": "ignite",
			"username": "root",
			"password": "",
		},
		"sqlite": map[string]any{
			"driver":   "sqlite",
			"database": "database/database.sqlite",
		},
	},
}
`
}

func routesWebTemplate() string {
	return `package routes

import (
	"net/http"
	"runtime"

	"github.com/sazzadh88/ignite/foundation"
	"github.com/sazzadh88/ignite/routing"
	"github.com/sazzadh88/ignite/view"
)

// RegisterWeb defines web routes.
// These routes are assigned the "web" middleware group.
func RegisterWeb() {
	r := routing.DefaultRouter

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
}
`
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

// User represents the user model.
type User struct {
	ID        uint64
	Name      string
	Email     string
	Password  string
	CreatedAt string
	UpdatedAt string
}

// Table returns the table name.
func (u *User) Table() string {
	return "users"
}

// Fillable returns mass-assignable fields.
func (u *User) Fillable() []string {
	return []string{"name", "email", "password"}
}

// Hidden returns fields hidden from serialization.
func (u *User) Hidden() []string {
	return []string{"password"}
}
`
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
	return fmt.Sprintf(`@extends("layouts.app")

@section("title")
%s - Welcome
@endsection

@section("styles")
    .container {
        display: flex;
        align-items: center;
        justify-content: center;
        min-height: 100vh;
        text-align: center;
        padding: 3rem;
    }
    .content { max-width: 720px; }
    .logo {
        font-size: 4rem;
        font-weight: 800;
        background: linear-gradient(135deg, #6366f1, #a855f7, #ec4899);
        -webkit-background-clip: text;
        -webkit-text-fill-color: transparent;
        background-clip: text;
        margin-bottom: 0.5rem;
    }
    .tagline { font-size: 1.25rem; color: #94a3b8; margin-bottom: 3rem; }
    .cards {
        display: grid;
        grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
        gap: 1.5rem;
        margin-bottom: 3rem;
    }
    .card {
        background: #1e293b;
        border: 1px solid #334155;
        border-radius: 12px;
        padding: 1.5rem;
        text-align: left;
        transition: border-color 0.2s, transform 0.2s;
    }
    .card:hover { border-color: #6366f1; transform: translateY(-2px); }
    .card h3 { font-size: 1rem; color: #f1f5f9; margin-bottom: 0.5rem; }
    .card p { font-size: 0.875rem; color: #64748b; line-height: 1.5; }
    .links { display: flex; gap: 2rem; justify-content: center; flex-wrap: wrap; }
    .links a { color: #818cf8; text-decoration: none; font-size: 0.9rem; transition: color 0.2s; }
    .links a:hover { color: #a5b4fc; }
    .version { margin-top: 3rem; color: #475569; font-size: 0.8rem; }
@endsection

@section("content")
    <div class="container">
        <div class="content">
            <h1 class="logo">%s</h1>
            <p class="tagline">Built with Ignite — Web artisan meets Go performance</p>

            <div class="cards">
                <div class="card">
                    <h3>Routing</h3>
                    <p>Expressive route definitions with groups, middleware, and resource controllers.</p>
                </div>
                <div class="card">
                    <h3>ORM</h3>
                    <p>Eloquent-style models with relationships, scopes, and query builder.</p>
                </div>
                <div class="card">
                    <h3>Authentication</h3>
                    <p>Session and token guards with policies and gates out of the box.</p>
                </div>
                <div class="card">
                    <h3>Queue &amp; Jobs</h3>
                    <p>Background job processing with retries, chains, and batches.</p>
                </div>
            </div>

            <div class="links">
                <a href="https://github.com/sazzadh88/ignite">Documentation</a>
                <a href="https://github.com/sazzadh88/ignite">GitHub</a>
                <a href="https://github.com/sazzadh88/ignite">Ecosystem</a>
            </div>

            <p class="version">Ignite v{{ .frameworkVersion }} &middot; {{ .goVersion }}</p>
        </div>
    </div>
@endsection
`, name, name)
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
	filename := fmt.Sprintf("%s_%s.go", timestamp, strings.ToLower(name))
	structName := toPascal(name)

	content := fmt.Sprintf(`package migrations

// %s migration.
type %s struct{}

// Up runs the migration.
func (m *%s) Up() {
	// schema.Create("%s", func(t *Blueprint) {
	//     t.ID()
	//     t.Timestamps()
	// })
}

// Down reverses the migration.
func (m *%s) Down() {
	// schema.DropIfExists("%s")
}
`, structName, structName, structName, strings.ToLower(name)+"s", structName, strings.ToLower(name)+"s")

	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
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
