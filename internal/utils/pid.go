package util

import (
	"os"
	"strconv"
	"strings"
	"syscall"
)

func GetPID(pidFile string) int {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0
	}
	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		return 0
	}
	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		return 0
	}
	return pid
}

func IsAlive(pid int) bool {
	if pid == 0 {
		return false
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false
	}
	err = proc.Signal(syscall.Signal(0))
	return err == nil
}

func StopPID(pidFile string) {
	pid := GetPID(pidFile)
	if pid == 0 || !IsAlive(pid) {
		_ = os.Remove(pidFile)
		return
	}
	proc, _ := os.FindProcess(pid)
	_ = proc.Kill()
	_ = os.Remove(pidFile)
}

func CleanupPID(pidFile string) {
	pid := GetPID(pidFile)
	if pid == 0 || !IsAlive(pid) {
		_ = os.Remove(pidFile)
	}
}
