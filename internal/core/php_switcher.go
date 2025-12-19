package core

import (
	"fmt"
	"os"
	"path/filepath"

	"pit/internal/services"
)

// SetPHPVersion switches PHP version by:
// 1. Validasi folder ada
// 2. Stop PHP-FPM lama
// 3. Ubah config
// 4. Save config
// 5. Rebuild PHP service
// 6. Start versi baru
func (e *Engine) SetPHPVersion(ver string) error {
	verPath := filepath.Join(e.BasePath, "php", ver)
	info, err := os.Stat(verPath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("php version directory not found: %s", verPath)
	}

	// Stop old php-fpm
	for _, s := range e.Services {
		if s.Name() == "php-fpm" {
			_ = s.Stop()
		}
	}

	// Save new version
	e.Config.PHPVersion = ver
	if err := e.saveConfig(); err != nil {
		return err
	}

	// Replace php-fpm service
	var newServices []services.Service
	for _, s := range e.Services {
		if s.Name() == "php-fpm" {
			newServices = append(newServices, services.NewPHPService(e.BasePath, ver))
		} else {
			newServices = append(newServices, s)
		}
	}
	e.Services = newServices

	// Start new PHP-FPM
	for _, s := range e.Services {
		if s.Name() == "php-fpm" {
			return s.Start()
		}
	}

	return fmt.Errorf("php-fpm service not found")
}
