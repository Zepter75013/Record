package backup

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// settingsFile stores the user-chosen backup directory so it survives
// server restarts, independently of where the backups themselves live.
//
// It must sit inside "backups/" (the default backup directory, mounted as
// a Docker volume in docker-compose.yml) rather than at the app root:
// the app root lives in the container's own filesystem layer, which is
// recreated from scratch on every image rebuild — a file written there
// "survives restarts" but is silently wiped on every deploy, which is
// exactly the durability this file exists to provide.
const settingsFile = "backups/backup-settings.json"

type persistedSettings struct {
	Directory string `json:"directory"`
}

func loadPersistedDirectory() (string, bool) {
	data, err := os.ReadFile(settingsFile)
	if err != nil {
		return "", false
	}

	var s persistedSettings
	if err := json.Unmarshal(data, &s); err != nil || s.Directory == "" {
		return "", false
	}

	return s.Directory, true
}

func savePersistedDirectory(dir string) error {
	if err := os.MkdirAll(filepath.Dir(settingsFile), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(persistedSettings{Directory: dir}, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(settingsFile, data, 0o644)
}
