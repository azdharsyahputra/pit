package util

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func EnsureHosts(domains []string) error {
	data, err := os.ReadFile("/etc/hosts")
	if err != nil {
		return err
	}

	content := string(data)

	for _, domain := range domains {
		entry := "127.0.0.1 " + domain

		if !strings.Contains(content, entry) {
			fmt.Println("[Hosts] Adding:", entry)
			cmd := exec.Command("sudo", "sh", "-c",
				fmt.Sprintf("echo '%s' >> /etc/hosts", entry),
			)
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("failed to write hosts entry: %w", err)
			}
		}
	}

	return nil
}
