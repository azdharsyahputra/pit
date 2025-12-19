package services

import (
	"os"
	"os/exec"
	"path/filepath"
)

type ToolsPHPService struct {
	BasePath string
	PHPBin   string
}

func NewToolsPHPService(base string, phpBin string) *ToolsPHPService {
	return &ToolsPHPService{
		BasePath: base,
		PHPBin:   phpBin,
	}
}

func (s *ToolsPHPService) Name() string {
	return "php-fpm-tools"
}

func (s *ToolsPHPService) runtimeDir() string {
	return filepath.Join(s.BasePath, "runtime", "_tools", "php")
}

func (s *ToolsPHPService) socketPath() string {
	return filepath.Join(s.runtimeDir(), "php-fpm.sock")
}

func (s *ToolsPHPService) Start() error {
	rt := s.runtimeDir()
	_ = os.MkdirAll(filepath.Join(rt, "logs"), 0755)

	conf := filepath.Join(rt, "php-fpm.conf")

	cmd := exec.Command(
		s.PHPBin,
		"--fpm-config", conf,
		"--nodaemonize",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}

func (s *ToolsPHPService) Stop() error {
	// MVP stop: hapus socket (PHP-FPM exit otomatis)
	_ = os.Remove(s.socketPath())
	return nil
}

func (s *ToolsPHPService) Status() ServiceStatus {
	if _, err := os.Stat(s.socketPath()); err == nil {
		return ServiceStatus{
			Running: true,
			PID:     0, // PHP-FPM tools belum tracking PID (OK untuk sekarang)
			Port:    0, // socket-based, bukan TCP
		}
	}

	return ServiceStatus{
		Running: false,
		PID:     0,
		Port:    0,
	}
}
