// Package config provides configuration for Automerge WASI server.
//
// This package defines all environment variables and configuration options
// for embedding the Automerge WASI server into your own applications.
package config

import (
	"os"
	"strconv"
)

// Config holds all configuration for the Automerge WASI HTTP server.
//
// Usage in your own main.go:
//
//	cfg := config.NewFromEnv()
//	srv, err := httpserver.New(cfg)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	log.Fatal(srv.ListenAndServe())
type Config struct {
	// Port to listen on (default: "8080")
	// Env: PORT
	Port string

	// StorageDir is where .am snapshot files are saved
	// (default: current directory)
	// Env: STORAGE_DIR
	StorageDir string

	// UserID identifies this server instance (for logging/debugging)
	// (default: "default")
	// Env: USER_ID
	UserID string

	// WASMPath is the path to automerge_wasi.wasm file
	// (default: "../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm")
	// Env: WASM_PATH
	//
	// If using embed.FS, set this to "" and provide WASMBytes instead.
	WASMPath string

	// WASMBytes is the embedded WASM binary (optional)
	// If provided, WASMPath is ignored.
	// Use with go:embed in your main.go:
	//
	//	//go:embed automerge_wasi.wasm
	//	var wasmBytes []byte
	//	cfg.WASMBytes = wasmBytes
	WASMBytes []byte

	// WebPath is the path to the web/ folder containing UI files
	// (default: "../web")
	// Env: WEB_PATH
	//
	// Structure expected:
	//   web/index.html
	//   web/css/
	//   web/js/
	//   web/components/
	//   web/vendor/automerge.js
	WebPath string

	// EnableUI enables the web UI routes (/, /web/*, /vendor/*)
	// (default: true)
	// Env: ENABLE_UI (set to "false" to disable)
	EnableUI bool
}

// NewFromEnv creates a Config from environment variables with sensible defaults.
//
// Environment Variables:
//   - PORT: HTTP port (default: "8080")
//   - STORAGE_DIR: Directory for .am snapshots (default: ".")
//   - USER_ID: Server instance identifier (default: "default")
//   - WASM_PATH: Path to .wasm file (default: relative path for dev)
//   - WEB_PATH: Path to web UI folder (default: "../web")
//   - ENABLE_UI: Enable web UI (default: "true")
//
// Example:
//
//	PORT=3000 STORAGE_DIR=/data go run main.go
func NewFromEnv() Config {
	return Config{
		Port:       getEnv("PORT", "8080"),
		StorageDir: getEnv("STORAGE_DIR", "."),
		UserID:     getEnv("USER_ID", "default"),
		WASMPath:   getEnv("WASM_PATH", "../rust/automerge_wasi/target/wasm32-wasip1/release/automerge_wasi.wasm"),
		WebPath:    getEnv("WEB_PATH", "../web"),
		EnableUI:   getEnvBool("ENABLE_UI", true),
	}
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool returns environment variable as bool or default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		parsed, err := strconv.ParseBool(value)
		if err == nil {
			return parsed
		}
	}
	return defaultValue
}
