package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/tagoKoder/db_manager/internal/commands/version"
	// futuros comandos:
	// "github.com/tagoKoder/db_manager/internal/schema"
	// "github.com/tagoKoder/db_manager/internal/migration_generator"
)

func main() {
	rootCmd := &cobra.Command{
		Use:   "command",
		Short: "CLI service to manage databases, migrations, and more.",
	}

	// 🔹 Add commands organized by functionality
	rootCmd.AddCommand(
		version.NewMigrateCommand(), // ✅ Migrations
		// seeder.NewMigrationGenerator,     // 🚧 future command to insert migrations
	)

	if err := rootCmd.Execute(); err != nil {
		log.Println("❌ Error executing the CLI:", err)
		os.Exit(1)
	}
}
