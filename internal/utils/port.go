package util

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

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
		// kill -9 <pid>
		_ = exec.Command("kill", "-9", line).Run()
	}
}

func FindFreePort(start int) int {
	for p := start; p < start+50; p++ {
		cmd := exec.Command("lsof", "-i", ":"+strconv.Itoa(p))
		if err := cmd.Run(); err != nil {
			return p
		}
	}
	return -1
}
