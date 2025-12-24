package core

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"pit/internal/config"
	"pit/internal/services"
	util "pit/internal/utils"
)

type Engine struct {
	BasePath string
	Config   config.EngineConfig
	Services []services.Service
}

// Constructor utama engine
func NewEngine(base string) *Engine {
	cfgPath := filepath.Join(base, "config", "engine.json")
	cfg := config.Load(cfgPath)

	www := filepath.Join(base, "www")

	e := &Engine{
		BasePath: base,
		Config:   cfg,
	}

	e.ensureToolsPHPConfig()

	e.Services = []services.Service{
		services.NewPHPService(base, cfg.PHPVersion),  // project PHP
		services.NewToolsPHPService(base, e.PHPBin()), // 👈 TOOLS PHP
		services.NewNginxService(base, www),
	}

	return e
}

// Menyimpan config (php_version dll)
func (e *Engine) saveConfig() error {
	cfgPath := filepath.Join(e.BasePath, "config", "engine.json")

	if err := os.MkdirAll(filepath.Dir(cfgPath), 0o755); err != nil {
		return err
	}

	return config.Save(cfgPath, e.Config)
}

// ---------- PROJECT RUNTIME CLEANUP ----------
// remove unix socket safely
func removeSocket(path string) {
	if _, err := os.Stat(path); err == nil {
		_ = os.Remove(path)
	}
}
func (e *Engine) cleanupProjectRuntimes() {
	runtimeDir := filepath.Join(e.BasePath, "runtime")

	entries, err := os.ReadDir(runtimeDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		project := entry.Name()
		projectRuntime := filepath.Join(runtimeDir, project)

		runDir := filepath.Join(projectRuntime, "run")
		phpDir := filepath.Join(projectRuntime, "php")

		phpPid := filepath.Join(runDir, "php-fpm.pid")
		nginxPid := filepath.Join(runDir, "nginx.pid")
		phpSock := filepath.Join(phpDir, "php-fpm.sock")

		// Kill by PID
		killPidFile(phpPid)
		killPidFile(nginxPid)

		// 🔥 IMPORTANT: remove stale socket
		removeSocket(phpSock)

		// Extra safety: kill by port (jika config ada)
		cfg, err := LoadProjectConfig(e.BasePath, project)
		if err == nil {
			killPort(cfg.Port)       // nginx
			killPort(cfg.Port + 100) // php-fpm (tcp fallback)
		}
	}
}

// ---------- ENGINE START / STOP ----------
func (e *Engine) StartAll() error {
	fmt.Println("=== pit START ===")

	pidFile := filepath.Join(e.BasePath, "runtime", "pit.pid")

	// 1️⃣ ALWAYS CLEAN FIRST (treat start like restart)
	fmt.Println("[Init] Cleaning global runtimes...")
	for i := len(e.Services) - 1; i >= 0; i-- {
		_ = e.Services[i].Stop()
	}

	// small wait to avoid race (php-fpm / nginx exit)
	time.Sleep(300 * time.Millisecond)

	// safety-net: kill PIT-owned php-fpm only (NOT OS php-fpm)
	_ = exec.Command("pkill", "-f", e.ToolsPHPRuntime()).Run()
	_ = exec.Command("pkill", "-f", filepath.Join(
		e.BasePath,
		"php",
		e.Config.PHPVersion,
		"etc",
		"php-fpm.conf",
	)).Run()

	// cleanup runtime markers
	_ = os.Remove(pidFile)
	_ = os.Remove(e.ToolsPHPSocket())

	// 2️⃣ WRITE PIT PID (AFTER CLEAN)
	if err := os.MkdirAll(filepath.Dir(pidFile), 0o755); err != nil {
		return fmt.Errorf("failed to create runtime dir: %w", err)
	}
	if err := WritePID(pidFile, os.Getpid()); err != nil {
		return fmt.Errorf("failed to write pit pid: %w", err)
	}

	// ============================================
	// AUTO HOSTS GENERATOR
	// ============================================
	wwwDir := filepath.Join(e.BasePath, "www")
	entries, _ := os.ReadDir(wwwDir)

	var domains []string
	for _, entry := range entries {
		if entry.IsDir() {
			domains = append(domains, entry.Name()+".test")
		}
	}

	if len(domains) > 0 {
		fmt.Println("[Hosts] Syncing domains...")
		if err := util.EnsureHosts(domains); err != nil {
			fmt.Println("[Hosts] Failed to update /etc/hosts:", err)
		}
	}
	// ============================================

	// 3️⃣ START GLOBAL SERVICES (FRESH)
	for _, s := range e.Services {
		fmt.Println("Starting:", s.Name())
		if err := s.Start(); err != nil {
			return err
		}
	}

	fmt.Println("pit running at http://localhost:8080")
	return nil
}

