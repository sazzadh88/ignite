package foundation

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/sazzadh88/ignite/config"
	"github.com/sazzadh88/ignite/container"
	"github.com/sazzadh88/ignite/encryption"
	"github.com/sazzadh88/ignite/internal/scaffold"
)

const Version = "0.1.5"

// ServiceProvider defines the contract for service providers.
type ServiceProvider interface {
	Register(app *Application)
	Boot(app *Application)
}

// DeferredProvider is loaded only when needed.
type DeferredProvider interface {
	ServiceProvider
	Provides() []string
}

// Application is the main application kernel.
type Application struct {
	*container.Container

	mu             sync.RWMutex
	basePath       string
	providers      []ServiceProvider
	booted         bool
	config         *config.Repository
	environment    string
}

// NewApplication creates a new application instance.
func NewApplication(basePath string) *Application {
	app := &Application{
		Container: container.New(),
		basePath:  basePath,
		config:    config.New(),
	}

	app.registerBaseBindings()
	return app
}

func (app *Application) registerBaseBindings() {
	app.Instance("app", app)
	app.Instance("config", app.config)
	app.Instance("path.base", app.basePath)
}

// Bootstrap loads .env and config.
func (app *Application) Bootstrap() {
	envPath := filepath.Join(app.basePath, ".env")
	config.LoadEnv(envPath)

	app.environment = config.Env("APP_ENV", "local")
	app.loadConfiguration()

	app.bootEncrypter()
}

// bootEncrypter initializes the encryption facade from APP_KEY when the key
// is valid. A missing or invalid key is not fatal here (console commands such
// as key:generate must still run); enforcement happens in serve().
func (app *Application) bootEncrypter() {
	keyBytes, err := parseAppKey(config.Env("APP_KEY", ""))
	if err != nil {
		return
	}

	enc, err := encryption.NewEncrypter(keyBytes)
	if err != nil {
		return
	}

	encryption.Crypt = enc
	app.Instance("encrypter", enc)
}

// Register registers a service provider.
func (app *Application) Register(provider ServiceProvider) {
	app.mu.Lock()
	defer app.mu.Unlock()

	provider.Register(app)
	app.providers = append(app.providers, provider)

	if app.booted {
		provider.Boot(app)
	}
}

// Boot boots all registered providers.
func (app *Application) Boot() {
	app.mu.Lock()
	defer app.mu.Unlock()

	if app.booted {
		return
	}

	for _, provider := range app.providers {
		provider.Boot(app)
	}

	app.booted = true
}

// Config returns the config repository.
func (app *Application) Config() *config.Repository {
	return app.config
}

// Environment returns the current environment.
func (app *Application) Environment() string {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.environment
}

// IsLocal returns true if environment is local.
func (app *Application) IsLocal() bool {
	return app.Environment() == "local"
}

// IsProduction returns true if environment is production.
func (app *Application) IsProduction() bool {
	return app.Environment() == "production"
}

// IsTesting returns true if environment is testing.
func (app *Application) IsTesting() bool {
	return app.Environment() == "testing"
}

// RunningInConsole returns true if running in CLI mode.
func (app *Application) RunningInConsole() bool {
	return os.Getenv("GOFRAME_RUNNING_IN_CONSOLE") == "true"
}

// Version returns the framework version.
func (app *Application) Version() string {
	return Version
}

// BasePath returns the base path with optional relative path appended.
func (app *Application) BasePath(path ...string) string {
	if len(path) > 0 {
		return filepath.Join(app.basePath, path[0])
	}
	return app.basePath
}

// StoragePath returns the storage path.
func (app *Application) StoragePath(path ...string) string {
	base := filepath.Join(app.basePath, "storage")
	if len(path) > 0 {
		return filepath.Join(base, path[0])
	}
	return base
}

// ConfigPath returns the config path.
func (app *Application) ConfigPath(path ...string) string {
	base := filepath.Join(app.basePath, "config")
	if len(path) > 0 {
		return filepath.Join(base, path[0])
	}
	return base
}

// DatabasePath returns the database path.
func (app *Application) DatabasePath(path ...string) string {
	base := filepath.Join(app.basePath, "database")
	if len(path) > 0 {
		return filepath.Join(base, path[0])
	}
	return base
}

// ResourcePath returns the resources path.
func (app *Application) ResourcePath(path ...string) string {
	base := filepath.Join(app.basePath, "resources")
	if len(path) > 0 {
		return filepath.Join(base, path[0])
	}
	return base
}

