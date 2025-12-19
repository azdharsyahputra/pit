package config

import (
	"encoding/json"
	"os"
)

type EngineConfig struct {
	PHPVersion string `json:"php_version"`
}

func DefaultConfig() EngineConfig {
	return EngineConfig{
		PHPVersion: "83",
	}
}

func Load(path string) EngineConfig {
	data, err := os.ReadFile(path)
	if err != nil {
		return DefaultConfig()
	}

	var cfg EngineConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return DefaultConfig()
	}

	if cfg.PHPVersion == "" {
		cfg.PHPVersion = DefaultConfig().PHPVersion
	}

	return cfg
}

func Save(path string, cfg EngineConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, data, 0o644)
}

func dir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' {
			return path[:i]
		}
	}
	return "."
}
