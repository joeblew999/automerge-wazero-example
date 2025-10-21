package automerge_test

import (
	"context"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Helper to create test document
func newTestDoc(t *testing.T) (*automerge.Document, context.Context) {
	t.Helper()
	ctx := context.Background()
	doc, err := automerge.New(ctx)
	if err != nil {
		t.Fatalf("New() failed: %v", err)
	}
	t.Cleanup(func() { doc.Close(ctx) })
	return doc, ctx
}

func TestDocument_GetCursor(t *testing.T) {
	doc, ctx := newTestDoc(t)
	path := automerge.Root().Get("content")

	// Add some text
	if err := doc.SpliceText(ctx, path, 0, 0, "Hello World"); err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		index     int
		wantError bool
	}{
		{"valid cursor at beginning", "ROOT.content", 0, false},
		{"valid cursor in middle", "ROOT.content", 5, false},
		{"valid cursor near end", "ROOT.content", 10, false}, // Last character 'd'
		{"invalid path", "INVALID.path", 0, true},
		{"index out of bounds", "ROOT.content", 999, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cursor, err := doc.GetCursor(ctx, tt.path, tt.index)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error but got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
			if cursor == nil {
				t.Fatal("Cursor is nil")
			}
			if cursor.Path != tt.path {
				t.Errorf("Expected path %s, got %s", tt.path, cursor.Path)
			}
			if cursor.Value == "" {
				t.Error("Cursor value is empty")
			}
			t.Logf("Cursor at index %d: %s", tt.index, cursor.Value)
		})
	}
}

func TestDocument_LookupCursor(t *testing.T) {
	doc, ctx := newTestDoc(t)
	path := automerge.Root().Get("content")

	if err := doc.SpliceText(ctx, path, 0, 0, "Hello World"); err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}

	cursor, err := doc.GetCursor(ctx, "ROOT.content", 5)
	if err != nil {
		t.Fatalf("Failed to get cursor: %v", err)
	}

	index, err := doc.LookupCursor(ctx, cursor)
	if err != nil {
		t.Fatalf("Failed to lookup cursor: %v", err)
	}

	if index != 5 {
		t.Errorf("Expected index 5, got %d", index)
	}
}

func TestDocument_CursorSurvivesEdits(t *testing.T) {
	doc, ctx := newTestDoc(t)
	path := automerge.Root().Get("content")

	if err := doc.SpliceText(ctx, path, 0, 0, "Hello World"); err != nil {
		t.Fatalf("Failed to add text: %v", err)
	}

	cursor, err := doc.GetCursor(ctx, "ROOT.content", 6)
	if err != nil {
		t.Fatalf("Failed to get cursor: %v", err)
	}

	t.Logf("Initial cursor at position 6: %s", cursor.Value)

	if err := doc.SpliceText(ctx, path, 0, 0, "Hi "); err != nil {
		t.Fatalf("Failed to insert text: %v", err)
	}

	text, err := doc.GetText(ctx, path)
	if err != nil {
		t.Fatalf("Failed to get text: %v", err)
	}

	if text != "Hi Hello World" {
		t.Errorf("Expected text 'Hi Hello World', got %q", text)
	}

	index, err := doc.LookupCursor(ctx, cursor)
	if err != nil {
		t.Fatalf("Failed to lookup cursor: %v", err)
	}

	if index != 9 {
		t.Errorf("Expected cursor at index 9, got %d", index)
	}

	t.Logf("Cursor moved from index 6 to index %d after inserting 'Hi '", index)
}

func TestCursor_String(t *testing.T) {
	tests := []struct {
		name     string
		cursor   *automerge.Cursor
		expected string
	}{
		{"nil cursor", nil, "<nil cursor>"},
		{"valid cursor", &automerge.Cursor{Path: "ROOT.content", Value: "abc123"}, "Cursor{path=ROOT.content, value=abc123}"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.cursor.String()
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
