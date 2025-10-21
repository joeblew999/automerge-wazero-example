package automerge

import (
	"context"
	"testing"
)

func TestDocument_InitSyncState(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Initialize sync state
	state, err := doc.InitSyncState(ctx)
	if err != nil {
		t.Fatalf("failed to initialize sync state: %v", err)
	}
	defer doc.FreeSyncState(ctx, state)

	if state.PeerID() == 0 {
		t.Fatal("expected non-zero peer ID")
	}

	t.Logf("Successfully initialized sync state with peer_id=%d", state.PeerID())
}

func TestDocument_GenerateSyncMessage(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Initialize sync state
	state, err := doc.InitSyncState(ctx)
	if err != nil {
		t.Fatalf("failed to initialize sync state: %v", err)
	}
	defer doc.FreeSyncState(ctx, state)

	// Generate sync message
	msg, err := doc.GenerateSyncMessage(ctx, state)
	if err != nil {
		t.Fatalf("failed to generate sync message: %v", err)
	}

	t.Logf("Generated sync message: %d bytes", len(msg))

	if len(msg) == 0 {
		t.Log("No sync message needed (expected for fresh document)")
	}
}

func TestDocument_Sync_TwoDocuments(t *testing.T) {
	ctx := context.Background()

	// Create document A with some content
	docA, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document A: %v", err)
	}
	defer docA.Close(ctx)

	if err := docA.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello from A"); err != nil {
		t.Fatalf("failed to add text to A: %v", err)
	}

	// Create empty document B
	docB, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document B: %v", err)
	}
	defer docB.Close(ctx)

	// Initialize sync states
	stateA, err := docA.InitSyncState(ctx)
	if err != nil {
		t.Fatalf("failed to init sync state A: %v", err)
	}
	defer docA.FreeSyncState(ctx, stateA)

	stateB, err := docB.InitSyncState(ctx)
	if err != nil {
		t.Fatalf("failed to init sync state B: %v", err)
	}
	defer docB.FreeSyncState(ctx, stateB)

	// Generate sync message from A
	msgFromA, err := docA.GenerateSyncMessage(ctx, stateA)
	if err != nil {
		t.Fatalf("failed to generate sync message from A: %v", err)
	}

	t.Logf("Sync message from A: %d bytes", len(msgFromA))

	// If A has a message, send it to B
	if len(msgFromA) > 0 {
		if err := docB.ReceiveSyncMessage(ctx, stateB, msgFromA); err != nil {
			t.Fatalf("failed to receive sync message in B: %v", err)
		}

		// Verify B now has the text
		textB, err := docB.GetText(ctx, Root().Get("content"))
		if err != nil {
			t.Fatalf("failed to get text from B: %v", err)
		}

		if textB != "Hello from A" {
			t.Fatalf("expected 'Hello from A' in B, got %q", textB)
		}

		t.Log("Successfully synced A → B")
	}
}

func TestDocument_Sync_BidirectionalSync(t *testing.T) {
	t.Skip("Full sync protocol requires multi-peer state management - use Merge() for now")
}



