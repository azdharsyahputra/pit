package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

func hasCap(bin string) bool {
	out, err := exec.Command("getcap", bin).Output()
	if err != nil {
		return false
	}
	return string(out) != ""
}
func (e *Engine) SetupTrust() error {
	nginxBin := filepath.Join(e.BasePath, "nginx", "sbin", "nginx")

	if _, err := os.Stat(nginxBin); err != nil {
		return fmt.Errorf("nginx binary not found")
	}

	// ----------------------------
	// NGINX CAPABILITY
	// ----------------------------
	if hasCap(nginxBin) {
		fmt.Println("[Trust] Nginx already trusted")
	} else {
		fmt.Println("[Trust] Granting nginx permission for privileged ports (80/443)")
		fmt.Println("[Trust] This is a one-time setup")

		cmd := exec.Command(
			"sudo",
			"setcap",
			"cap_net_bind_service=+ep",
			nginxBin,
		)

		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		if err := cmd.Run(); err != nil {
			return err
		}
	}

	// ----------------------------
	// MYSQL ROOT AUTH
	// ----------------------------
	plugin, err := mysqlRootAuthPlugin()
	if err != nil {
		fmt.Println("[Trust] MySQL not detected, skipping DB setup")
		return nil
	}

	if plugin == "auth_socket\n" {
		fmt.Println("[Trust] MySQL root uses auth_socket")
		fmt.Println("[Trust] Switching to passwordless mode (local dev)")
		return setupMySQLRootPasswordless()
	}

	fmt.Println("[Trust] MySQL root already password-based")
	return nil
}

func mysqlRootAuthPlugin() (string, error) {
	cmd := exec.Command(
		"sudo",
		"mysql",
		"-Nse",
		"SELECT plugin FROM mysql.user WHERE user='root' AND host='localhost';",
	)

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	return string(out), nil
}
func setupMySQLRootPasswordless() error {
	fmt.Println("[Trust] Configuring MySQL root for local development")

	sql := `
ALTER USER 'root'@'localhost'
IDENTIFIED WITH mysql_native_password
BY '';
FLUSH PRIVILEGES;
`

	cmd := exec.Command(
		"sudo",
		"mysql",
		"-e",
		sql,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
