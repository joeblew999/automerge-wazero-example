package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// CursorGetRequest represents a request to get a cursor at a position
type CursorGetRequest struct {
	Path  string `json:"path"`  // Object path (e.g., "ROOT.content")
	Index int    `json:"index"` // Position in the object
}

// CursorGetResponse represents the response with cursor information
type CursorGetResponse struct {
	Path   string `json:"path"`   // Object path
	Index  int    `json:"index"`  // Original position
	Cursor string `json:"cursor"` // Cursor value (opaque string)
}

// CursorLookupRequest represents a request to lookup a cursor's position
type CursorLookupRequest struct {
	Path   string `json:"path"`   // Object path
	Cursor string `json:"cursor"` // Cursor value to lookup
}

// CursorLookupResponse represents the response with cursor position
type CursorLookupResponse struct {
	Path   string `json:"path"`   // Object path
	Cursor string `json:"cursor"` // Cursor value
	Index  int    `json:"index"`  // Current position
}

// CursorGetHandler returns a handler for GET /api/cursor and POST /api/cursor
// GET with query params: ?path=ROOT.content&index=5
// POST with JSON body: {"path": "ROOT.content", "index": 5}
func CursorGetHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		var path string
		var index int
		var err error

		if r.Method == http.MethodGet {
			// Parse query parameters
			path = r.URL.Query().Get("path")
			if path == "" {
				http.Error(w, "missing path parameter", http.StatusBadRequest)
				return
			}

			indexStr := r.URL.Query().Get("index")
			if indexStr == "" {
				http.Error(w, "missing index parameter", http.StatusBadRequest)
				return
			}

			index, err = strconv.Atoi(indexStr)
			if err != nil {
				http.Error(w, "invalid index parameter", http.StatusBadRequest)
				return
			}
		} else if r.Method == http.MethodPost {
			// Parse JSON body
			var req CursorGetRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, "invalid JSON", http.StatusBadRequest)
				return
			}

			path = req.Path
			index = req.Index
		} else {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Get cursor from server
		cursor, err := srv.GetCursor(ctx, path, index)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return response
		resp := CursorGetResponse{
			Path:   cursor.Path,
			Index:  index,
			Cursor: cursor.Value,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// CursorLookupHandler returns a handler for POST /api/cursor/lookup
// POST with JSON body: {"path": "ROOT.content", "cursor": "..."}
func CursorLookupHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()

		// Parse JSON body
		var req CursorLookupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid JSON", http.StatusBadRequest)
			return
		}

		if req.Path == "" {
			http.Error(w, "missing path", http.StatusBadRequest)
			return
		}

		if req.Cursor == "" {
			http.Error(w, "missing cursor", http.StatusBadRequest)
			return
		}

		// Create cursor object
		cursor := &automerge.Cursor{
			Path:  req.Path,
			Value: req.Cursor,
		}

		// Lookup cursor position
		index, err := srv.LookupCursor(ctx, cursor)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return response
		resp := CursorLookupResponse{
			Path:   req.Path,
			Cursor: req.Cursor,
			Index:  index,
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
