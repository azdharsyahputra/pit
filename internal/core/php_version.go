package core

import (
	"os"
	"path/filepath"
)

// ListPHPVersions scans php/* directory for portable PHP versions
func (e *Engine) ListPHPVersions() ([]string, error) {
	base := filepath.Join(e.BasePath, "php")

	entries, err := os.ReadDir(base)
	if err != nil {
		return nil, err
	}

	var versions []string

	for _, entry := range entries {
		if entry.IsDir() {
			// cek apakah folder tersebut valid PHP portable
			fpm := filepath.Join(base, entry.Name(), "sbin/php-fpm")
			if _, err := os.Stat(fpm); err == nil {
				versions = append(versions, entry.Name())
			}
		}
	}

	return versions, nil
}

// CurrentPHPVersion returns active php version
func (e *Engine) CurrentPHPVersion() string {
	return e.Config.PHPVersion
}
