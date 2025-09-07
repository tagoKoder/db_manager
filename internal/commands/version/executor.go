package version

import (
	"fmt"
	"log"
	"path/filepath"
	"time"

	"github.com/tagoKoder/db_manager/internal/models"
)

func ExecuteMigrations(
	dbDsn string,
	baseDir string,
	logPath string,
	dbName string,
	action string, // "up" o "down"
	all bool,
	stepsVersion int,
	stepsFile int,
) error {
	if action != "up" && action != "down" {
		return fmt.Errorf("❌ Invalid action: %s. Must be 'up' or 'down'", action)
	}

	logData, err := LoadMigrationLog(logPath, action)
	if err != nil {
		return fmt.Errorf("error loading migration log: %w", err)
	}

	canConnect, err := CanConnectDSN(dbDsn)
	if err != nil {
		return fmt.Errorf("error checking database connection: %w", err)
	}
	if !canConnect {
		return fmt.Errorf("could not connect to the database with the provided DSN")
	}

	applied := logData[dbName].AppliedVersions
	filesRemaining := stepsFile
	versionsRemaining := stepsVersion

	versions, versionFiles, err := GetOrderedVersionFoldersAndFiles(baseDir, action)
	if err != nil {
		return err
	}

	if action == "up" {
		for _, version := range versions {
			files := versionFiles[version]
			var appliedFiles []string
			found := false
			for _, v := range applied {
				if v.Version == version {
					appliedFiles = v.Files
					found = true
					break
				}
			}

			// compare files to see which are pending, not only count
			filesMap := make(map[string]bool)
			for _, f := range appliedFiles {
				filesMap[f] = true
			}
			toExecute := []string{}
			for _, f := range files {
				if !filesMap[f] {
					toExecute = append(toExecute, f)
				}
			}

			if len(toExecute) == 0 {
				continue // version completely applied
			}

			executed := append([]string{}, appliedFiles...)
			for _, file := range toExecute {
				log.Printf("📂 Executing %s from version %s", file, version)
				if ok := runSQLFileWithTransaction(dbDsn, filepath.Join(baseDir, action, version), file); !ok {
					return fmt.Errorf("❌ Error executing file %s in version %s", file, version)
				}
				executed = append(executed, file)
			}

			if found {
				for i := range applied {
					if applied[i].Version == version {
						applied[i].Files = executed
						break
					}
				}
			} else {
				applied = append(applied, models.AppliedFile{
					Version:   version,
					AppliedAt: time.Now().UTC().Format(time.RFC3339),
					Files:     executed,
				})
			}

			logData[dbName] = models.DBHistory{AppliedVersions: applied}
			OverwriteMigrationLog(logPath, logData)

			versionsRemaining--
			if !all && stepsVersion > 0 && versionsRemaining == 0 {
				break
			}
		}
	} else if action == "down" {
		for _, version := range versions {
			files, ok := versionFiles[version]
			if !ok {
				continue
			}
			// Buscar si esta versión fue aplicada
			appliedIdx := -1
			for i, v := range applied {
				if v.Version == version {
					appliedIdx = i
					break
				}
			}
			if appliedIdx == -1 {
				continue // esta versión no fue aplicada
			}

			appliedFiles := applied[appliedIdx].Files
			appliedMap := make(map[string]bool)
			for _, f := range appliedFiles {
				appliedMap[f] = true
			}

			// Revertir solo archivos que realmente fueron aplicados
			for _, file := range files {
				if !appliedMap[file] {
					continue // no fue aplicado, se salta
				}
				log.Printf("⏪ Reverting %s from version %s", file, version)
				if ok := runSQLFileWithTransaction(dbDsn, filepath.Join(baseDir, action, version), file); !ok {
					return fmt.Errorf("❌ Error reverting file %s from version %s", file, version)
				}
				filesRemaining--

				// remove file from applied files
				newFiles := []string{}
				for _, f := range applied[appliedIdx].Files {
					if f != file {
						newFiles = append(newFiles, f)
					}
				}
				applied[appliedIdx].Files = newFiles
			}

			// si ya no queda ningún archivo aplicado, eliminar del log
			if len(applied[appliedIdx].Files) == 0 {
				applied = append(applied[:appliedIdx], applied[appliedIdx+1:]...)
			}

			logData[dbName] = models.DBHistory{AppliedVersions: applied}
			OverwriteMigrationLog(logPath, logData)

			if !all && filesRemaining <= 0 {
				break
			}
		}
	} else {
		return fmt.Errorf("unknown action: %s", action)
	}

	return nil
}
