package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"pit/internal/services"
)

type ProjectEngine struct {
	BasePath    string
	Name        string
	Config      *ProjectConfig
	ProjectRoot string // path ke public user
	RuntimeRoot string // path ke runtime/{project}
	Services    []services.Service
}

// -----------------------------------------------------------
// INIT
// -----------------------------------------------------------
func NewProjectEngine(base, name string) (*ProjectEngine, error) {

	cfg, err := LoadProjectConfig(base, name)
	if err != nil {
		return nil, err
	}

	projectRoot := filepath.Join(base, "projects", name)
	publicDir := filepath.Join(projectRoot, cfg.Root)

	runtimeRoot := filepath.Join(base, "runtime", name)

	// Dir untuk runtime
	runDir := filepath.Join(runtimeRoot, "run")
	phpDir := filepath.Join(runtimeRoot, "php")
	nginxDir := filepath.Join(runtimeRoot, "nginx")
	logsDir := filepath.Join(runtimeRoot, "logs")

	for _, d := range []string{
		runtimeRoot, runDir, phpDir, nginxDir, logsDir,
	} {
		_ = os.MkdirAll(d, 0o755)
	}

	e := &ProjectEngine{
		BasePath:    base,
		Name:        name,
		Config:      cfg,
		ProjectRoot: publicDir,
		RuntimeRoot: runtimeRoot,
	}

	e.Services = []services.Service{
		services.NewProjectPHPService(base, name, cfg.PHPVersion, cfg.Port+100),
		services.NewProjectNginxService(base, name, cfg.Port),
	}

	return e, nil
}

// -----------------------------------------------------------
// START PROJECT
// -----------------------------------------------------------
func (e *ProjectEngine) Start() error {
	fmt.Println("=== START PROJECT:", e.Name, "===")

	// Safety: kill leftover processes
	e.killLeftovers()

	// Start both services
	for _, svc := range e.Services {
		fmt.Println("Starting:", svc.Name())
		if err := svc.Start(); err != nil {
			return fmt.Errorf("error starting %s: %v", svc.Name(), err)
		}
	}

	return nil
}

// -----------------------------------------------------------
// STOP PROJECT — clean stop
// -----------------------------------------------------------
func (e *ProjectEngine) Stop() error {
	fmt.Println("=== STOP PROJECT:", e.Name, "===")

	// stop each service (php + nginx)
	for _, svc := range e.Services {
		fmt.Println("Stopping:", svc.Name())
		_ = svc.Stop()
	}

	// kill workers & free ports
	e.killLeftovers()

	// cleanup pid files
	os.Remove(filepath.Join(e.RuntimeRoot, "run", "php-fpm.pid"))
	os.Remove(filepath.Join(e.RuntimeRoot, "run", "nginx.pid"))

	// cleanup sockets
	os.Remove(filepath.Join(e.RuntimeRoot, "php", "php-fpm.sock"))
	os.Remove(filepath.Join(e.RuntimeRoot, "nginx", "nginx.sock"))

	return nil
}

// -----------------------------------------------------------
// FORCE STOP (API)
// -----------------------------------------------------------
func (e *ProjectEngine) ForceStopAll() {
	fmt.Println("=== FORCE STOP PROJECT:", e.Name, "===")

	for _, svc := range e.Services {
		_ = svc.Stop()
	}

	e.killLeftovers()
	e.cleanRuntime()
}

// -----------------------------------------------------------
// KILL ALL ZOMBIE PROCESSES (nginx + php-fpm)
// -----------------------------------------------------------
func (e *ProjectEngine) killLeftovers() {

	project := e.Name

	// pattern runtime dir khusus project
	runtimePattern := fmt.Sprintf("%s/runtime/%s", e.BasePath, project)

	// ---- Kill php-fpm workers for this project
	out, _ := exec.Command("bash", "-c",
		fmt.Sprintf("ps aux | grep php-fpm | grep '%s' | awk '{print $2}'", project),
	).Output()

	for _, pid := range strings.Split(string(out), "\n") {
		p := strings.TrimSpace(pid)
		if p != "" {
			exec.Command("kill", "-9", p).Run()
		}
	}

	// ---- Kill nginx master + workers for this project
	out, _ = exec.Command("bash", "-c",
		fmt.Sprintf("ps aux | grep nginx | grep '%s' | awk '{print $2}'", runtimePattern),
	).Output()

	for _, pid := range strings.Split(string(out), "\n") {
		p := strings.TrimSpace(pid)
		if p != "" {
			exec.Command("kill", "-9", p).Run()
		}
	}

	// ---- Kill ports
	exec.Command("bash", "-c",
		fmt.Sprintf("lsof -t -i:%d | xargs -r kill -9", e.Config.Port),
	).Run()

	exec.Command("bash", "-c",
		fmt.Sprintf("lsof -t -i:%d | xargs -r kill -9", e.Config.Port+100),
	).Run()
}

// -----------------------------------------------------------
// CLEAN RUNTIME FOLDER
// -----------------------------------------------------------
func (e *ProjectEngine) cleanRuntime() {
	os.RemoveAll(e.RuntimeRoot)
}

// -----------------------------------------------------------
// STATUS
// -----------------------------------------------------------
func (e *ProjectEngine) Status() map[string]services.ServiceStatus {
	resp := map[string]services.ServiceStatus{}
	for _, svc := range e.Services {
		resp[svc.Name()] = svc.Status()
	}
	return resp
}
