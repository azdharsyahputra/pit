package core

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type ProjectConfig struct {
	Name       string `json:"name"`
	PHPVersion string `json:"php_version"`
	Port       int    `json:"port"`
	Root       string `json:"root"`
}

func (cfg *ProjectConfig) Save(base string) error {
	path := filepath.Join(base, "projects", cfg.Name, ".pit", "config.json")

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
