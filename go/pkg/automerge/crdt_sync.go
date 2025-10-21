// ==============================================================================
// Layer 4: Go High-Level CRDT API - Sync Protocol
// ==============================================================================
// ARCHITECTURE: This is the high-level Go API layer (Layer 4/7).
//
// RESPONSIBILITIES:
// - Pure CRDT operations (stateless, no mutex, no persistence)
// - Type-safe Go interface wrapping FFI calls
// - Convenience methods combining multiple FFI calls
// - Documentation of CRDT semantics
//
// DEPENDENCIES:
// - Layer 3: pkg/wazero (FFI to WASM)
// - Context: Takes context.Context for FFI calls
// - Runtime: Uses *wazero.Runtime directly
//
// DEPENDENTS:
// - Layer 5: pkg/server (stateful, thread-safe operations)
//
// RELATED FILES (1:1 mapping):
// - Layer 2: rust/automerge_wasi/src/sync.rs (WASI exports)
// - Layer 3: pkg/wazero/crdt_sync.go (FFI wrappers)
// - Layer 5: pkg/server/sync.go (stateful server operations)
// - Layer 6: pkg/api/sync.go (HTTP handlers)
// - Layer 7: web/js/sync.js + web/components/sync.html (frontend)
//
// NOTES:
// - This layer is pure CRDT - no state, no mutex, no persistence
// - All state management happens in Layer 5 (pkg/server)
// - Sync is per-peer, not global (each peer has its own SyncState)
// ==============================================================================

package automerge

import (
	"context"
	"fmt"
)

// Sync Protocol Operations - M1 Milestone
//
// The sync protocol enables efficient delta-based synchronization between
// peers. Instead of sending the entire document, peers exchange only the
// changes they don't have yet.

// InitSyncState initializes the sync state for a new peer connection.
//
// Call this once before starting a sync session with a peer.
// Returns a SyncState that must be used in subsequent sync calls.
//
// Status: ✅ Implemented
func (d *Document) InitSyncState(ctx context.Context) (*SyncState, error) {
	if d.runtime == nil {
		return nil, fmt.Errorf("document not initialized")
	}

	peerID, err := d.runtime.AmSyncStateInit(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to init sync state: %w", err)
	}

	return &SyncState{peerID: peerID}, nil
}

// FreeSyncState frees the sync state for a peer connection.
//
// Call this when done with a sync session.
//
// Status: ✅ Implemented
func (d *Document) FreeSyncState(ctx context.Context, state *SyncState) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}
	if state == nil {
		return fmt.Errorf("sync state is nil")
	}

	return d.runtime.AmSyncStateFree(ctx, state.peerID)
}

// GenerateSyncMessage generates a sync message to send to a peer.
//
// The sync state tracks what the peer has already seen, so we only send
// changes they're missing.
//
// Returns empty slice if there's nothing to send.
//
// Status: ✅ Implemented
func (d *Document) GenerateSyncMessage(ctx context.Context, state *SyncState) ([]byte, error) {
	if d.runtime == nil {
		return nil, fmt.Errorf("document not initialized")
	}
	if state == nil {
		return nil, fmt.Errorf("sync state is nil")
	}

	msg, err := d.runtime.AmSyncGen(ctx, state.peerID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate sync message: %w", err)
	}

	return msg, nil
}

// ReceiveSyncMessage processes a sync message from a peer.
//
// This applies any changes we don't have yet and updates the sync state.
// After receiving a message, you should call GenerateSyncMessage to see if
// we need to reply with our own changes.
//
// Status: ✅ Implemented
func (d *Document) ReceiveSyncMessage(ctx context.Context, state *SyncState, msg []byte) error {
	if d.runtime == nil {
		return fmt.Errorf("document not initialized")
	}
	if state == nil {
		return fmt.Errorf("sync state is nil")
	}

	return d.runtime.AmSyncRecv(ctx, state.peerID, msg)
}

// EncodeSyncMessage encodes a sync message for transmission.
//
// Status: Not needed - sync messages are already encoded bytes
func EncodeSyncMessage(msg []byte) ([]byte, error) {
	// Sync messages from am_sync_gen are already encoded
	return msg, nil
}

// DecodeSyncMessage decodes a received sync message.
//
// Status: Not needed - am_sync_recv handles decoding internally
func DecodeSyncMessage(data []byte) ([]byte, error) {
	// Sync messages are passed directly to am_sync_recv
	return data, nil
}
