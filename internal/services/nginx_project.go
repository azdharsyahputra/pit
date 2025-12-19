package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

type ProjectNginxService struct {
	BasePath string
	Project  string
	Port     int
}

func NewProjectNginxService(base, project string, port int) *ProjectNginxService {
	return &ProjectNginxService{
		BasePath: base,
		Project:  project,
		Port:     port,
	}
}

func (s *ProjectNginxService) Name() string {
	return "nginx-project:" + s.Project
}

func (s *ProjectNginxService) Start() error {
	runtimeRoot := filepath.Join(s.BasePath, "runtime", s.Project)
	nginxRuntime := filepath.Join(runtimeRoot, "nginx")
	runDir := filepath.Join(runtimeRoot, "run")
	logDir := filepath.Join(nginxRuntime, "logs")

	projectPublic := filepath.Join(s.BasePath, "projects", s.Project, "public")

	_ = os.MkdirAll(nginxRuntime, 0o755)
	_ = os.MkdirAll(runDir, 0o755)
	_ = os.MkdirAll(logDir, 0o755)

	confFile := filepath.Join(nginxRuntime, "nginx.conf")
	pidFile := filepath.Join(runDir, "nginx.pid")

	nginxBin := filepath.Join(s.BasePath, "nginx/sbin/nginx")
	sockPath := filepath.Join(runtimeRoot, "php", "php-fpm.sock")

	conf := fmt.Sprintf(`
	worker_processes 1;

	events {
		worker_connections 1024;
	}

	http {
		include %s;
		default_type application/octet-stream;

		access_log %s/access.log;
		error_log %s/error.log;

		server {
			listen %d;
			server_name %s.local;

			root %s;
			index index.php index.html;

			location / {
				try_files $uri $uri/ /index.php?$query_string;
			}

			location ~ \.php$ {
				fastcgi_pass unix:%s;
				include %s;
				fastcgi_param SCRIPT_FILENAME %s$fastcgi_script_name;
			}
		}
	}
	`, filepath.Join(s.BasePath, "nginx/conf/mime.types"),
		logDir, logDir,
		s.Port,
		s.Project,
		projectPublic,
		sockPath,
		filepath.Join(s.BasePath, "nginx/conf/fastcgi.conf"),
		projectPublic,
	)

	_ = os.WriteFile(confFile, []byte(conf), 0644)

	cmd := exec.Command(nginxBin,
		"-p", nginxRuntime,
		"-c", confFile,
		"-g", fmt.Sprintf("pid %s;", pidFile),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func (s *ProjectNginxService) Stop() error {
	pidFile := filepath.Join(s.BasePath, "runtime", s.Project, "run/nginx.pid")

	data, err := os.ReadFile(pidFile)
	if err == nil {
		if pid, err := strconv.Atoi(string(data)); err == nil {
			proc, _ := os.FindProcess(pid)
			_ = proc.Kill()
		}
		_ = os.Remove(pidFile)
	}

	killCmd := fmt.Sprintf(
		"ps aux | grep nginx | grep '%s/runtime/%s' | awk '{print $2}' | xargs -r kill -9",
		s.BasePath, s.Project,
	)
	exec.Command("bash", "-c", killCmd).Run()

	return nil
}

func (s *ProjectNginxService) Status() ServiceStatus {
	pidFile := filepath.Join(s.BasePath, "runtime", s.Project, "run/nginx.pid")

	data, err := os.ReadFile(pidFile)
	if err != nil {
		return ServiceStatus{Running: false, Port: s.Port}
	}

	pid, _ := strconv.Atoi(string(data))

	return ServiceStatus{
		Running: true,
		PID:     pid,
		Port:    s.Port,
	}
}