// PublicPath returns the public path.
func (app *Application) PublicPath(path ...string) string {
	base := filepath.Join(app.basePath, "public")
	if len(path) > 0 {
		return filepath.Join(base, path[0])
	}
	return base
}

// IsBooted returns whether the application has booted.
func (app *Application) IsBooted() bool {
	app.mu.RLock()
	defer app.mu.RUnlock()
	return app.booted
}

// RouterInterface allows Run() to access the router without importing routing package.
type RouterInterface interface {
	http.Handler
	PrintRoutes()
	RouteList() string
}

// SetRouter sets the application router.
func (app *Application) SetRouter(router RouterInterface) {
	app.Instance("router", router)
}

// Run dispatches the application — either handles a CLI command or starts the HTTP server.
// This is the equivalent of Laravel's Kernel dispatch.
func (app *Application) Run(router ...RouterInterface) {
	app.Boot()

	// Resolve router
	var r RouterInterface
	if len(router) > 0 {
		r = router[0]
	} else if bound := app.Make("router"); bound != nil {
		r, _ = bound.(RouterInterface)
	}

	// Check for CLI command
	if len(os.Args) > 1 {
		command := os.Args[1]
		args := os.Args[2:]
		app.handleCommand(command, args, r)
		return
	}

	// No args: show help (like `php artisan`)
	app.printBanner()
	app.printHelp()
}

