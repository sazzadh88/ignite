<p align="center">
  <h1 align="center">Ignite</h1>
  <p align="center"><strong>Web artisan meets Go performance.</strong></p>
</p>

<p align="center">
  <a href="https://github.com/sazzadh88/ignite/actions"><img src="https://github.com/sazzadh88/ignite/workflows/tests/badge.svg" alt="Build Status"></a>
  <a href="https://pkg.go.dev/github.com/sazzadh88/ignite"><img src="https://pkg.go.dev/badge/github.com/sazzadh88/ignite.svg" alt="Go Reference"></a>
  <a href="https://goreportcard.com/report/github.com/sazzadh88/ignite"><img src="https://goreportcard.com/badge/github.com/sazzadh88/ignite" alt="Go Report Card"></a>
  <a href="https://github.com/sazzadh88/ignite/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-brightgreen.svg" alt="License"></a>
</p>

## About Ignite

Ignite is a full-stack web application framework for Go where web artisans meet Go performance. It provides an expressive, intuitive API for building modern web applications with minimal configuration and maximum productivity.

With Ignite, you get everything you need to build production-ready applications: routing, ORM with relationships, authentication, authorization, validation, queues, events, scheduling, caching, mail, notifications, and more. All powered by pure Go with zero external runtime dependencies.

Built with 33 carefully crafted packages, over 812 tests, and 43,000+ lines of production-grade code, Ignite is designed for developers who value clean code, convention over configuration, and rapid development without sacrificing performance or type safety.

## Official Documentation

Documentation for the framework can be found in the [docs directory](docs/). Start with the [Getting Started Guide](docs/getting-started.md) to build your first Ignite application.

## Quick Start

Install the Ignite CLI:

```bash
go install github.com/sazzadh88/ignite/cmd/ignite@latest
```

### PATH Setup

After installation, ensure Go's bin directory is in your PATH:

**macOS / Linux (zsh):**
```bash
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.zshrc
source ~/.zshrc
```

**macOS / Linux (bash):**
```bash
echo 'export PATH=$PATH:$HOME/go/bin' >> ~/.bashrc
source ~/.bashrc
```

**Windows (PowerShell):**
```powershell
[Environment]::SetEnvironmentVariable("Path", $env:Path + ";$env:USERPROFILE\go\bin", "User")
```

**Windows (CMD):**
```cmd
setx PATH "%PATH%;%USERPROFILE%\go\bin"
```

Verify installation:
```bash
ignite version
```

Create a new application:

```bash
ignite new myapp
cd myapp
go mod tidy
```

Start the development server:

```bash
go run main.go serve
```

Your application is now running at `http://localhost:8080`.

## Features

Ignite comes with a comprehensive suite of features for building modern web applications:

### Core Framework

| Feature | Description |
|---------|-------------|
| **Routing** | Expressive routing with groups, middleware, resource controllers, named routes, and parameter binding |
| **HTTP Kernel** | Request/response lifecycle with middleware pipeline and dependency injection |
| **Service Container** | Powerful IoC container with automatic dependency resolution |
| **Configuration** | Environment-based configuration with type-safe access |

### Database & ORM

| Feature | Description |
|---------|-------------|
| **Eloquent ORM** | Fluent, expressive ORM with relationships (hasOne, hasMany, belongsTo, belongsToMany) |
| **Query Builder** | Database-agnostic query builder with fluent interface |
| **Migrations** | Version control for your database schema |
| **Schema Builder** | Expressive API for creating and modifying tables |
| **Seeders** | Populate your database with test data |
| **Model Features** | Soft deletes, scopes, observers, events, eager loading, and more |

### Security & Authentication

| Feature | Description |
|---------|-------------|
| **Authentication** | Session and token-based authentication with multiple guards |
| **API Tokens** | Built-in API token management for your users |
| **Authorization** | Gates and policies for fine-grained access control |
| **Validation** | 40+ validation rules with custom error messages and form requests |
| **Encryption** | AES-256-GCM encryption for sensitive data |
| **Hashing** | HMAC-SHA256 password hashing |
| **CSRF Protection** | Cross-site request forgery protection middleware |
| **Rate Limiting** | Configurable rate limiting with named limiters |

### Frontend & Views

| Feature | Description |
|---------|-------------|
| **Blade Templates** | Powerful templating engine with `@extends`, `@section`, `@if`, `@foreach`, and more |
| **View Composer** | Bind data to views using composers and creators |
| **Asset Compilation** | Built-in support for asset bundling workflows |
| **Sessions** | File, memory, and cookie-based session storage |

### Background Processing

| Feature | Description |
|---------|-------------|
| **Queue System** | Dispatch jobs to background queues with chains and batches |
| **Job Middleware** | Apply middleware to your queued jobs |
| **Failed Jobs** | Automatic retry with exponential backoff |
| **Task Scheduling** | Cron-like task scheduling with fluent API |

### Communication

| Feature | Description |
|---------|-------------|
| **Mail** | Send emails using SMTP, log, or array drivers with markdown templates |
| **Notifications** | Multi-channel notifications (mail, database, Slack) |
| **Broadcasting** | Real-time event broadcasting with channel authentication |
| **Events** | Event-driven architecture with listeners and subscribers |

