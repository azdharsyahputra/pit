package core

import (
	"fmt"
	"os"
	"path/filepath"
)

type ProjectRegistry struct {
	BasePath string
}

func NewProjectRegistry(base string) *ProjectRegistry {
	return &ProjectRegistry{BasePath: base}
}

// -----------------------
// Helper Paths
// -----------------------

func (r *ProjectRegistry) projectPath(name string) string {
	return filepath.Join(r.BasePath, "projects", name)
}

// -----------------------
// CREATE PROJECT
// -----------------------

func (r *ProjectRegistry) Create(name string) error {
	root := r.projectPath(name)
	cfgDir := filepath.Join(root, ".pit")

	if err := os.MkdirAll(filepath.Join(root, "public"), 0755); err != nil {
		return err
	}
	if err := os.MkdirAll(cfgDir, 0755); err != nil {
		return err
	}

	cfg := &ProjectConfig{
		Name:       name,
		PHPVersion: "83",
		Port:       10000,
		Root:       "public",
	}

	return r.SaveConfig(name, cfg)
}

// -----------------------
// LOAD CONFIG
// -----------------------

func (r *ProjectRegistry) LoadConfig(name string) (*ProjectConfig, error) {
	return LoadProjectConfig(r.BasePath, name)
}

// -----------------------
// SAVE CONFIG
// -----------------------

func (r *ProjectRegistry) SaveConfig(name string, cfg *ProjectConfig) error {
	return SaveProjectConfig(r.BasePath, name, cfg)
}

// -----------------------
// UPDATE CONFIG
// -----------------------

func (r *ProjectRegistry) Update(name string, cfg *ProjectConfig) error {
	return r.SaveConfig(name, cfg)
}

// -----------------------
// LIST PROJECTS
// -----------------------

func (r *ProjectRegistry) List() ([]string, error) {
	dir := filepath.Join(r.BasePath, "projects")

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	list := []string{}
	for _, e := range entries {
		if e.IsDir() {
			list = append(list, e.Name())
		}
	}
	return list, nil
}

// -----------------------
// LOAD ENGINE
// -----------------------

func (r *ProjectRegistry) Load(name string) (*ProjectEngine, error) {
	if _, err := os.Stat(r.projectPath(name)); err != nil {
		return nil, fmt.Errorf("project not found: %s", name)
	}
	return NewProjectEngine(r.BasePath, name)
}

// -----------------------
// READ CONFIG (API)
// -----------------------

func (r *ProjectRegistry) ReadConfig(name string) (*ProjectConfig, error) {
	return r.LoadConfig(name)
}
