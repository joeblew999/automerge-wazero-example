// Package httpserver provides a reusable HTTP server for Automerge WASI.
//
// This package allows you to embed the Automerge WASI server into your own
// Go applications with minimal configuration.
//
// Example usage:
//
//	import (
//	    "log"
//	    "github.com/joeblew999/automerge-wazero-example/pkg/config"
//	    "github.com/joeblew999/automerge-wazero-example/pkg/httpserver"
//	)
//
//	func main() {
//	    cfg := config.NewFromEnv()
//	    srv, err := httpserver.New(cfg)
//	    if err != nil {
//	        log.Fatal(err)
//	    }
//	    log.Fatal(srv.ListenAndServe())
//	}
package httpserver

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/joeblew999/automerge-wazero-example/pkg/api"
	"github.com/joeblew999/automerge-wazero-example/pkg/config"
	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// HTTPServer wraps the Automerge server with HTTP routes.
type HTTPServer struct {
	cfg    config.Config
	server *server.Server
	mux    *http.ServeMux
}

// New creates a new HTTP server with the given configuration.
//
// This initializes the Automerge document and sets up all HTTP routes.
// Returns an error if initialization fails (e.g., WASM file not found).
func New(cfg config.Config) (*HTTPServer, error) {
	ctx := context.Background()

	// Create and initialize Automerge server
	srv := server.New(server.Config{
		StorageDir: cfg.StorageDir,
		UserID:     cfg.UserID,
	})

	if err := srv.Initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize document: %w", err)
	}

	// Create HTTP server
	h := &HTTPServer{
		cfg:    cfg,
		server: srv,
		mux:    http.NewServeMux(),
	}

	// Setup routes
	h.setupRoutes()

	return h, nil
}

// setupRoutes configures all HTTP routes
func (h *HTTPServer) setupRoutes() {
	// M0 - Core Document & Text operations
	h.mux.HandleFunc("/api/text", api.TextHandler(h.server))
	h.mux.HandleFunc("/api/stream", api.StreamHandler(h.server))
	h.mux.HandleFunc("/api/merge", api.MergeHandler(h.server))
	h.mux.HandleFunc("/api/doc", api.DocHandler(h.server))

	// M0 - Map operations
	h.mux.HandleFunc("/api/map", api.MapHandler(h.server))
	h.mux.HandleFunc("/api/map/keys", api.MapKeysHandler(h.server))

	// M0 - List operations
	h.mux.HandleFunc("/api/list/push", api.ListPushHandler(h.server))
	h.mux.HandleFunc("/api/list/insert", api.ListInsertHandler(h.server))
	h.mux.HandleFunc("/api/list", api.ListGetHandler(h.server))
	h.mux.HandleFunc("/api/list/delete", api.ListDeleteHandler(h.server))
	h.mux.HandleFunc("/api/list/len", api.ListLenHandler(h.server))

	// M0 - Counter operations
	h.mux.HandleFunc("/api/counter", api.CounterHandler(h.server))
	h.mux.HandleFunc("/api/counter/increment", api.CounterIncrementHandler(h.server))
	h.mux.HandleFunc("/api/counter/get", api.CounterGetHandler(h.server))

	// M0 - History operations
	h.mux.HandleFunc("/api/heads", api.HeadsHandler(h.server))
	h.mux.HandleFunc("/api/changes", api.ChangesHandler(h.server))

	// M1 - Sync operations
	h.mux.HandleFunc("/api/sync", api.SyncHandler(h.server))

	// M2 - RichText operations
	h.mux.HandleFunc("/api/richtext/mark", api.RichTextMarkHandler(h.server))
	h.mux.HandleFunc("/api/richtext/unmark", api.RichTextUnmarkHandler(h.server))
	h.mux.HandleFunc("/api/richtext/marks", api.RichTextMarksHandler(h.server))

	// Cursor operations (stable position tracking)
	h.mux.HandleFunc("/api/cursor", api.CursorGetHandler(h.server))
	h.mux.HandleFunc("/api/cursor/lookup", api.CursorLookupHandler(h.server))

	// Static files (if UI enabled)
	if h.cfg.EnableUI {
		staticCfg := api.StaticConfig{
			WebPath: h.cfg.WebPath,
		}

		h.mux.Handle("/web/", api.WebHandler(staticCfg))        // Serves web/css/, web/js/, web/components/
		h.mux.Handle("/vendor/", api.VendorHandler(staticCfg))  // Serves web/vendor/automerge.js
		h.mux.HandleFunc("/", h.handleRoot)                     // Serves web/index.html for root
	}
}

// handleRoot serves the web UI index page
func (h *HTTPServer) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, h.cfg.WebPath+"/index.html")
	} else {
		http.NotFound(w, r)
	}
}

// ListenAndServe starts the HTTP server on the configured port.
//
// This is a blocking call that runs until the server is stopped or encounters an error.
func (h *HTTPServer) ListenAndServe() error {
	addr := ":" + h.cfg.Port
	log.Printf("[%s] Server starting on http://localhost:%s", h.cfg.UserID, h.cfg.Port)

	if !h.cfg.EnableUI {
		log.Printf("[%s] UI disabled - only API routes available", h.cfg.UserID)
	}

	return http.ListenAndServe(addr, h.mux)
}

// Server returns the underlying Automerge server.
//
// Use this if you need direct access to server methods for custom routes.
func (h *HTTPServer) Server() *server.Server {
	return h.server
}

// Mux returns the HTTP mux for adding custom routes.
//
// Example:
//
//	srv, _ := httpserver.New(cfg)
//	srv.Mux().HandleFunc("/custom", customHandler)
//	srv.ListenAndServe()
func (h *HTTPServer) Mux() *http.ServeMux {
	return h.mux
}
