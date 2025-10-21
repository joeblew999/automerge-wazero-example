// ═══════════════════════════════════════════════════════════════
// LAYER 3: Go FFI Wrappers (wazero → WASM)
//
// Responsibilities:
// - Call WASM functions via wazero runtime
// - Marshal Go strings/data to WASM linear memory
// - Translate WASM error codes to Go errors
// - Manage WASM memory allocation/deallocation
//
// Dependencies:
// ⬇️  Calls: WASM functions (am_sync_*)
//           Implemented in: rust/automerge_wasi/src/sync.rs (Layer 2)
// ⬆️  Called by: go/pkg/automerge/crdt_sync.go (Layer 4 - high-level API)
//
// Related Files:
// 🔁 Siblings: crdt_text.go, crdt_map.go, crdt_list.go, crdt_counter.go, crdt_richtext.go
// 📝 Tests: crdt_sync_test.go (FFI boundary tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-3-go-ffi
// ═══════════════════════════════════════════════════════════════

package wazero

import (
	"context"
	"fmt"
)

// AmSyncStateInit creates a new sync state for a peer connection
// Returns peer_id (> 0) on success, or 0 on error
func (r *Runtime) AmSyncStateInit(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_sync_state_init")
	if err != nil {
		return 0, err
	}
	peerID := uint32(results[0])
	if peerID == 0 {
		return 0, fmt.Errorf("am_sync_state_init failed to create peer")
	}
	return peerID, nil
}

// AmSyncStateFree frees a peer's sync state
func (r *Runtime) AmSyncStateFree(ctx context.Context, peerID uint32) error {
	results, err := r.callExport(ctx, "am_sync_state_free", uint64(peerID))
	if err != nil {
		return err
	}
	return checkErrorCode("am_sync_state_free", results)
}

// AmSyncGenLen returns the length of the sync message to generate
func (r *Runtime) AmSyncGenLen(ctx context.Context, peerID uint32) (uint32, error) {
	results, err := r.callExport(ctx, "am_sync_gen_len", uint64(peerID))
	if err != nil {
		return 0, err
	}
	len := uint32(results[0])
	if len == ^uint32(0) { // u32::MAX indicates error
		return 0, fmt.Errorf("am_sync_gen_len returned error")
	}
	return len, nil
}

// AmSyncGen generates a sync message to send to a peer
func (r *Runtime) AmSyncGen(ctx context.Context, peerID uint32) ([]byte, error) {
	// Get message length
	msgLen, err := r.AmSyncGenLen(ctx, peerID)
	if err != nil {
		return nil, err
	}
	if msgLen == 0 {
		return []byte{}, nil // Nothing to send
	}

	// Allocate buffer
	msgPtr, err := r.AmAlloc(ctx, msgLen)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate sync message buffer: %w", err)
	}
	defer r.AmFree(ctx, msgPtr, msgLen)

	// Generate sync message
	results, err := r.callExport(ctx, "am_sync_gen", uint64(peerID), uint64(msgPtr))
	if err != nil {
		return nil, err
	}

	code := int32(results[0])
	if code == 1 {
		// Nothing to send (not an error)
		return []byte{}, nil
	}
	if code != 0 {
		return nil, fmt.Errorf("am_sync_gen returned error code: %d", code)
	}

	// Read message from memory
	mem := r.Memory()
	msgBytes, ok := mem.Read(msgPtr, msgLen)
	if !ok {
		return nil, fmt.Errorf("failed to read sync message from WASM memory")
	}

	// Make a copy
	result := make([]byte, len(msgBytes))
	copy(result, msgBytes)
	return result, nil
}

// AmSyncRecv receives and processes a sync message from a peer
func (r *Runtime) AmSyncRecv(ctx context.Context, peerID uint32, msg []byte) error {
	if len(msg) == 0 {
		return fmt.Errorf("empty sync message")
	}

	msgLen := uint32(len(msg))
	msgPtr, err := r.AmAlloc(ctx, msgLen)
	if err != nil {
		return fmt.Errorf("failed to allocate sync message buffer: %w", err)
	}
	defer r.AmFree(ctx, msgPtr, msgLen)

	// Write message to memory
	mem := r.Memory()
	if !mem.Write(msgPtr, msg) {
		return fmt.Errorf("failed to write sync message to WASM memory")
	}

	// Receive sync message
	results, err := r.callExport(ctx, "am_sync_recv", uint64(peerID), uint64(msgPtr), uint64(msgLen))
	if err != nil {
		return err
	}
	return checkErrorCode("am_sync_recv", results)
}
