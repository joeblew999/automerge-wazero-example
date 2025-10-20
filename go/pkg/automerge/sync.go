package automerge

import "context"

// Sync Protocol Operations - For M1 Milestone
//
// The sync protocol enables efficient delta-based synchronization between
// peers. Instead of sending the entire document, peers exchange only the
// changes they don't have yet.

// GenerateSyncMessage generates a sync message to send to a peer.
//
// The sync state tracks what the peer has already seen, so we only send
// changes they're missing.
//
// Returns nil if there's nothing to send.
//
// Status: ❌ Not implemented (requires M1)
func (d *Document) GenerateSyncMessage(ctx context.Context, state *SyncState) ([]byte, error) {
	return nil, &NotImplementedError{
		Feature:   "GenerateSyncMessage",
		Milestone: "M1",
		Message:   "Requires am_sync_gen WASI export",
	}
}

// ReceiveSyncMessage processes a sync message from a peer.
//
// This applies any changes we don't have yet and updates the sync state.
// After receiving a message, you should call GenerateSyncMessage to see if
// we need to reply with our own changes.
//
// Status: ❌ Not implemented (requires M1)
func (d *Document) ReceiveSyncMessage(ctx context.Context, state *SyncState, msg []byte) error {
	return &NotImplementedError{
		Feature:   "ReceiveSyncMessage",
		Milestone: "M1",
		Message:   "Requires am_sync_recv WASI export",
	}
}

// EncodeSyncMessage encodes a sync message for transmission.
//
// Status: ❌ Not implemented (requires M1)
func EncodeSyncMessage(msg []byte) ([]byte, error) {
	return nil, &NotImplementedError{
		Feature:   "EncodeSyncMessage",
		Milestone: "M1",
		Message:   "Requires am_sync_encode WASI export",
	}
}

// DecodeSyncMessage decodes a received sync message.
//
// Status: ❌ Not implemented (requires M1)
func DecodeSyncMessage(data []byte) ([]byte, error) {
	return nil, &NotImplementedError{
		Feature:   "DecodeSyncMessage",
		Milestone: "M1",
		Message:   "Requires am_sync_decode WASI export",
	}
}
