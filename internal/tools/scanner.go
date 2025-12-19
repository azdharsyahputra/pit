package tools

import (
	"os"
	"path/filepath"
)

func Scan(base string) ([]Manifest, error) {
	toolsDir := filepath.Join(base, "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return nil, err
	}

	var out []Manifest
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		manifestPath := filepath.Join(toolsDir, e.Name(), "tool.json")
		if _, err := os.Stat(manifestPath); err != nil {
			continue
		}
		m, err := LoadManifest(manifestPath)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, nil
}
