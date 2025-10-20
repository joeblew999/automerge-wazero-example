package api

import (
	"net/http"
	"path/filepath"
)

// StaticConfig holds configuration for static file serving
type StaticConfig struct {
	UIPath     string
	VendorPath string
}

// UIHandler serves the main UI HTML file
func UIHandler(cfg StaticConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, cfg.UIPath)
	}
}

// VendorHandler serves static files from the vendor directory
// Handles requests like /vendor/automerge.js
func VendorHandler(cfg StaticConfig) http.Handler {
	// Resolve absolute path for security
	absPath, err := filepath.Abs(cfg.VendorPath)
	if err != nil {
		// Fallback to relative path if absolute resolution fails
		absPath = cfg.VendorPath
	}

	fileServer := http.FileServer(http.Dir(absPath))
	return http.StripPrefix("/vendor/", fileServer)
}
