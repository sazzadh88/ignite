package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/sazzadh88/ignite/foundation"
	"github.com/sazzadh88/ignite/internal/scaffold"
)

// projectCommands are commands that operate on an existing application. They
// must run through the project's own console (./ignite) so they use the
// framework version pinned in that project's go.mod — not this global binary.
var projectCommands = map[string]bool{
	"serve": true, "key:generate": true, "route:list": true,
	"make:controller": true, "make:model": true, "make:middleware": true,
	"make:migration": true, "make:seeder": true, "make:request": true,
	"make:command": true, "make:event": true, "make:listener": true,
	"make:job": true, "make:policy": true,
	"migrate": true, "migrate:rollback": true, "migrate:refresh": true,
	"migrate:fresh": true, "migrate:reset": true, "migrate:status": true,
	"db:seed": true,
	"queue:work": true, "queue:listen": true, "queue:restart": true,
	"schedule:run": true, "schedule:work": true,
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]
	args := os.Args[2:]
	_, positional := parseFlags(args)

	if projectCommands[command] {
		redirectToProjectConsole(command)
		return
	}

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
		fmt.Printf("    ./ignite key:generate\n")
		fmt.Printf("    ./ignite serve\n\n")

	case "version", "--version", "-v":
		printBanner()

	case "self-update", "upgrade":
		fmt.Println("  Upgrading Ignite...")
		cmd := exec.Command("go", "install", "github.com/sazzadh88/ignite/cmd/ignite@latest")
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "  Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("  ✓ Ignite upgraded to latest version.")

	case "help", "--help", "-h":
		printUsage()

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

// redirectToProjectConsole tells the user to run project commands through the
// project-local console instead of the global binary (Rails bin/rails model).
func redirectToProjectConsole(command string) {
	fmt.Fprintf(os.Stderr, "\n  '%s' is a project command — it does not run from the global ignite binary.\n\n", command)
	fmt.Fprintf(os.Stderr, "  Run it from your project's console so it uses the framework\n")
	fmt.Fprintf(os.Stderr, "  version pinned in that project's go.mod:\n\n")
	fmt.Fprintf(os.Stderr, "    ./ignite %s\n\n", strings.Join(os.Args[1:], " "))
	fmt.Fprintf(os.Stderr, "  The global 'ignite' binary only creates new projects (ignite new <name>).\n\n")
	os.Exit(1)
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

func printBanner() {
	lines := []string{
		`  ___            _ _       `,
		` |_ _|__ _ _ _ (_) |_ ___ `,
		`  | |/ _` + "`" + ` | ' \| |  _/ -_)`,
		` |___\__, |_||_|_|\__\___|`,
		`     |___/                `,
	}
	colors := []string{
		"\033[38;5;196m", // red
		"\033[38;5;208m", // orange
		"\033[38;5;226m", // yellow
		"\033[38;5;46m",  // green
		"\033[38;5;21m",  // blue
	}
	reset := "\033[0m"

	fmt.Println()
	for i, line := range lines {
		fmt.Printf("%s%s%s\n", colors[i%len(colors)], line, reset)
	}
	fmt.Printf("\033[38;5;141m v%s%s\n", foundation.Version, reset)
	fmt.Printf("\033[38;5;245m Web artisan meets Go performance.%s\n", reset)
}

func printUsage() {
	printBanner()
	fmt.Println(`
Usage:
  ignite <command>                Global commands (project creation)
  ./ignite <command>              Project commands (run from a project)

Global Commands:
  new <name>            Create a new Ignite application
  version               Display framework version
  upgrade               Upgrade the global ignite binary

Project Commands (run via ./ignite inside your project):
  serve                 Start the development server [--host --port]
  key:generate          Set the application encryption key
  route:list            List registered routes
  make:<type> <Name>    Scaffold controllers, models, migrations, etc.
  migrate               Run pending migrations [migrate:rollback|fresh|...]
  db:seed               Run database seeders [--class=SeederName]
  queue:work            Process queued jobs [queue:listen|restart]
  schedule:run          Run scheduled tasks [schedule:work]

A project's ./ignite uses the framework version pinned in its go.mod,
so project commands never depend on this global binary's version.

Run 'ignite help' for more information.`)
}