func (e *Engine) StopAll() error {
	fmt.Println("=== pit STOP ===")

	// 1️⃣ Stop global services (reverse order)
	for i := len(e.Services) - 1; i >= 0; i-- {
		s := e.Services[i]
		fmt.Println("Stopping:", s.Name())
		_ = s.Stop()
	}

	// 2️⃣ WAIT for processes to exit (IMPORTANT)
	time.Sleep(500 * time.Millisecond)

	// 3️⃣ VERIFY php-fpm benar-benar mati (safety net)
	_ = exec.Command("pkill", "-f", e.BasePath+"/pit/runtime/_tools/php/php-fpm.conf").Run()
	_ = exec.Command("pkill", "-f", e.BasePath+"/pit/php/"+e.Config.PHPVersion+"/etc/php-fpm.conf").Run()

	// 4️⃣ Cleanup sockets (karena php-fpm unix socket)
	_ = os.Remove(e.ToolsPHPSocket())
	_ = os.Remove(filepath.Join(
		e.BasePath,
		"php",
		e.Config.PHPVersion,
		"var",
		"run",
		"php-fpm.sock",
	))

	// 5️⃣ REMOVE PID FILE (NO SELF KILL)
	mainPIDFile := filepath.Join(e.BasePath, "runtime", "pit.pid")
	_ = os.Remove(mainPIDFile)

	fmt.Println("pit stopped cleanly.")
	return nil
}

func (e *Engine) ReloadNginx() error {
	for _, s := range e.Services {
		if s.Name() == "nginx" {
			if r, ok := s.(interface {
				Reload() error
			}); ok {
				return r.Reload()
			}
			return fmt.Errorf("nginx service does not support reload")
		}
	}
	return fmt.Errorf("nginx service not found")
}

// ---------- STATUS ----------

// Return status tiap service (running / stopped)
func (e *Engine) ServiceStatuses() map[string]services.ServiceStatus {
	statuses := make(map[string]services.ServiceStatus)

	for _, s := range e.Services {
		statuses[s.Name()] = s.Status()
	}

	return statuses
}

// ---------- HELPERS (PID & PORT) ----------
func killPidFile(pidFile string) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return
	}

	pidStr := strings.TrimSpace(string(data))
	if pidStr == "" {
		_ = os.Remove(pidFile)
		return
	}

	pid, err := strconv.Atoi(pidStr)
	if err != nil {
		_ = os.Remove(pidFile)
		return
	}

	proc, err := os.FindProcess(pid)
	if err == nil {
		// GRACEFUL FIRST
		_ = proc.Signal(syscall.SIGTERM)
	}

	_ = os.Remove(pidFile)
}

// kill proses berdasarkan port (fallback paling brutal)
func killPort(port int) {
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
		_ = exec.Command("kill", "-9", line).Run()
	}
}

// ========== PUBLIC: FORCE CLEANUP (dipakai oleh main.go/API) ==========

// Membersihkan semua runtime project secara brutal (PID + PORT)
func (e *Engine) ForceKillAllProjectRuntimes() {
	fmt.Println("[ForceKill] Cleaning all project runtimes ...")

	// 1) cleanup normal
	e.cleanupProjectRuntimes()

	// 2) scan semua project config
	projectsDir := filepath.Join(e.BasePath, "projects")

	entries, err := os.ReadDir(projectsDir)
	if err != nil {
		return
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		project := entry.Name()

		cfg, err := LoadProjectConfig(e.BasePath, project)
		if err != nil {
			continue
		}

		// Kill runtime ports
		killPort(cfg.Port)
		killPort(cfg.Port + 100)

		// Runtime paths
		projectRuntime := filepath.Join(e.BasePath, "runtime", project)
		runDir := filepath.Join(projectRuntime, "run")
		phpDir := filepath.Join(projectRuntime, "php")

		// Kill PID files
		killPidFile(filepath.Join(runDir, "php-fpm.pid"))
		killPidFile(filepath.Join(runDir, "nginx.pid"))

		// 🔥 REMOVE SOCKET (ROOT CAUSE BUG)
		removeSocket(filepath.Join(phpDir, "php-fpm.sock"))
	}
}
func (e *Engine) ToolsPHPRuntime() string {
	return filepath.Join(e.BasePath, "runtime", "_tools", "php")
}

func (e *Engine) ToolsPHPSocket() string {
	return filepath.Join(e.ToolsPHPRuntime(), "php-fpm.sock")
}
func (e *Engine) ensureToolsPHPConfig() error {
	rt := e.ToolsPHPRuntime()
	_ = os.MkdirAll(filepath.Join(rt, "logs"), 0755)

	conf := filepath.Join(rt, "php-fpm.conf")
	if _, err := os.Stat(conf); err == nil {
		return nil
	}

	content := `
[global]
error_log = ` + rt + `/logs/error.log
daemonize = no

[www]
listen = ` + e.ToolsPHPSocket() + `
listen.mode = 0660
pm = dynamic
pm.max_children = 5
pm.start_servers = 1
pm.min_spare_servers = 1
pm.max_spare_servers = 3
`

	return os.WriteFile(conf, []byte(content), 0644)
}
func (e *Engine) PHPBin() string {
	// sesuaikan dengan cara lu resolve php-fpm sekarang
	// contoh umum:
	return filepath.Join(
		e.BasePath,
		"php",
		e.Config.PHPVersion,
		"sbin",
		"php-fpm",
	)
}
