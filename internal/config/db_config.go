package config

import (
	"fmt"

	"github.com/joho/godotenv"
)

type DBConfig struct {
	DbDSN            string
	DbName           string
	LogMigrationPath string
	MigrationsPath   string
}

// Loads a specific .env file
func LoadConfigFromFile(envPath string) (*DBConfig, error) {
	if err := godotenv.Load(envPath); err != nil {
		return nil, fmt.Errorf("error loading .env file (%s): %w", envPath, err)
	}

	cfg := &DBConfig{
		DbDSN:            getEnv("DB_DSN", ""),
		DbName:           getEnv("DB_NAME", ""),
		LogMigrationPath: getEnv("LOG_MIGRATION_PATH", "./schema/migration_log.json"),
		MigrationsPath:   getEnv("MIGRATIONS_PATH", "./migrations"),
	}

	return cfg, nil
}
