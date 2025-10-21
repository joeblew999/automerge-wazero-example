package main

import (
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/config"
	"github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
)

func main() {
	// Load configuration from environment variables
	cfg := config.NewFromEnv()

	// Create and start HTTP server
	srv, err := httpserver.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// Start server (blocking)
	log.Fatal(srv.ListenAndServe())
}
