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
		WebPath: "../../../web", // Web folder (index.html, css/, js/, components/, vendor/)
	}

	// Setup HTTP routes

	// M0 - Core Document & Text operations
	http.HandleFunc("/api/text", api.TextHandler(srv))
	http.HandleFunc("/api/stream", api.StreamHandler(srv))
	http.HandleFunc("/api/merge", api.MergeHandler(srv))
	http.HandleFunc("/api/doc", api.DocHandler(srv))

	// M0 - Map operations
	http.HandleFunc("/api/map", api.MapHandler(srv))
	http.HandleFunc("/api/map/keys", api.MapKeysHandler(srv))

	// M0 - List operations
	http.HandleFunc("/api/list/push", api.ListPushHandler(srv))
	http.HandleFunc("/api/list/insert", api.ListInsertHandler(srv))
	http.HandleFunc("/api/list", api.ListGetHandler(srv))
	http.HandleFunc("/api/list/delete", api.ListDeleteHandler(srv))
	http.HandleFunc("/api/list/len", api.ListLenHandler(srv))

	// M0 - Counter operations
	http.HandleFunc("/api/counter", api.CounterHandler(srv))
	http.HandleFunc("/api/counter/increment", api.CounterIncrementHandler(srv))
	http.HandleFunc("/api/counter/get", api.CounterGetHandler(srv))

	// M0 - History operations
	http.HandleFunc("/api/heads", api.HeadsHandler(srv))
	http.HandleFunc("/api/changes", api.ChangesHandler(srv))

	// M1 - Sync operations
	http.HandleFunc("/api/sync", api.SyncHandler(srv))

	// M2 - RichText operations
	http.HandleFunc("/api/richtext/mark", api.RichTextMarkHandler(srv))
	http.HandleFunc("/api/richtext/unmark", api.RichTextUnmarkHandler(srv))
	http.HandleFunc("/api/richtext/marks", api.RichTextMarksHandler(srv))

	// Cursor operations (stable position tracking)
	http.HandleFunc("/api/cursor", api.CursorGetHandler(srv))         // GET/POST to get cursor
	http.HandleFunc("/api/cursor/lookup", api.CursorLookupHandler(srv)) // POST to lookup cursor position

	// Static files
	http.Handle("/web/", api.WebHandler(staticCfg))            // Serves web/css/, web/js/, web/components/
	http.Handle("/vendor/", api.VendorHandler(staticCfg))      // Serves web/vendor/automerge.js
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Serve web/index.html for root path
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "../../../web/index.html")
		} else {
			http.NotFound(w, r)
		}
	})

	// Start server
	log.Printf("[%s] Server starting on http://localhost:%s", userID, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
