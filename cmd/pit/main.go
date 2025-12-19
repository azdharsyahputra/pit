package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"pit/internal/api"
	"pit/internal/core"
	"pit/internal/tools"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	base, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	engine := core.NewEngine(base)

	switch os.Args[1] {

	// ----------------------------
	// ONE-TIME SETUP
	// ----------------------------
	case "setup":
		if err := engine.SetupTrust(); err != nil {
			fmt.Println("Setup failed:", err)
			os.Exit(1)
		}
		fmt.Println("✔ Setup completed")
		os.Exit(0)

	// ----------------------------
	// START ENGINE
	// ----------------------------
	case "start":
		results := engine.PreflightChecks()
		if !printChecks(results) {
			fmt.Println("\nCannot start: preflight checks failed.")
			os.Exit(1)
		}

		fmt.Println("\n[Init] Cleaning project runtimes...")
		engine.ForceKillAllProjectRuntimes()

		fmt.Printf("Starting services.\n")
		if err := engine.StartAll(); err != nil {
			fmt.Println("Error starting services:", err)
			os.Exit(1)
		}

		go func() {
			fmt.Println("✔ API started on http://localhost:7070")
			api.StartAPIServer(engine)
		}()

		fmt.Println("\nEngine is ready.")
		select {}

	// ----------------------------
	// STOP ENGINE
	// ----------------------------
	case "stop":
		if err := engine.StopAll(); err != nil {
			fmt.Println("Error stopping engine:", err)
		} else {
			fmt.Println("✔ Engine stopped")
		}
		os.Exit(0)

	// ----------------------------
	// API ONLY
	// ----------------------------
	case "api":
		engine.ForceKillAllProjectRuntimes()
		api.StartAPIServer(engine)

	// ----------------------------
	// SUB COMMANDS
	// ----------------------------
	case "php":
		handlePHPCommand(engine)

	case "project":
		handleProjectCommand(engine)
	case "tools":
		handleToolsCommand(engine)

	default:
		fmt.Println("Unknown command:", os.Args[1])
		printUsage()
	}
}

func printChecks(results []core.CheckResult) bool {
	ok := true
	for _, r := range results {
		if r.OK {
			fmt.Printf("[Check] %-18s OK\n", r.Name)
		} else {
			ok = false
			fmt.Printf("[Check] %-18s FAIL\n", r.Name)
			fmt.Println("        Reason:", r.Reason)
			if r.Fix != "" {
				fmt.Println("        Fix   :", r.Fix)
			}
		}
	}
	return ok
}

////////////////////////////////////////////////////////
// PHP SUBCOMMANDS
////////////////////////////////////////////////////////

func handlePHPCommand(engine *core.Engine) {
	if len(os.Args) < 3 {
		printPHPUsage()
		return
	}

	switch os.Args[2] {

	case "use":
		if len(os.Args) < 4 {
			fmt.Println("Missing version.")
			printPHPUsage()
			return
		}
		ver := os.Args[3]

		engine.ForceKillAllProjectRuntimes()

		if err := engine.SetPHPVersion(ver); err != nil {
			fmt.Println("Error setting PHP version:", err)
			return
		}

		fmt.Println("PHP version switched to", ver)

	case "versions":
		versions, err := engine.ListPHPVersions()
		if err != nil {
			fmt.Println("Error listing versions:", err)
			return
		}
		fmt.Println("Available PHP versions:", versions)

	case "current":
		fmt.Println("Current PHP version:", engine.CurrentPHPVersion())

	default:
		fmt.Println("Unknown php command:", os.Args[2])
		printPHPUsage()
	}
}

////////////////////////////////////////////////////////
// PROJECT SUBCOMMANDS (FINAL)
////////////////////////////////////////////////////////

