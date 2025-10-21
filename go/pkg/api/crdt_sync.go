// ==============================================================================
// Layer 6: HTTP API - Sync Protocol
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
// - Layer 2: rust/automerge_wasi/src/sync.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_sync.go (FFI wrappers)
// - Layer 4: pkg/automerge/crdt_sync.go (pure CRDT API)
// - Layer 5: pkg/server/crdt_sync.go (stateful server operations)
// - Layer 7: web/js/sync.js + web/components/sync.html (frontend)
//
// NOTES:
// - This layer is stateless (doesn't own any application state)
// - All state management is delegated to Layer 5 (pkg/server)
// - Handles HTTP protocol concerns (status codes, content-type, etc.)
// ==============================================================================

package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
	"github.com/joeblew999/automerge-wazero-example/pkg/server"
)

// SyncPayload represents the JSON payload for sync operations
type SyncPayload struct {
	PeerID  string `json:"peer_id"`           // Peer identifier
	Message string `json:"message,omitempty"` // Base64-encoded sync message
}

// SyncResponse represents the response for sync operations
type SyncResponse struct {
	Message string `json:"message,omitempty"` // Base64-encoded sync message
	HasMore bool   `json:"has_more"`          // Whether more sync messages are needed
}

// SyncHandler handles POST /api/sync - Process sync message and generate response
// M1 Milestone: Automerge sync protocol
func SyncHandler(srv *server.Server) http.HandlerFunc {
	// In-memory sync states per peer
	// TODO: Move this to server package for persistence
	peerStates := make(map[string]*automerge.SyncState)

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

		var payload SyncPayload
		if err := json.Unmarshal(body, &payload); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		if payload.PeerID == "" {
			http.Error(w, "Missing peer_id", http.StatusBadRequest)
			return
		}

		// Get or create sync state for this peer
		state, exists := peerStates[payload.PeerID]
		if !exists {
			// Initialize sync state with the server's document
			newState, err := srv.InitSyncState(ctx)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to init sync state: %v", err), http.StatusInternalServerError)
				return
			}
			state = newState
			peerStates[payload.PeerID] = state
		}

		// If peer sent a sync message, receive it first
		if payload.Message != "" {
			messageBytes, err := base64.StdEncoding.DecodeString(payload.Message)
			if err != nil {
				http.Error(w, fmt.Sprintf("Failed to decode message: %v", err), http.StatusBadRequest)
				return
			}

			if err := srv.ReceiveSyncMessage(ctx, state, messageBytes); err != nil {
				http.Error(w, fmt.Sprintf("Failed to receive sync message: %v", err), http.StatusInternalServerError)
				return
			}

			log.Printf("Sync RECEIVE: peer=%s, message_size=%d bytes", payload.PeerID, len(messageBytes))
		}

		// Generate sync message for this peer
		responseMessage, err := srv.GenerateSyncMessage(ctx, state)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to generate sync message: %v", err), http.StatusInternalServerError)
			return
		}

		// Encode response as base64
		var responseB64 string
		if len(responseMessage) > 0 {
			responseB64 = base64.StdEncoding.EncodeToString(responseMessage)
		}

		// Check if more messages are needed
		hasMore := len(responseMessage) > 0

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(SyncResponse{
			Message: responseB64,
			HasMore: hasMore,
		})

		log.Printf("Sync SEND: peer=%s, message_size=%d bytes, has_more=%v", payload.PeerID, len(responseMessage), hasMore)
	}
}
