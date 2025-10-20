package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func main() {
	ctx := context.Background()

	// Configuration from environment
	storageDir := getEnv("STORAGE_DIR", "../../../")
	port := getEnv("PORT", "8080")
	userID := getEnv("USER_ID", "default")

	// Create and initialize server
	srv := server.New(server.Config{
		StorageDir: storageDir,
		UserID:     userID,
	})

	if err := srv.Initialize(ctx); err != nil {
		log.Fatalf("Failed to initialize document: %v", err)
	}

	// Static file configuration
	staticCfg := api.StaticConfig{
		UIPath:     "../../../ui/ui.html",
		VendorPath: "../../../ui/vendor",
	}

	// Setup HTTP routes
	http.HandleFunc("/api/text", api.TextHandler(srv))
	http.HandleFunc("/api/stream", api.StreamHandler(srv))
	http.HandleFunc("/api/merge", api.MergeHandler(srv))
	http.HandleFunc("/api/doc", api.DocHandler(srv))
	http.Handle("/vendor/", api.VendorHandler(staticCfg))
	http.HandleFunc("/", api.UIHandler(staticCfg))

	// Start server
	log.Printf("[%s] Server starting on http://localhost:%s", userID, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
