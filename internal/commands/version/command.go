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

	// NUEVO: subcomando status
	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Shows current applied migration (version and file)",
		Run: func(cmd *cobra.Command, args []string) {
			if envPath == "" {
				log.Fatal("❌ You must specify --env")
			}

			cfg, err := config.LoadConfigFromFile(envPath)
			if err != nil {
				log.Fatalf("❌ Could not load configuration: %v", err)
			}

			// Cargamos el log y lo ordenamos en sentido "up" para tomar el último aplicado
			logMap, err := LoadMigrationLog(cfg.LogMigrationPath, "up")
			if err != nil {
				log.Fatalf("❌ Could not read migration log: %v", err)
			}

			history, ok := logMap[cfg.DbName]
			if !ok || len(history.AppliedVersions) == 0 {
				fmt.Println("ℹ️ No migrations applied yet.")
				return
			}

			lastVer := history.AppliedVersions[len(history.AppliedVersions)-1]
			var currentFile string
			if len(lastVer.Files) > 0 {
				currentFile = lastVer.Files[len(lastVer.Files)-1]
			}

			// Solo imprimir versión y archivo actual aplicado (como pediste)
			fmt.Printf("Version: %s\n", lastVer.Version)
			if currentFile != "" {
				fmt.Printf("File: %s\n", currentFile)
			} else {
				fmt.Println("File: (none)")
			}
		},
	}

	// Flags compartidas para ambos comandos
	for _, c := range []*cobra.Command{upCmd, downCmd} {
		c.Flags().StringVar(&envPath, "env", "", "Path to the .env file with credentials")
		c.Flags().BoolVar(&all, "all", false, "Apply/revert all migrations")
		c.Flags().IntVar(&stepsVersion, "steps-version", 0, "Number of versions to apply/revert")
		c.Flags().IntVar(&stepsFile, "steps-file", 0, "Number of files to apply/revert")
	}
	// Flag solo para status
	statusCmd.Flags().StringVar(&envPath, "env", "", "Path to the .env file with credentials")

	migrateCmd.AddCommand(upCmd, downCmd)
	return migrateCmd
}
