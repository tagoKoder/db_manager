package version

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func runSQLFileWithTransaction(dbDsn, path, fileName string) bool {
	seedFile := fmt.Sprintf("%s\\%s", path, fileName)
	log.Printf("🔍 Reading SQL file: %s\n", seedFile)

	queryBytes, err := os.ReadFile(seedFile)
	if err != nil {
		log.Fatalf("❌ Error reading SQL file %s: %v", seedFile, err)
		return false
	}

	query := string(queryBytes)
	containsTransaction := strings.Contains(query, "BEGIN;") || strings.Contains(query, "COMMIT;")

	db, err := sql.Open("postgres", dbDsn)
	if err != nil {
		log.Printf("❌ Error connecting to the database: %v", err)
		return false
	}
	defer db.Close()

	if containsTransaction {
		_, err = db.Exec(query)
	} else {
		tx, err := db.Begin()
		if err != nil {
			log.Fatalf("❌ Error starting transaction: %v", err)
			return false
		}

		statements := splitSQLStatements(query)
		for i, stmt := range statements {
			stmt = strings.TrimSpace(stmt)
			if stmt == "" {
				continue
			}

			_, err = tx.Exec(stmt)
			if err != nil {
				tx.Rollback()
				log.Printf("❌ Error in %s, statement #%d:\n--- SQL ---\n%s\n--- ERROR ---\n%v", fileName, i+1, stmt, err)
				return false
			}
		}

		err = tx.Commit()
		if err != nil {
			log.Printf("❌ Error committing transaction in %s: %v", fileName, err)
			return false
		}
	}

	if err != nil {
		log.Printf("❌ Error executing SQL in %s: %v", fileName, err)
		return false
	}

	log.Printf("✅ %s executed successfully\n", fileName)
	return true
}

func splitSQLStatements(query string) []string {
	var statements []string
	var sb strings.Builder
	inSingleQuote := false
	inDoubleQuote := false

	for i := 0; i < len(query); i++ {
		c := query[i]

		// Check if we are inside quotes
		if c == '\'' && !inDoubleQuote {
			// If it's a single quote and not escaped
			if i == 0 || query[i-1] != '\\' {
				inSingleQuote = !inSingleQuote
			}
		} else if c == '"' && !inSingleQuote {
			// If it's a double quote and not escaped
			if i == 0 || query[i-1] != '\\' {
				inDoubleQuote = !inDoubleQuote
			}
		}

		// If we find a `;` outside of quotes, we split
		if c == ';' && !inSingleQuote && !inDoubleQuote {
			statements = append(statements, strings.TrimSpace(sb.String()))
			sb.Reset()
		} else {
			sb.WriteByte(c)
		}
	}

	// Add what's left in the buffer
	if s := strings.TrimSpace(sb.String()); s != "" {
		statements = append(statements, s)
	}

	return statements
}

func GetOrderedVersionFoldersAndFiles(baseDir, action string) ([]string, map[string][]string, error) {
	resultDirs := []string{}
	resultFiles := make(map[string][]string)

	// Ruta base del tipo ./db/commerce_core/up o down
	dirPath := filepath.Join(baseDir, action)

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, nil, fmt.Errorf("❌ could not read folder %s: %w", dirPath, err)
	}

	// Recolectar solo carpetas (versiones)
	for _, entry := range entries {
		if entry.IsDir() {
			resultDirs = append(resultDirs, entry.Name())
		}
	}

	// Ordenar versiones
	sort.Slice(resultDirs, func(i, j int) bool {
		if action == "down" {
			return resultDirs[i] > resultDirs[j]
		}
		return resultDirs[i] < resultDirs[j]
	})

	// Leer y ordenar archivos por cada versión
	for _, version := range resultDirs {
		versionPath := filepath.Join(dirPath, version)
		filesInVersion := []string{}

		files, err := os.ReadDir(versionPath)
		if err != nil {
			return nil, nil, fmt.Errorf("❌ error reading files in %s: %w", versionPath, err)
		}

		for _, f := range files {
			if !f.IsDir() {
				filesInVersion = append(filesInVersion, f.Name())
			}
		}

		sort.Slice(filesInVersion, func(i, j int) bool {
			if action == "down" {
				return filesInVersion[i] > filesInVersion[j]
			}
			return filesInVersion[i] < filesInVersion[j]
		})

		resultFiles[version] = filesInVersion
	}

	return resultDirs, resultFiles, nil
}
