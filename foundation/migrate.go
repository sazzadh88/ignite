package foundation

import (
	"fmt"

	"github.com/sazzadh88/ignite/database"
	"github.com/sazzadh88/ignite/schema"
)

// runMigrate executes a migrate* command against the default database
// connection. Migrations register themselves via schema.RegisterMigration
// (from the project's blank-imported database/migrations package).
func (app *Application) runMigrate(command string) error {
	dbCfg, ok := app.config.Get("database").(map[string]any)
	if !ok {
		return fmt.Errorf("database configuration not found")
	}

	defaultName, _ := dbCfg["default"].(string)
	conns, _ := dbCfg["connections"].(map[string]any)
	active, _ := conns[defaultName].(map[string]any)
	if active == nil {
		return fmt.Errorf("database connection %q not configured", defaultName)
	}
	driver, _ := active["driver"].(string)

	mgr := database.NewManager(dbCfg)
	conn, err := mgr.Default()
	if err != nil {
		return fmt.Errorf("connect (%s): %w", defaultName, err)
	}
	defer mgr.Close()

	sch := schema.NewSchemaWithDriver(conn, driver)
	migrator := schema.NewMigrator(sch)
	migrator.LoadRegistered()

	switch command {
	case "migrate":
		if err := migrator.Migrate(); err != nil {
			return err
		}
		fmt.Println("  Migrations complete.")

	case "migrate:rollback":
		if err := migrator.Rollback(); err != nil {
			return err
		}
		fmt.Println("  Rolled back last batch.")

	case "migrate:reset":
		if err := migrator.Reset(); err != nil {
			return err
		}
		fmt.Println("  All migrations rolled back.")

	case "migrate:refresh":
		if err := migrator.Reset(); err != nil {
			return err
		}
		if err := migrator.Migrate(); err != nil {
			return err
		}
		fmt.Println("  Database refreshed.")

	case "migrate:fresh":
		if err := migrator.Fresh(); err != nil {
			return err
		}
		fmt.Println("  Database wiped and re-migrated.")

	case "migrate:status":
		statuses, err := migrator.Status()
		if err != nil {
			return err
		}
		if len(statuses) == 0 {
			fmt.Println("  No migrations found.")
			return nil
		}
		fmt.Printf("\n  %-45s %-8s %s\n", "MIGRATION", "BATCH", "STATUS")
		fmt.Printf("  %s\n", "-------------------------------------------------------------")
		for _, s := range statuses {
			batch := "-"
			if s.Status == "ran" {
				batch = fmt.Sprintf("%d", s.Batch)
			}
			fmt.Printf("  %-45s %-8s %s\n", s.Name, batch, s.Status)
		}
		fmt.Println()
	}

	return nil
}
