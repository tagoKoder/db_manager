package version

import (
	"fmt"
	"log"

	"github.com/spf13/cobra"
	"github.com/tagoKoder/db_manager/internal/config"
)

func NewMigrateCommand() *cobra.Command {
	var envPath string
	var all bool
	var stepsVersion, stepsFile int

	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Database migration management",
	}

	upCmd := &cobra.Command{
		Use:   "up",
		Short: "Executes all migrations (or up to a certain point)",
		Run: func(cmd *cobra.Command, args []string) {
			if envPath == "" {
				log.Fatal("❌ You must specify --env")
			}

			cfg, err := config.LoadConfigFromFile(envPath)
			if err != nil {
				log.Fatalf("❌ Could not load configuration: %v", err)
			}

			baseDir := cfg.MigrationsPath
			logPath := cfg.LogMigrationPath

			err = ExecuteMigrations(cfg.DbDSN, baseDir, logPath, cfg.DbName, "up", all, stepsVersion, stepsFile)
			if err != nil {
				log.Fatalf("❌ Error applying migrations: %v", err)
			}
			fmt.Println("✅ Migrations applied successfully.")
		},
	}

	downCmd := &cobra.Command{
		Use:   "down",
		Short: "Reverts migrations by steps or all",
		Run: func(cmd *cobra.Command, args []string) {
			if envPath == "" {
				log.Fatal("❌ You must specify --env")
			}

			cfg, err := config.LoadConfigFromFile(envPath)
			if err != nil {
				log.Fatalf("❌ Could not load configuration: %v", err)
			}

			baseDir := cfg.MigrationsPath
			logPath := cfg.LogMigrationPath

			err = ExecuteMigrations(cfg.DbDSN, baseDir, logPath, cfg.DbName, "down", all, stepsVersion, stepsFile)
			if err != nil {
				log.Fatalf("❌ Error reverting migrations: %v", err)
			}
			fmt.Println("⏪ Migrations reverted successfully.")
		},
	}

	// Flags compartidas para ambos comandos
	for _, c := range []*cobra.Command{upCmd, downCmd} {
		c.Flags().StringVar(&envPath, "env", "", "Path to the .env file with credentials")
		c.Flags().BoolVar(&all, "all", false, "Apply/revert all migrations")
		c.Flags().IntVar(&stepsVersion, "steps-version", 0, "Number of versions to apply/revert")
		c.Flags().IntVar(&stepsFile, "steps-file", 0, "Number of files to apply/revert")
	}

	migrateCmd.AddCommand(upCmd, downCmd)
	return migrateCmd
}
