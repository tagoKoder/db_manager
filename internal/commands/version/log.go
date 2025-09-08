package version

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/tagoKoder/db_manager/internal/models"
)

func LoadMigrationLog(logPath string, action string) (models.MigrationLog, error) {
	logMap := make(models.MigrationLog)

	// If the file doesn't exist, return an empty log
	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		return logMap, nil
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		return nil, err
	}

	// Verify if the file is empty
	if len(data) == 0 {
		return logMap, nil
	}

	if err := json.Unmarshal(data, &logMap); err != nil {
		return nil, err
	}

	// Sort versions and files according to direction
	for dbKey, history := range logMap {
		applied := history.AppliedVersions

		// Sort versions
		sort.Slice(applied, func(i, j int) bool {
			if action == "down" {
				return applied[i].Version > applied[j].Version
			}
			return applied[i].Version < applied[j].Version
		})

		// Sort files within each version
		for idx := range applied {
			files := applied[idx].Files
			sort.Slice(files, func(i, j int) bool {
				if action == "down" {
					return files[i] > files[j]
				}
				return files[i] < files[j]
			})
			applied[idx].Files = files
		}

		logMap[dbKey] = models.DBHistory{AppliedVersions: applied}
	}

	return logMap, nil
}

// OverwriteMigrationLog overwrites the log file with the content of logData.
// If the file doesn't exist, it creates it. If it's empty or invalid, it also rewrites it.
func OverwriteMigrationLog(logPath string, logData models.MigrationLog) error {
	// Check if the file exists
	_, err := os.Stat(logPath)
	if os.IsNotExist(err) {
		fmt.Printf("📁 The log file does not exist. It will be created: %s\n", logPath)
	} else if err != nil {
		return fmt.Errorf("❌ error checking log: %w", err)
	}

	// Serialize logData to JSON
	data, err := json.MarshalIndent(logData, "", "  ")
	if err != nil {
		return fmt.Errorf("❌ error serializing log: %w", err)
	}

	// Write the content to the file (create or overwrite)
	err = os.WriteFile(logPath, data, 0644)
	if err != nil {
		return fmt.Errorf("❌ error writing log file: %w", err)
	}

	fmt.Printf("✅ Log overwritten successfully in '%s'\n", logPath)
	return nil
}
