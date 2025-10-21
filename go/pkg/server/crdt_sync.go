// ==============================================================================
// Layer 5: Go Server - Sync Protocol (Stateful + Thread-safe)
// ==============================================================================
// ARCHITECTURE: This is the stateful server layer (Layer 5/7).
//
// RESPONSIBILITIES:
// - Thread-safe CRDT operations (mutex protection)
// - State management (owns *automerge.Document, sync.RWMutex)
// - Persistence (saveDocument after mutations)
// - SSE broadcasting to connected clients
//
// DEPENDENCIES:
// - Layer 4: pkg/automerge (pure CRDT operations)
//
// DEPENDENTS:
// - Layer 6: pkg/api (HTTP handlers)
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/sync.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_sync.go (FFI wrappers)
// - Layer 4: pkg/automerge/crdt_sync.go (pure CRDT API)
// - Layer 6: pkg/api/crdt_sync.go (HTTP handlers)
// - Layer 7: web/js/sync.js + web/components/sync.html (frontend)
//
// NOTES:
// - All public methods are thread-safe (use s.mu.RLock/Lock)
// - Sync state is per-peer (not global)
// - This layer delegates to Layer 4 for actual CRDT operations
// ==============================================================================

package server

import (
	"context"
	"log"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Sync operations - maps to automerge/sync.go (M1 milestone)

// InitSyncState initializes a new sync state for a peer (thread-safe)
func (s *Server) InitSyncState(ctx context.Context) (*automerge.SyncState, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.InitSyncState(ctx)
}

// FreeSyncState frees a peer's sync state (thread-safe)
func (s *Server) FreeSyncState(ctx context.Context, state *automerge.SyncState) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.FreeSyncState(ctx, state)
}

// GenerateSyncMessage generates a sync message for the given peer (thread-safe)
func (s *Server) GenerateSyncMessage(ctx context.Context, state *automerge.SyncState) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.doc.GenerateSyncMessage(ctx, state)
}

// ReceiveSyncMessage processes a sync message from a peer (thread-safe)
func (s *Server) ReceiveSyncMessage(ctx context.Context, state *automerge.SyncState, message []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.doc.ReceiveSyncMessage(ctx, state, message); err != nil {
		return err
	}

	// Save after receiving sync (document may have been updated)
	if err := s.saveDocument(ctx); err != nil {
		log.Printf("Warning: failed to save snapshot after sync: %v", err)
	}

	return nil
}
