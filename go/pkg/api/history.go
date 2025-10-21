package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// HistoryResponse represents the response for history operations
type HistoryResponse struct {
	Heads []string `json:"heads"`
}

// ChangesResponse represents the response for changes operations
type ChangesResponse struct {
	Changes string `json:"changes"` // Base64-encoded change data
	Size    int    `json:"size"`    // Size in bytes
}

// HeadsHandler handles GET /api/heads - Get current heads
func HeadsHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		heads, err := srv.GetHeads(ctx)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get heads: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HistoryResponse{Heads: heads})
	}
}

// ChangesHandler handles GET /api/changes - Get changes (optionally filtered)
// Query params: ?since=hash1,hash2 (comma-separated)
func ChangesHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()

		// For now, get all changes
		// TODO: Parse 'since' parameter to filter changes
		changes, err := srv.GetChanges(ctx, nil)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get changes: %v", err), http.StatusInternalServerError)
			return
		}

		// Encode binary changes as base64
		changesB64 := base64.StdEncoding.EncodeToString(changes)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(ChangesResponse{
			Changes: changesB64,
			Size:    len(changes),
		})
	}
}
