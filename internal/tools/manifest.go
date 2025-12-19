package tools

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Manifest struct {
	Name   string `json:"name"`
	Domain string `json:"domain"`
	Type   string `json:"type"`  // "php" (MVP)
	Root   string `json:"root"`  // relative to base
	Index  string `json:"index"` // "index.php"
}

func LoadManifest(path string) (Manifest, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, err
	}
	var m Manifest
	if err := json.Unmarshal(b, &m); err != nil {
		return Manifest{}, err
	}
	// Normalize root to clean relative path
	m.Root = filepath.Clean(m.Root)
	return m, nil
}
