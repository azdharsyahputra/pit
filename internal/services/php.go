package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

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

	pidFile := filepath.Join(base, "var", "run", "php-fpm.pid")
	sockFile := filepath.Join(base, "var", "run", "php-fpm.sock")

	util.PreparePHPDirs(base)

	// === PREVENT DOUBLE START ===
	if data, err := os.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if util.IsAlive(pid) {
				fmt.Println("PHP-FPM already running, skip start")
				return nil
			}
		}
		_ = os.Remove(pidFile)
	}

	// remove stale socket
	_ = os.Remove(sockFile)

	fpmBin := filepath.Join(base, "sbin/php-fpm")
	conf := filepath.Join(base, "etc/php-fpm.conf")
	ini := filepath.Join(base, "etc/php.ini")

	fmt.Println("Starting PHP-FPM version", s.Version, "...")

	cmd := exec.Command(
		fpmBin,
		"-p", base,
		"-y", conf,
		"-c", ini,
		"--nodaemonize",
	)

	cmd.Env = append(os.Environ(),
		"LD_LIBRARY_PATH="+filepath.Join(base, "libs")+":"+os.Getenv("LD_LIBRARY_PATH"),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
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
