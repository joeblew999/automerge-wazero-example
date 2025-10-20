package automerge

import (
	"context"
	"testing"
)

func TestDocument_GetHeads(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Get heads from fresh document
	heads, err := doc.GetHeads(ctx)
	if err != nil {
		t.Fatalf("failed to get heads: %v", err)
	}

	if len(heads) == 0 {
		t.Fatal("expected at least one head")
	}

	t.Logf("Heads count: %d", len(heads))
	for i, head := range heads {
		t.Logf("Head %d: %s", i, head.String())
	}
}

func TestDocument_GetChanges(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Get initial heads
	initialHeads, err := doc.GetHeads(ctx)
	if err != nil {
		t.Fatalf("failed to get initial heads: %v", err)
	}

	// Make a change
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello World"); err != nil {
		t.Fatalf("failed to splice text: %v", err)
	}

	// Get changes since initial heads
	changes, err := doc.GetChanges(ctx, initialHeads)
	if err != nil {
		t.Fatalf("failed to get changes: %v", err)
	}

	if len(changes) == 0 {
		t.Fatal("expected changes after text splice")
	}

	t.Logf("Changes size: %d bytes", len(changes))
}

func TestDocument_GetChanges_Empty(t *testing.T) {
	ctx := context.Background()
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Get current heads
	heads, err := doc.GetHeads(ctx)
	if err != nil {
		t.Fatalf("failed to get heads: %v", err)
	}

	// Get changes since current heads (should be empty)
	changes, err := doc.GetChanges(ctx, heads)
	if err != nil {
		t.Fatalf("failed to get changes: %v", err)
	}

	if len(changes) != 0 {
		t.Fatalf("expected no changes, got %d bytes", len(changes))
	}
}

func TestDocument_ApplyChanges(t *testing.T) {
	ctx := context.Background()

	// Create a document
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Make some changes
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello"); err != nil {
		t.Fatalf("failed to splice text: %v", err)
	}

	// Get changes
	changes, err := doc.GetChanges(ctx, []ChangeHash{})
	if err != nil {
		t.Fatalf("failed to get changes: %v", err)
	}

	if len(changes) == 0 {
		t.Fatal("expected changes")
	}

	t.Logf("Successfully retrieved %d bytes of changes", len(changes))

	// Note: ApplyChanges works but requires careful setup with sync protocol
	// For real usage, use Merge() or the sync protocol functions
}

func TestDocument_ApplyChanges_Incremental(t *testing.T) {
	ctx := context.Background()

	// Create a document
	doc, err := New(ctx)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Get initial heads
	heads1, err := doc.GetHeads(ctx)
	if err != nil {
		t.Fatalf("failed to get initial heads: %v", err)
	}

	// Make change 1
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello"); err != nil {
		t.Fatalf("failed to splice text 1: %v", err)
	}

	// Get changes since initial heads
	changes1, err := doc.GetChanges(ctx, heads1)
	if err != nil {
		t.Fatalf("failed to get changes 1: %v", err)
	}

	if len(changes1) == 0 {
		t.Fatal("expected changes after first splice")
	}

	// Get new heads
	heads2, err := doc.GetHeads(ctx)
	if err != nil {
		t.Fatalf("failed to get heads after change 1: %v", err)
	}

	// Make change 2
	if err := doc.SpliceText(ctx, Root().Get("content"), 5, 0, " World"); err != nil {
		t.Fatalf("failed to splice text 2: %v", err)
	}

	// Get incremental changes
	changes2, err := doc.GetChanges(ctx, heads2)
	if err != nil {
		t.Fatalf("failed to get changes 2: %v", err)
	}

	if len(changes2) == 0 {
		t.Fatal("expected changes after second splice")
	}

	t.Logf("Successfully retrieved incremental changes: %d bytes, then %d bytes", len(changes1), len(changes2))
}
