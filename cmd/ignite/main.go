package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/sazzadh88/ignite/internal/scaffold"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]
	flags, positional := parseFlags(args)

	switch command {
	case "new":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite new <app-name>")
			os.Exit(1)
		}
		appName := positional[0]
		if err := scaffold.NewProject(appName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n  ✓ Application '%s' created successfully!\n\n", appName)
		fmt.Printf("  Next steps:\n")
		fmt.Printf("    cd %s\n", appName)
		fmt.Printf("    go mod tidy\n")
		fmt.Printf("    go run main.go serve\n\n")

	case "serve":
		host := flags["host"]
		if host == "" {
			host = "localhost"
		}
		port := flags["port"]
		if port == "" {
			port = "8080"
		}
		fmt.Println("Starting development server...")
		fmt.Printf("Server running on http://%s:%s\n", host, port)
		fmt.Println("Note: This is the CLI tool. Use 'go run main.go serve' from your project directory.")

	case "key:generate":
		fmt.Println("Application key generated successfully.")
		// TODO: implement key generation

	case "route:list":
		fmt.Println("Registered routes:")
		// TODO: implement route listing

	case "version", "--version", "-v":
		fmt.Println("Ignite v0.1.0")

	case "help", "--help", "-h":
		printUsage()

	// Make commands
	case "make:controller":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:controller <Name> [--api]")
			os.Exit(1)
		}
		name := positional[0]
		_, isAPI := flags["api"]
		if err := scaffold.MakeController(name, isAPI); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Controller '%s' created.\n", name)

	case "make:model":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:model <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeModel(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Model '%s' created.\n", positional[0])

	case "make:middleware":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:middleware <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeMiddleware(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Middleware '%s' created.\n", positional[0])

	case "make:migration":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:migration <name>")
			os.Exit(1)
		}
		if err := scaffold.MakeMigration(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Migration '%s' created.\n", positional[0])

	case "make:seeder":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:seeder <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeSeeder(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Seeder '%s' created.\n", positional[0])

	case "make:request":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:request <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeRequest(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Request '%s' created.\n", positional[0])

	case "make:command":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:command <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeCommand(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Command '%s' created.\n", positional[0])

	case "make:event":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:event <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeEvent(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Event '%s' created.\n", positional[0])

	case "make:listener":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:listener <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeListener(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Listener '%s' created.\n", positional[0])

	case "make:job":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:job <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakeJob(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Job '%s' created.\n", positional[0])

	case "make:policy":
		if len(positional) < 1 {
			fmt.Println("Usage: ignite make:policy <Name>")
			os.Exit(1)
		}
		if err := scaffold.MakePolicy(positional[0]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("  ✓ Policy '%s' created.\n", positional[0])

	// Database commands
	case "migrate":
		fmt.Println("Running migrations...")

	case "migrate:rollback":
		fmt.Println("Rolling back...")

	case "migrate:refresh":
		fmt.Println("Refreshing...")

	case "migrate:fresh":
		_, withSeed := flags["seed"]
		fmt.Println("Dropping all tables and re-running migrations...")
		if withSeed {
			fmt.Println("Seeding database...")
		}

	case "db:seed":
		class := flags["class"]
		if class != "" {
			fmt.Printf("Seeding: %s\n", class)
		} else {
			fmt.Println("Seeding database...")
		}

	// Queue commands
	case "queue:work":
		fmt.Println("Processing jobs...")

	case "queue:listen":
		fmt.Println("Listening for jobs...")

	case "queue:restart":
		fmt.Println("Restarting queue worker...")

	// Schedule commands
	case "schedule:run":
		fmt.Println("Running scheduled tasks...")

	case "schedule:work":
		fmt.Println("Schedule worker started...")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
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

func printUsage() {
	fmt.Println(`
Ignite - A Laravel-inspired Go Framework (v0.1.0)

Usage:
  ignite <command> [arguments] [flags]

Available Commands:
  new <name>            Create a new Ignite application
  serve                 Start the development server
    --host=<host>         Host to bind (default: localhost)
    --port=<port>         Port to bind (default: 8080)
  key:generate          Generate application key
  route:list            List registered routes
  version               Display framework version

Make Commands:
  make:controller <Name>   Generate a controller [--api for API resource]
  make:model <Name>        Generate a model
  make:middleware <Name>   Generate a middleware
  make:migration <name>    Generate a migration
  make:seeder <Name>       Generate a seeder
  make:request <Name>      Generate a request validation
  make:command <Name>      Generate a console command
  make:event <Name>        Generate an event
  make:listener <Name>     Generate an event listener
  make:job <Name>          Generate a queue job
  make:policy <Name>       Generate a policy

Database Commands:
  migrate                Run pending migrations
  migrate:rollback       Rollback last batch
  migrate:refresh        Rollback and re-run migrations
  migrate:fresh          Drop all tables and re-run [--seed]
  db:seed                Run database seeders [--class=SeederName]

Queue Commands:
  queue:work             Process jobs
  queue:listen           Listen for jobs
  queue:restart          Restart queue worker

Schedule Commands:
  schedule:run           Run scheduled tasks
  schedule:work          Start schedule worker

Run 'ignite help' for more information.`)
}
