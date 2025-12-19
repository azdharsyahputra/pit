package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func WritePID(pidFile string, pid int) error {
	if err := os.MkdirAll(filepath.Dir(pidFile), 0o755); err != nil {
		return err
	}
	return os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", pid)), 0644)
}

func ReadPID(pidFile string) (int, error) {
	raw, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil {
		return 0, err
	}
	return pid, nil
}

func KillPID(pidFile string) {
	pid, err := ReadPID(pidFile)
	if err != nil {
		return
	}

	proc, _ := os.FindProcess(pid)
	proc.Kill()

	_ = os.Remove(pidFile)
}

// kill all processes on a port (brutal fallback)
func KillPort(port int) {
	out, err := exec.Command("lsof", "-t", "-i", fmt.Sprintf(":%d", port)).Output()
	if err != nil {
		return
	}

	lines := strings.Split(strings.TrimSpace(string(out)), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		exec.Command("kill", "-9", line).Run()
	}
}
