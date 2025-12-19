package api

import (
	"encoding/json"
	"net/http"

	"pit/internal/core"
)

func RegisterRoutes(mux *http.ServeMux, engine *core.Engine) {

	// ================================
	// SERVICE STATUS
	// ================================
	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, engine.ServiceStatuses())
	})

	// ================================
	// START ALL SERVICES
	// ================================
	mux.HandleFunc("/start", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_ = engine.StartAll()
		writeJSON(w, map[string]string{"status": "ok"})
	})

	// ================================
	// STOP ALL SERVICES
	// ================================
	mux.HandleFunc("/stop", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		_ = engine.StopAll()
		writeJSON(w, map[string]string{"status": "ok"})
	})

	// ================================
	// GET CURRENT PHP VERSION
	// ================================
	mux.HandleFunc("/php/current", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, map[string]string{
			"current": engine.CurrentPHPVersion(),
		})
	})

	// ================================
	// LIST ALL PHP VERSIONS
	// ================================
	mux.HandleFunc("/php/versions", func(w http.ResponseWriter, r *http.Request) {
		versions, err := engine.ListPHPVersions()
		if err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}
		writeJSON(w, versions)
	})

	// ================================
	// SWITCH PHP VERSION
	// ================================
	mux.HandleFunc("/php/use", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		// 1. Try: URL Query (?version=82)
		version := r.URL.Query().Get("version")

		// 2. If empty → try parse JSON
		if version == "" {
			var body struct {
				Version string `json:"version"`
			}
			_ = json.NewDecoder(r.Body).Decode(&body)
			version = body.Version
		}

		if version == "" {
			writeJSON(w, map[string]string{"error": "missing version"})
			return
		}

		if err := engine.SetPHPVersion(version); err != nil {
			writeJSON(w, map[string]string{"error": err.Error()})
			return
		}

		writeJSON(w, map[string]string{
			"status":  "ok",
			"version": version,
		})
	})
}

func writeJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(data)
}
