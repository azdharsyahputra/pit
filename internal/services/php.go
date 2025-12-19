package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	util "pit/internal/utils"
)

type PHPService struct {
	Root    string
	Version string
}

func NewPHPService(root string, version string) *PHPService {
	return &PHPService{
		Root:    root,
		Version: version,
	}
}

func (s *PHPService) Name() string { return "php-fpm" }

func (s *PHPService) basePath() string {
	if s.Version == "" {
		return filepath.Join(s.Root, "php")
	}
	return filepath.Join(s.Root, "php", s.Version)
}

func (s *PHPService) Start() error {
	base := s.basePath()

	util.KillPort(9099)
	util.CleanupPID(filepath.Join(base, "logs/php-fpm.pid"))
	util.PreparePHPDirs(base)

	fpmBin := filepath.Join(base, "sbin/php-fpm")
	conf := filepath.Join(base, "etc/php-fpm.conf")
	ini := filepath.Join(base, "etc/php.ini")

	fmt.Println("Starting PHP-FPM version", s.Version, "...")

	cmd := exec.Command(fpmBin,
		"-p", base,
		"-y", conf,
		"-c", ini,
		"--daemonize",
	)

	cmd.Env = append(os.Environ(),
		"LD_LIBRARY_PATH="+filepath.Join(base, "libs")+":"+os.Getenv("LD_LIBRARY_PATH"),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *PHPService) Stop() error {
	base := s.basePath()
	util.StopPID(filepath.Join(base, "logs/php-fpm.pid"))
	util.KillPort(9099)
	return nil
}

func (s *PHPService) Status() ServiceStatus {
	base := s.basePath()
	pid := util.GetPID(filepath.Join(base, "logs/php-fpm.pid"))
	return ServiceStatus{
		Running: util.IsAlive(pid),
		PID:     pid,
		Port:    9099,
	}
}
