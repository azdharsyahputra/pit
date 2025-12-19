package services

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	util "pit/internal/utils"
)

type NginxService struct {
	Base    string
	WWWRoot string
}

func NewNginxService(root string, www string) *NginxService {
	return &NginxService{
		Base:    filepath.Join(root, "nginx"),
		WWWRoot: www,
	}
}

func (s *NginxService) Name() string { return "nginx" }

func (s *NginxService) Start() error {
	// Bersihkan port dan PID lama
	util.KillPort(80)
	util.CleanupPID(filepath.Join(s.Base, "logs/nginx.pid"))
	util.PrepareNginxDirs(s.Base)

	// Generate nginx.conf dinamis
	confFile := filepath.Join(s.Base, "conf/nginx.conf")
	if err := s.generateConfig(confFile); err != nil {
		return err
	}

	nginxBin := filepath.Join(s.Base, "sbin/nginx")

	fmt.Println("Starting Nginx global (www mode)...")

	cmd := exec.Command(nginxBin,
		"-p", s.Base,
		"-c", confFile,
		"-g", "daemon off;",
	)

	// Fix lib dependency untuk portable build
	cmd.Env = append(os.Environ(),
		"LD_LIBRARY_PATH="+filepath.Join(s.Base, "libs")+":"+os.Getenv("LD_LIBRARY_PATH"),
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (s *NginxService) Stop() error {
	pidFile := filepath.Join(s.Base, "logs/nginx.pid")

	pid := util.GetPID(pidFile)
	if pid <= 0 {
		return nil
	}

	// SIGTERM = graceful shutdown
	proc, err := os.FindProcess(pid)
	if err != nil {
		return err
	}

	_ = proc.Signal(syscall.SIGTERM)

	// optional: wait & cleanup
	time.Sleep(500 * time.Millisecond)
	util.CleanupPID(pidFile)

	return nil
}

func (s *NginxService) Status() ServiceStatus {
	pid := util.GetPID(filepath.Join(s.Base, "logs/nginx.pid"))
	return ServiceStatus{
		Running: util.IsAlive(pid),
		PID:     pid,
		Port:    80,
	}
}

// -------------------------------------------------
// CONFIG GENERATOR
// -------------------------------------------------
func (s *NginxService) generateConfig(outPath string) error {
	mimeTypes := filepath.Join(s.Base, "conf", "mime.types")
	fastcgiConf := filepath.Join(s.Base, "conf", "fastcgi.conf")
	logDir := filepath.Join(s.Base, "logs")

	toolsInclude := filepath.Join(s.Base, "conf.d", "tools", "*.conf")

	serverBlocks, err := s.buildServerBlocks(fastcgiConf)
	if err != nil {
		return err
	}

	// fallback kalau www kosong
	if serverBlocks == "" {
		serverBlocks = fmt.Sprintf(`
server {
    listen 80 default_server;
    server_name localhost;
    root %s;
    index index.php index.html;

    location / {
        try_files $uri $uri/ =404;
    }
}
`, filepath.Join(s.Base, "html"))
	}

	conf := fmt.Sprintf(`
worker_processes  1;

events {
    worker_connections  1024;
}

http {
    include       %s;
    default_type  application/octet-stream;

    access_log  %s/access.log;
    error_log   %s/error.log;

    # pit tools
    include %s;

%s
}
`, mimeTypes, logDir, logDir, toolsInclude, serverBlocks)

	return os.WriteFile(outPath, []byte(conf), 0644)
}

// -------------------------------------------------
// VHOST GENERATOR (PER-FOLDER DI /WWW)
// -------------------------------------------------

func (s *NginxService) buildServerBlocks(fastcgiConf string) (string, error) {
	entries, err := os.ReadDir(s.WWWRoot)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}

	var servers []string
	first := true

	for _, e := range entries {
		if !e.IsDir() {
			continue
		}

		name := e.Name()
		siteRoot := filepath.Join(s.WWWRoot, name)

		// detect public folder (Laravel / CI4)
		publicRoot := filepath.Join(siteRoot, "public")
		if st, err := os.Stat(publicRoot); err == nil && st.IsDir() {
			siteRoot = publicRoot
		}

		// first server becomes default server
		listen := "80"
		if first {
			listen = "80 default_server"
			first = false
		}

		serverBlock := fmt.Sprintf(`
server {
    listen %s;
    server_name %s.test;

    root %s;
    index index.php index.html;

    location / {
        try_files $uri $uri/ /index.php?$query_string;
    }

    location ~ \.php$ {
        include %s;
        fastcgi_pass 127.0.0.1:9099;
        fastcgi_param SCRIPT_FILENAME $document_root$fastcgi_script_name;
    }
}
`, listen, name, siteRoot, fastcgiConf)

		servers = append(servers, serverBlock)
	}

	return strings.Join(servers, "\n"), nil
}
func (s *NginxService) Reload() error {
	cmd := exec.Command(
		filepath.Join(s.Base, "sbin", "nginx"),
		"-p", s.Base,
		"-c", filepath.Join(s.Base, "conf", "nginx.conf"),
		"-s", "reload",
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("[DEBUG] Reload nginx with -p -c from pit")
	return cmd.Run()
}
