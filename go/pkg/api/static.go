package api

import (
	"net/http"
	"path/filepath"
)

// StaticConfig holds configuration for static file serving
type StaticConfig struct {
	WebPath string // Path to web folder (contains index.html, css/, js/, components/, vendor/)
}

// WebHandler serves static files from the web directory
// Handles requests like /web/css/main.css, /web/js/app.js, /web/components/text.html
func WebHandler(cfg StaticConfig) http.Handler {
	// Resolve absolute path for security
	absPath, err := filepath.Abs(cfg.WebPath)
	if err != nil {
		// Fallback to relative path if absolute resolution fails
		absPath = cfg.WebPath
	}

	fileServer := http.FileServer(http.Dir(absPath))
	return http.StripPrefix("/web/", fileServer)
}

// VendorHandler serves static files from the vendor directory (web/vendor/)
// Handles requests like /vendor/automerge.js
func VendorHandler(cfg StaticConfig) http.Handler {
	// Vendor is inside web folder
	vendorPath := filepath.Join(cfg.WebPath, "vendor")

	// Resolve absolute path for security
	absPath, err := filepath.Abs(vendorPath)
	if err != nil {
		// Fallback to relative path if absolute resolution fails
		absPath = vendorPath
	}

	fileServer := http.FileServer(http.Dir(absPath))
	return http.StripPrefix("/vendor/", fileServer)
}