func (app *Application) handleCommand(command string, args []string, router RouterInterface) {
	flags, positional := parseFlags(args)

	switch command {
	case "serve":
		host := flags["host"]
		if host == "" {
			host = "localhost"
		}
		port := flags["port"]
		if port == "" {
			port = fmt.Sprintf("%d", app.config.GetInt("app.port", 8080))
		}
		app.serve(router, host, port)

	case "route:list":
		if router != nil {
			fmt.Println()
			fmt.Println("  Registered Routes:")
			fmt.Println()
			fmt.Print(router.RouteList())
		} else {
			fmt.Println("  No routes registered.")
		}

	case "env":
		fmt.Printf("  Environment: %s\n", app.Environment())

	case "version", "--version", "-v":
		fmt.Printf("  Ignite v%s\n", Version)

	case "key:generate":
		key, err := GenerateAppKey(app.basePath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		os.Setenv("APP_KEY", key)
		app.config.Set("app.key", key)
		app.bootEncrypter()
		fmt.Println("  Application key set successfully.")

	// Make commands
	case "make:controller":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:controller <Name> [--api]")
			os.Exit(1)
		}
		name := positional[0]
		_, isAPI := flags["api"]
		if err := scaffold.MakeController(name, isAPI); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Controller '%s' created.\n", name)

	case "make:model":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:model <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeModel(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Model '%s' created.\n", positional[0])

	case "make:middleware":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:middleware <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeMiddleware(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Middleware '%s' created.\n", positional[0])

	case "make:migration":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:migration <name>")
			os.Exit(1)
		}
		if err := scaffold.MakeMigration(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Migration '%s' created.\n", positional[0])

	case "make:seeder":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:seeder <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeSeeder(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Seeder '%s' created.\n", positional[0])

	case "make:request":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:request <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeRequest(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Request '%s' created.\n", positional[0])

	case "make:command":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:command <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeCommand(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Command '%s' created.\n", positional[0])

	case "make:event":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:event <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeEvent(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Event '%s' created.\n", positional[0])

	case "make:listener":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:listener <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeListener(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Listener '%s' created.\n", positional[0])

	case "make:job":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:job <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeJob(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Job '%s' created.\n", positional[0])

	case "make:policy":
		if len(positional) < 1 {
			fmt.Println("  Usage: make:policy <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakePolicy(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Policy '%s' created.\n", positional[0])

	// Database commands
	case "migrate":
		fmt.Println("  Running migrations...")

	case "migrate:rollback":
		fmt.Println("  Rolling back...")

	case "migrate:refresh":
		fmt.Println("  Refreshing...")

	case "migrate:fresh":
		_, withSeed := flags["seed"]
		fmt.Println("  Dropping all tables and re-running migrations...")
		if withSeed {
			fmt.Println("  Seeding database...")
		}

	case "db:seed":
		class := flags["class"]
		if class != "" {
			fmt.Printf("  Seeding: %s\n", class)
		} else {
			fmt.Println("  Seeding database...")
		}

	// Queue commands
	case "queue:work":
		fmt.Println("  Processing jobs...")

	case "queue:listen":
		fmt.Println("  Listening for jobs...")

	case "queue:restart":
		fmt.Println("  Restarting queue worker...")

	// Schedule commands
	case "schedule:run":
		fmt.Println("  Running scheduled tasks...")

	case "schedule:work":
		fmt.Println("  Schedule worker started...")

	default:
		fmt.Fprintf(os.Stderr, "  Unknown command: %s\n\n", command)
		app.printHelp()
		os.Exit(1)
	}
}

func (app *Application) serve(router RouterInterface, host, port string) {
	if _, err := parseAppKey(config.Env("APP_KEY", "")); err != nil {
		fmt.Fprintf(os.Stderr, "\n  %s\n\n", missingAppKeyError)
		os.Exit(1)
	}

	name := app.config.GetString("app.name", "Ignite")

	fmt.Printf("\n  %s v%s\n", name, Version)
	fmt.Printf("  Environment: %s\n", app.Environment())

	if router != nil {
		fmt.Println()
		router.PrintRoutes()
	}

	addr := fmt.Sprintf("%s:%s", host, port)
	displayAddr := addr
	if host == "0.0.0.0" {
		displayAddr = fmt.Sprintf("localhost:%s", port)
	}

	fmt.Printf("\n  Server running on http://%s\n\n", displayAddr)

	var handler http.Handler
	if router != nil {
		handler = router
	} else {
		handler = http.DefaultServeMux
	}

	log.Fatal(http.ListenAndServe(addr, handler))
}

func (app *Application) printBanner() {
	name := app.config.GetString("app.name", "Ignite")
	fmt.Printf("\n  %s v%s\n\n", name, Version)
}

func (app *Application) printHelp() {
	fmt.Println("  Usage:")
	fmt.Println("    go run main.go <command> [arguments] [flags]")
	fmt.Println()
	fmt.Println("  Available commands:")
	fmt.Println("    serve                     Start the HTTP server")
	fmt.Println("      --host=<host>             Host to bind (default: localhost)")
	fmt.Println("      --port=<port>             Port to bind (default: from config)")
	fmt.Println("    route:list                Display registered routes")
	fmt.Println("    env                       Show current environment")
	fmt.Println("    version                   Show framework version")
	fmt.Println("    key:generate              Set the application encryption key")
	fmt.Println()
	fmt.Println("  Make commands:")
	fmt.Println("    make:controller <Name>    Generate a controller [--api]")
	fmt.Println("    make:model <Name>         Generate a model")
	fmt.Println("    make:middleware <Name>    Generate a middleware")
	fmt.Println("    make:migration <name>     Generate a migration")
	fmt.Println("    make:seeder <Name>        Generate a seeder")
	fmt.Println("    make:request <Name>       Generate a request validation")
	fmt.Println("    make:command <Name>       Generate a console command")
	fmt.Println("    make:event <Name>         Generate an event")
	fmt.Println("    make:listener <Name>      Generate an event listener")
	fmt.Println("    make:job <Name>           Generate a queue job")
	fmt.Println("    make:policy <Name>        Generate a policy")
	fmt.Println()
	fmt.Println("  Database commands:")
	fmt.Println("    migrate                   Run pending migrations")
	fmt.Println("    migrate:rollback          Rollback last batch")
	fmt.Println("    migrate:refresh           Rollback and re-run migrations")
	fmt.Println("    migrate:fresh             Drop all tables and re-run [--seed]")
	fmt.Println("    db:seed                   Run database seeders [--class=Name]")
	fmt.Println()
	fmt.Println("  Queue commands:")
	fmt.Println("    queue:work                Process jobs")
	fmt.Println("    queue:listen              Listen for jobs")
	fmt.Println("    queue:restart             Restart queue worker")
	fmt.Println()
	fmt.Println("  Schedule commands:")
	fmt.Println("    schedule:run              Run scheduled tasks")
	fmt.Println("    schedule:work             Start schedule worker")
	fmt.Println()
}

// parseFlags separates --key=value and --key value flags from positional arguments.
// Returns (flags map, positional args).
func parseFlags(args []string) (map[string]string, []string) {
	flags := make(map[string]string)
	var positional []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		if strings.HasPrefix(arg, "--") {
			// Remove leading --
			arg = arg[2:]

			// Check for --key=value format
			if idx := strings.Index(arg, "="); idx != -1 {
				key := arg[:idx]
				value := arg[idx+1:]
				flags[key] = value
			} else {
				// Check if next arg is the value
				if i+1 < len(args) && !strings.HasPrefix(args[i+1], "--") {
					flags[arg] = args[i+1]
					i++ // Skip next arg
				} else {
					// Boolean flag
					flags[arg] = "true"
				}
			}
		} else {
			positional = append(positional, arg)
		}
	}

	return flags, positional
}
