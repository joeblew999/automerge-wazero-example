package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// MapPayload represents the JSON payload for map operations
type MapPayload struct {
	Path  string `json:"path"`  // Path to the map object (e.g., "ROOT" or "ROOT.users")
	Key   string `json:"key"`   // Map key
	Value string `json:"value"` // String value (for simplicity, we'll support strings first)
}

// MapKeysResponse represents the response for GET /api/map/keys
type MapKeysResponse struct {
	Keys []string `json:"keys"`
}

// MapHandler handles Map CRDT operations
// GET /api/map?path=ROOT&key=name - Get value at key
// POST /api/map {path, key, value} - Set key/value
// DELETE /api/map?path=ROOT&key=name - Delete key
func MapHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		switch r.Method {
		case http.MethodGet:
			// Get value at key
			path := r.URL.Query().Get("path")
			key := r.URL.Query().Get("key")

			if path == "" || key == "" {
				http.Error(w, "Missing path or key parameter", http.StatusBadRequest)
				return
			}

			value, err := srv.GetMapValue(ctx, parsePathString(path), key)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to get value: %v", err), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"value": value})

		case http.MethodPost:
			// Set key/value
			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read body", http.StatusBadRequest)
				return
			}

			var payload MapPayload
			if err := json.Unmarshal(body, &payload); err != nil {
				http.Error(w, "Invalid JSON", http.StatusBadRequest)
				return
			}

			if payload.Path == "" || payload.Key == "" {
				http.Error(w, "Missing path or key", http.StatusBadRequest)
				return
			}

			if err := srv.PutMapValue(ctx, parsePathString(payload.Path), payload.Key, payload.Value); err != nil {
				http.Error(w, fmt.Sprintf("Failed to put value: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			log.Printf("Map PUT: path=%s, key=%s, value=%s", payload.Path, payload.Key, payload.Value)

		case http.MethodDelete:
			// Delete key
			path := r.URL.Query().Get("path")
			key := r.URL.Query().Get("key")

			if path == "" || key == "" {
				http.Error(w, "Missing path or key parameter", http.StatusBadRequest)
				return
			}

			if err := srv.DeleteMapKey(ctx, parsePathString(path), key); err != nil {
				http.Error(w, fmt.Sprintf("Failed to delete key: %v", err), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusNoContent)
			log.Printf("Map DELETE: path=%s, key=%s", path, key)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}

// MapKeysHandler handles GET /api/map/keys?path=ROOT - List all keys in map
func MapKeysHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		path := r.URL.Query().Get("path")
		if path == "" {
			http.Error(w, "Missing path parameter", http.StatusBadRequest)
			return
		}

		keys, err := srv.GetMapKeys(ctx, parsePathString(path))
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to get keys: %v", err), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MapKeysResponse{Keys: keys})
	}
}
