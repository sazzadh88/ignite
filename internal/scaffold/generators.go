package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// MakeController generates a controller file.
// Supports path prefixes (e.g., "Api/UserController") and optional API mode.
func MakeController(name string, api ...bool) error {
	isAPI := len(api) > 0 && api[0]

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

	stubName := "make/controller.go.stub"
	if isAPI {
		stubName = "make/controller-api.go.stub"
	}

	content := stub(stubName, map[string]string{"__NAME__": controllerName})
	return os.WriteFile(filepath.Join(dir, controllerName+".go"), []byte(content), 0644)
}

// MakeModel generates a model file.
func MakeModel(name string) error {
	dir := "app/Models"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	tableName := strings.ToLower(name) + "s"
	content := stub("make/model.go.stub", map[string]string{
		"__NAME__":  name,
		"__TABLE__": tableName,
	})
	return os.WriteFile(filepath.Join(dir, name+".go"), []byte(content), 0644)
}

// MakeMiddleware generates a middleware file.
func MakeMiddleware(name string) error {
	dir := "app/Http/Middleware"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := stub("make/middleware.go.stub", map[string]string{"__NAME__": name})
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

	content := stub("make/migration.go.stub", map[string]string{
		"__NAME__":           structName,
		"__MIGRATION_NAME__": migrationName,
		"__TABLE__":          table,
	})
	return os.WriteFile(filepath.Join(dir, filename), []byte(content), 0644)
}

// MakeSeeder generates a seeder file.
func MakeSeeder(name string) error {
	dir := "database/seeders"
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	content := stub("make/seeder.go.stub", map[string]string{"__NAME__": name})
	return os.WriteFile(filepath.Join(dir, name+".go"), []byte(content), 0644)
}

// MakeRequest generates a form request validation file.
func MakeRequest(name string) error {
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

	content := stub("make/request.go.stub", map[string]string{"__NAME__": requestName})
	return os.WriteFile(filepath.Join(dir, requestName+".go"), []byte(content), 0644)
}

// MakeCommand generates a console command file.
func MakeCommand(name string) error {
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

	content := stub("make/command.go.stub", map[string]string{
		"__NAME__":      commandName,
		"__SIGNATURE__": signature,
	})
	return os.WriteFile(filepath.Join(dir, commandName+".go"), []byte(content), 0644)
}

// MakeEvent generates an event file.
func MakeEvent(name string) error {
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

	content := stub("make/event.go.stub", map[string]string{"__NAME__": eventName})
	return os.WriteFile(filepath.Join(dir, eventName+".go"), []byte(content), 0644)
}

// MakeListener generates an event listener file.
func MakeListener(name string) error {
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

	content := stub("make/listener.go.stub", map[string]string{"__NAME__": listenerName})
	return os.WriteFile(filepath.Join(dir, listenerName+".go"), []byte(content), 0644)
}

// MakeJob generates a queue job file.
func MakeJob(name string) error {
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

	content := stub("make/job.go.stub", map[string]string{"__NAME__": jobName})
	return os.WriteFile(filepath.Join(dir, jobName+".go"), []byte(content), 0644)
}

// MakePolicy generates a policy file for authorization.
func MakePolicy(name string) error {
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

	content := stub("make/policy.go.stub", map[string]string{"__NAME__": policyName})
	return os.WriteFile(filepath.Join(dir, policyName+".go"), []byte(content), 0644)
}

func tableFromMigrationName(lower string) string {
	if m := regexp.MustCompile(`^create_(.+?)_table$`).FindStringSubmatch(lower); m != nil {
		return m[1]
	}
	if m := regexp.MustCompile(`^create_(.+)$`).FindStringSubmatch(lower); m != nil {
		return m[1]
	}
	return lower
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
