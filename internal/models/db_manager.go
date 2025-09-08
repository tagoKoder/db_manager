package models

type AppliedFile struct {
	Version   string   `json:"version"`
	AppliedAt string   `json:"applied_at"`
	Files     []string `json:"files"`
}

type DBHistory struct {
	AppliedVersions []AppliedFile `json:"applied_versions"`
}

type MigrationLog map[string]DBHistory // key: service name (e.g. "commerce_core")
