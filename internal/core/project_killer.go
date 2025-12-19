package core

import (
	"os"
	"path/filepath"
	"strconv"
)

func KillAllProjectRuntimes(base string) {
	runtimeDir := filepath.Join(base, "runtime")

	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return
	}

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		projectName := e.Name()
		projectRun := filepath.Join(runtimeDir, projectName, "run")

		// kill php
		killPid(filepath.Join(projectRun, "php-fpm.pid"))

		// kill nginx
		killPid(filepath.Join(projectRun, "nginx.pid"))

		// cleanup optional
	}
}

func killPid(pidFile string) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return
	}

	pid, _ := strconv.Atoi(string(data))
	proc, _ := os.FindProcess(pid)
	_ = proc.Kill()

	os.Remove(pidFile)
}