### HTTP & APIs

| Feature | Description |
|---------|-------------|
| **HTTP Client** | Fluent HTTP client with retry logic and connection pooling |
| **API Resources** | Transform models and collections for API responses |
| **Pagination** | Automatic pagination with customizable page sizes |
| **CORS** | Cross-origin resource sharing middleware |

### Storage & Caching

| Feature | Description |
|---------|-------------|
| **File Storage** | Local and public disk drivers with unified API |
| **Cache** | Multiple backends (file, memory) with tags and locks |
| **Cookie Management** | Secure cookie handling with encryption |

### Developer Tools

| Feature | Description |
|---------|-------------|
| **Collections** | Generic collections with 40+ helper methods (Map, Filter, Reduce, etc.) |
| **String Helpers** | Camel case, snake case, slug, plural, singular, and more |
| **Array Helpers** | Pluck, flatten, dot notation access, and more |
| **Pipeline** | Process data through a series of operations |
| **Testing Utilities** | HTTP assertions, database factories, and mocking helpers |
| **Error Handling** | Centralized exception handling with custom error pages |
| **Logging** | Structured logging with multiple channels and levels |

## CLI Commands

The Ignite CLI provides artisan-like commands for rapid development:

```bash
# Application
ignite new <name>           # Create a new Ignite application
ignite serve                # Start the development server
ignite env                  # Display the current environment

# Make Commands
ignite make:controller      # Create a new controller
ignite make:model           # Create a new model
ignite make:migration       # Create a new migration
ignite make:seeder          # Create a new seeder
ignite make:middleware      # Create a new middleware
ignite make:request         # Create a new form request
ignite make:policy          # Create a new policy
ignite make:job             # Create a new job
ignite make:event           # Create a new event
ignite make:listener        # Create a new event listener
ignite make:mail            # Create a new mail template
ignite make:notification    # Create a new notification
ignite make:resource        # Create a new API resource

# Database
ignite migrate              # Run database migrations
ignite migrate:rollback     # Rollback the last migration
ignite migrate:fresh        # Drop all tables and re-run migrations
ignite migrate:reset        # Rollback all migrations
ignite migrate:status       # Show migration status
ignite db:seed              # Seed the database

# Queue
ignite queue:work           # Start processing jobs
ignite queue:listen         # Listen for jobs
ignite queue:failed         # List failed jobs
ignite queue:retry          # Retry failed jobs
ignite queue:flush          # Flush failed jobs

# Cache
ignite cache:clear          # Clear the application cache
ignite cache:forget         # Remove an item from the cache

# Route
ignite route:list           # List all registered routes
ignite route:cache          # Cache the routes for faster registration

# Config
ignite config:cache         # Cache the configuration
ignite config:clear         # Clear the configuration cache

# Schedule
ignite schedule:run         # Run scheduled commands
ignite schedule:list        # List scheduled commands
```

## Project Structure

A fresh Ignite application has the following structure:

```
myapp/
├── app/
│   ├── Http/
│   │   ├── Controllers/
│   │   │   └── Controller.go
│   │   ├── Middleware/
│   │   │   └── Authenticate.go
│   │   ├── Requests/
│   │   └── Kernel.go
│   ├── Models/
│   │   └── User.go
│   ├── Providers/
│   │   ├── AppServiceProvider.go
│   │   ├── AuthServiceProvider.go
│   │   └── RouteServiceProvider.go
│   ├── Jobs/
│   ├── Events/
│   ├── Listeners/
│   ├── Notifications/
│   ├── Policies/
│   └── Mail/
├── bootstrap/
│   └── app.go
├── config/
│   ├── app.go
│   ├── database.go
│   ├── auth.go
│   ├── cache.go
│   ├── mail.go
│   ├── queue.go
│   └── session.go
├── database/
│   ├── migrations/
│   ├── seeders/
│   └── factories/
├── public/
│   ├── index.go
│   ├── css/
│   └── js/
├── resources/
│   ├── views/
│   │   └── welcome.ignite.html
│   └── assets/
├── routes/
│   ├── web.go
│   ├── api.go
│   └── channels.go
├── storage/
│   ├── app/
│   ├── framework/
│   │   ├── cache/
│   │   ├── sessions/
│   │   └── views/
│   └── logs/
├── tests/
│   ├── Feature/
│   └── Unit/
├── .env
├── .env.example
├── go.mod
├── go.sum
└── main.go
```

## Requirements

- Go 1.22 or higher
- SQLite3 (for testing, optional for production)

## Contributing

Thank you for considering contributing to the Ignite framework! Please read our [Contributing Guide](CONTRIBUTING.md) to learn about our development process, how to propose bugfixes and improvements, and how to build and test your changes.

## Code of Conduct

In order to ensure that the Ignite community is welcoming to all, please review and abide by the [Code of Conduct](CODE_OF_CONDUCT.md).

## Security Vulnerabilities

If you discover a security vulnerability within Ignite, please send an email to security@igniteframework.dev. All security vulnerabilities will be promptly addressed.

## License

The Ignite framework is open-sourced software licensed under the [MIT license](LICENSE).

---

<p align="center">
  <strong>Built with ❤️ for the Go community</strong>
</p>
