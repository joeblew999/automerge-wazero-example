// ==============================================================================
// Layer 6: HTTP API - Rich Text (Marks & Formatting)
// ==============================================================================
// ARCHITECTURE: This is the HTTP protocol layer (Layer 6/7).
//
// RESPONSIBILITIES:
// - HTTP request parsing (JSON body, query params, headers)
// - HTTP response formatting (JSON, status codes, headers)
// - Input validation (HTTP-level)
// - Protocol translation (HTTP ↔ Go function calls)
//
// DEPENDENCIES:
// - Layer 5: pkg/server (business logic, state management)
//
// DEPENDENTS:
// - None (top of backend stack)
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/richtext.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_richtext.go (FFI wrappers)
// - Layer 4: pkg/automerge/crdt_richtext.go (pure CRDT API)
// - Layer 5: pkg/server/crdt_richtext.go (stateful server operations)
// - Layer 7: web/js/crdt_richtext.js + web/components/crdt_richtext.html
//
// NOTES:
// - This layer is stateless (doesn't own any application state)
// - All state management is delegated to Layer 5 (pkg/server)
// - Handles HTTP protocol concerns (status codes, content-type, etc.)
// ==============================================================================

package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// RichTextMarkPayload represents the JSON payload for mark operations
type RichTextMarkPayload struct {
	Path   string `json:"path"`   // Path to text object
	Name   string `json:"name"`   // Mark name (e.g., "bold", "italic", "link")
	Value  string `json:"value"`  // Mark value (e.g., "true", URL for links)
	Start  uint   `json:"start"`  // Start position (inclusive)
	End    uint   `json:"end"`    // End position (exclusive)
	Expand string `json:"expand"` // "before", "after", "both", "none"
}

// RichTextUnmarkPayload represents the JSON payload for unmark operations
type RichTextUnmarkPayload struct {
	Path   string `json:"path"`   // Path to text object
	Name   string `json:"name"`   // Mark name to remove
	Start  uint   `json:"start"`  // Start position (inclusive)
	End    uint   `json:"end"`    // End position (exclusive)
	Expand string `json:"expand"` // "before", "after", "both", "none"
}

// RichTextMarksResponse represents the response for GetMarks
type RichTextMarksResponse struct {
	Marks []MarkJSON `json:"marks"` // Array of marks at position
}

// MarkJSON is the JSON representation of a Mark
type MarkJSON struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"` // Can be string, bool, int, float
	Start uint        `json:"start"`
	End   uint        `json:"end"`
}

// parseExpandMark converts string to ExpandMark enum
func parseExpandMark(expand string) automerge.ExpandMark {
	switch expand {
	case "before":
		return automerge.ExpandBefore
	case "after":
		return automerge.ExpandAfter
	case "both":
		return automerge.ExpandBoth
	case "none":
		return automerge.ExpandNone
	default:
		return automerge.ExpandNone
	}
}

// RichTextMarkHandler handles POST /api/richtext/mark - Apply formatting mark
// M2 Milestone: Rich text marks
func RichTextMarkHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var payload RichTextMarkPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Name == "" {
			http.Error(w, "Missing mark name", http.StatusBadRequest)
			return
		}

		// Create Mark struct
		mark := automerge.Mark{
			Name:  payload.Name,
			Value: automerge.NewString(payload.Value),
			Start: payload.Start,
			End:   payload.End,
		}

		expand := parseExpandMark(payload.Expand)

		if err := srv.RichTextMark(ctx, parsePathString(payload.Path), mark, expand); err != nil {
			http.Error(w, fmt.Sprintf("Failed to apply mark: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("RichText MARK: path=%s, name=%s, range=[%d,%d), expand=%s", payload.Path, payload.Name, payload.Start, payload.End, payload.Expand)
	}
}

// RichTextUnmarkHandler handles POST /api/richtext/unmark - Remove formatting mark
// M2 Milestone: Rich text marks
func RichTextUnmarkHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Failed to read body", http.StatusBadRequest)
			return
		}

		var payload RichTextUnmarkPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.Name == "" {
			http.Error(w, "Missing mark name", http.StatusBadRequest)
			return
		}

		expand := parseExpandMark(payload.Expand)

		if err := srv.RichTextUnmark(ctx, parsePathString(payload.Path), payload.Name, payload.Start, payload.End, expand); err != nil {
			http.Error(w, fmt.Sprintf("Failed to remove mark: %v", err), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
		log.Printf("RichText UNMARK: path=%s, name=%s, range=[%d,%d), expand=%s", payload.Path, payload.Name, payload.Start, payload.End, payload.Expand)
	}
}

// RichTextMarksHandler handles GET /api/richtext/marks?path=ROOT.content&pos=5 - Get marks at position
// M2 Milestone: Rich text marks
func RichTextMarksHandler(srv *server.Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := r.Context()
		path := r.URL.Query().Get("path")
		posStr := r.URL.Query().Get("pos")

		if path == "" || posStr == "" {
			http.Error(w, "Missing path or pos parameter", http.StatusBadRequest)
			return
		}

		var pos uint
		if _, err := fmt.Sscanf(posStr, "%d", &pos); err != nil {
			http.Error(w, "Invalid pos parameter", http.StatusBadRequest)
			return
		}

		marks, err := srv.GetRichTextMarks(ctx, parsePathString(path), pos)
		if err != nil {
			log.Printf("ERROR: Failed to get marks at pos %d: %v", pos, err)
			http.Error(w, fmt.Sprintf("Failed to get marks: %v", err), http.StatusInternalServerError)
			return
		}

		log.Printf("GetMarks: path=%s, pos=%d, found %d marks", path, pos, len(marks))

		// Convert automerge.Mark to MarkJSON
		jsonMarks := make([]MarkJSON, len(marks))
		for i, mark := range marks {
			var value interface{}
			if s, ok := mark.Value.AsString(); ok {
				value = s
			} else if b, ok := mark.Value.AsBool(); ok {
				value = b
			} else if n, ok := mark.Value.AsInt(); ok {
				value = n
			} else if f, ok := mark.Value.AsFloat(); ok {
				value = f
			} else {
				value = nil
			}

			jsonMarks[i] = MarkJSON{
				Name:  mark.Name,
				Value: value,
				Start: mark.Start,
				End:   mark.End,
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(RichTextMarksResponse{Marks: jsonMarks})
	}
}