func handleProjectCommand(engine *core.Engine) {
	if len(os.Args) < 3 {
		printProjectUsage()
		return
	}

	reg := core.NewProjectRegistry(engine.BasePath)

	switch os.Args[2] {

	case "list":
		projects, err := reg.List()
		if err != nil {
			fmt.Println("Error:", err)
			return
		}
		for _, p := range projects {
			fmt.Println("-", p)
		}

	case "info":
		if len(os.Args) < 4 {
			fmt.Println("Missing project name.")
			return
		}
		name := os.Args[3]

		cfg, err := reg.LoadConfig(name)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		fmt.Println("Project:", cfg.Name)
		fmt.Println("PHP Version:", cfg.PHPVersion)
		fmt.Println("Port:", cfg.Port)
		fmt.Println("Root:", cfg.Root)

	case "set-port":
		if len(os.Args) < 5 {
			fmt.Println("Usage: pit project set-port <name> <port>")
			return
		}

		name := os.Args[3]
		portStr := os.Args[4]

		port, err := strconv.Atoi(portStr)
		if err != nil {
			fmt.Println("Invalid port:", portStr)
			return
		}

		cfg, err := reg.LoadConfig(name)
		if err != nil {
			fmt.Println("Error loading config:", err)
			return
		}

		cfg.Port = port

		if err := reg.SaveConfig(name, cfg); err != nil {
			fmt.Println("Error saving config:", err)
			return
		}

		fmt.Println("Restarting project", name)

		peng, err := reg.Load(name)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		_ = peng.Stop()
		if err := peng.Start(); err != nil {
			fmt.Println("Failed to restart:", err)
			return
		}

		fmt.Println("Project", name, "updated to port", port)

	case "restart":
		if len(os.Args) < 4 {
			fmt.Println("Missing project name.")
			return
		}
		name := os.Args[3]

		peng, err := reg.Load(name)
		if err != nil {
			fmt.Println("Error:", err)
			return
		}

		_ = peng.Stop()
		if err := peng.Start(); err != nil {
			fmt.Println("Restart failed:", err)
			return
		}

		fmt.Println("Project restarted:", name)

	default:
		fmt.Println("Unknown project command:", os.Args[2])
		printProjectUsage()
	}
}

////////////////////////////////////////////////////////
// TOOLS SUBCOMMANDS (MVP)
////////////////////////////////////////////////////////

func handleToolsCommand(engine *core.Engine) {
	if len(os.Args) < 3 {
		printToolsUsage()
		return
	}

	switch os.Args[2] {

	case "sync":
		fmt.Println("[Tools] Scanning tools...")

		mgr := tools.Manager{
			Base:       engine.BasePath,
			PhpSockAbs: engine.ToolsPHPSocket(),

			NginxReload: func() error {
				return engine.ReloadNginx()
			},
		}

		if err := mgr.SyncAll(); err != nil {
			fmt.Println("Tools sync failed:", err)
			return
		}

		fmt.Println("✔ Tools synced successfully")

	default:
		fmt.Println("Unknown tools command:", os.Args[2])
		printToolsUsage()
	}
}
func printToolsUsage() {
	fmt.Println("Tools Commands:")
	fmt.Println("  pit tools sync")
}

////////////////////////////////////////////////////////
// HELPERS
////////////////////////////////////////////////////////

func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  pit setup")
	fmt.Println("  pit start")
	fmt.Println("  pit stop")
	fmt.Println("  pit api")
	fmt.Println("  pit php use <version>")
	fmt.Println("  pit php versions")
	fmt.Println("  pit php current")
	fmt.Println("  pit project list")
	fmt.Println("  pit project info <name>")
	fmt.Println("  pit project set-port <name> <port>")
	fmt.Println("  pit project restart <name>")
	fmt.Println("  pit tools sync")
}

func printPHPUsage() {
	fmt.Println("PHP Commands:")
	fmt.Println("  pit php use <version>")
	fmt.Println("  pit php versions")
	fmt.Println("  pit php current")
}

func printProjectUsage() {
	fmt.Println("Project Commands:")
	fmt.Println("  pit project list")
	fmt.Println("  pit project info <name>")
	fmt.Println("  pit project set-port <name> <port>")
	fmt.Println("  pit project restart <name>")
}
