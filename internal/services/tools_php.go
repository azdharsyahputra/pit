package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	util "pit/internal/utils"
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

func (s *ToolsPHPService) pidFile() string {
	return filepath.Join(s.runtimeDir(), "php-fpm.pid")
}

func (s *ToolsPHPService) confPath() string {
	return filepath.Join(s.runtimeDir(), "php-fpm.conf")
}

func (s *ToolsPHPService) Start() error {
	rt := s.runtimeDir()
	_ = os.MkdirAll(filepath.Join(rt, "logs"), 0755)

	pidFile := s.pidFile()
	sock := s.socketPath()

	// 🛑 PREVENT DOUBLE START
	if data, err := os.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if util.IsAlive(pid) {
				fmt.Println("php-fpm-tools already running, skip")
				return nil
			}
		}
		_ = os.Remove(pidFile)
	}

	// cleanup stale socket
	_ = os.Remove(sock)

	fmt.Println("Starting Tools PHP-FPM...")

	cmd := exec.Command(
		s.PHPBin,
		"--fpm-config", s.confPath(),
		"--pid", pidFile,
		"--nodaemonize",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Start()
}

func (s *ToolsPHPService) Stop() error {
	pidFile := s.pidFile()
	sock := s.socketPath()

	// graceful stop via PID
	if data, err := os.ReadFile(pidFile); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil && pid > 1 {
			if util.IsAlive(pid) {
				_ = syscall.Kill(pid, syscall.SIGTERM)

				// wait max ~1s
				for i := 0; i < 20; i++ {
					if !util.IsAlive(pid) {
						break
					}
					time.Sleep(50 * time.Millisecond)
				}
			}
		}
		_ = os.Remove(pidFile)
	}

	// safety: remove socket
	_ = os.Remove(sock)

	return nil
}

func (s *ToolsPHPService) Status() ServiceStatus {
	if data, err := os.ReadFile(s.pidFile()); err == nil {
		if pid, err := strconv.Atoi(strings.TrimSpace(string(data))); err == nil {
			if util.IsAlive(pid) {
				return ServiceStatus{
					Running: true,
					PID:     pid,
					Port:    0,
				}
			}
		}
	}

	return ServiceStatus{
		Running: false,
		PID:     0,
		Port:    0,
	}
}
