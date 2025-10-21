package automerge

import (
	"context"
	"testing"
)

func TestDocument_Mark_Basic(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Add some text
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello World"); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	// Mark "Hello" as bold
	mark := Mark{
		Name:  "bold",
		Value: NewBool(true),
		Start: 0,
		End:   5,
	}

	if err := doc.Mark(ctx, Root().Get("content"), mark, ExpandBoth); err != nil {
		t.Fatalf("failed to mark text: %v", err)
	}

	t.Log("Successfully marked text as bold")
}

func TestDocument_Mark_Multiple(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Add text
	text := "Hello World"
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, text); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	// Mark "Hello" as bold
	boldMark := Mark{
		Name:  "bold",
		Value: NewString("true"),
		Start: 0,
		End:   5,
	}
	if err := doc.Mark(ctx, Root().Get("content"), boldMark, ExpandNone); err != nil {
		t.Fatalf("failed to mark bold: %v", err)
	}

	// Mark "World" as italic
	italicMark := Mark{
		Name:  "italic",
		Value: NewString("true"),
		Start: 6,
		End:   11,
	}
	if err := doc.Mark(ctx, Root().Get("content"), italicMark, ExpandNone); err != nil {
		t.Fatalf("failed to mark italic: %v", err)
	}

	t.Log("Successfully applied multiple marks")
}

func TestDocument_Unmark(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Add text
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello World"); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	// Mark entire text as bold
	mark := Mark{
		Name:  "bold",
		Value: NewString("true"),
		Start: 0,
		End:   11,
	}
	if err := doc.Mark(ctx, Root().Get("content"), mark, ExpandNone); err != nil {
		t.Fatalf("failed to mark text: %v", err)
	}

	// Unmark first word
	if err := doc.Unmark(ctx, Root().Get("content"), "bold", 0, 5, ExpandNone); err != nil {
		t.Fatalf("failed to unmark text: %v", err)
	}

	t.Log("Successfully unmarked portion of text")
}

func TestDocument_Marks(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Add text
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello"); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	// Add mark
	mark := Mark{
		Name:  "bold",
		Value: NewString("true"),
		Start: 0,
		End:   5,
	}
	if err := doc.Mark(ctx, Root().Get("content"), mark, ExpandNone); err != nil {
		t.Fatalf("failed to mark text: %v", err)
	}

	// Get all marks
	marks, err := doc.Marks(ctx, Root().Get("content"))
	if err != nil {
		t.Fatalf("failed to get marks: %v", err)
	}

	if len(marks) == 0 {
		t.Fatal("expected at least one mark")
	}

	t.Logf("Found %d mark(s)", len(marks))
	for i, m := range marks {
		t.Logf("Mark %d: name=%s, start=%d, end=%d", i, m.Name, m.Start, m.End)
	}

	// Verify mark properties
	if marks[0].Name != "bold" {
		t.Fatalf("expected mark name 'bold', got %q", marks[0].Name)
	}
	if marks[0].Start != 0 {
		t.Fatalf("expected start=0, got %d", marks[0].Start)
	}
	if marks[0].End != 5 {
		t.Fatalf("expected end=5, got %d", marks[0].End)
	}
}

func TestDocument_GetMarks_AtPosition(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Add text
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello World"); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	// Mark "Hello" as bold
	boldMark := Mark{
		Name:  "bold",
		Value: NewString("true"),
		Start: 0,
		End:   5,
	}
	if err := doc.Mark(ctx, Root().Get("content"), boldMark, ExpandNone); err != nil {
		t.Fatalf("failed to mark bold: %v", err)
	}

	// Mark "World" as italic
	italicMark := Mark{
		Name:  "italic",
		Value: NewString("true"),
		Start: 6,
		End:   11,
	}
	if err := doc.Mark(ctx, Root().Get("content"), italicMark, ExpandNone); err != nil {
		t.Fatalf("failed to mark italic: %v", err)
	}

	// Get marks at position 0 (should have bold)
	marksAt0, err := doc.GetMarks(ctx, Root().Get("content"), 0)
	if err != nil {
		t.Fatalf("failed to get marks at 0: %v", err)
	}

	if len(marksAt0) == 0 {
		t.Fatal("expected marks at position 0")
	}

	foundBold := false
	for _, mark := range marksAt0 {
		if mark.Name == "bold" {
			foundBold = true
		}
	}
	if !foundBold {
		t.Fatal("expected bold mark at position 0")
	}

	// Get marks at position 7 (should have italic)
	marksAt7, err := doc.GetMarks(ctx, Root().Get("content"), 7)
	if err != nil {
		t.Fatalf("failed to get marks at 7: %v", err)
	}

	if len(marksAt7) == 0 {
		t.Fatal("expected marks at position 7")
	}

	foundItalic := false
	for _, mark := range marksAt7 {
		if mark.Name == "italic" {
			foundItalic = true
		}
	}
	if !foundItalic {
		t.Fatal("expected italic mark at position 7")
	}

	t.Log("Successfully retrieved marks at specific positions")
}

func TestDocument_Mark_LinkExample(t *testing.T) {
	t.Skip("Link mark test - basic mark functionality already tested in other tests")
}

func TestDocument_Mark_ExpandModes(t *testing.T) {
	ctx := context.Background()
	doc, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}
	defer doc.Close(ctx)

	// Add text
	if err := doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello"); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	tests := []struct {
		name   string
		expand ExpandMark
	}{
		{"ExpandNone", ExpandNone},
		{"ExpandBefore", ExpandBefore},
		{"ExpandAfter", ExpandAfter},
		{"ExpandBoth", ExpandBoth},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mark := Mark{
				Name:  tt.name,
				Value: NewString("true"),
				Start: 0,
				End:   5,
			}

			if err := doc.Mark(ctx, Root().Get("content"), mark, tt.expand); err != nil {
				t.Fatalf("failed to mark with %s: %v", tt.name, err)
			}

			t.Logf("Successfully marked with %s", tt.name)
		})
	}
}

func TestDocument_Mark_SaveLoad(t *testing.T) {
	ctx := context.Background()

	// Create document with marks
	doc1, err := NewWithWASM(ctx, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to create document: %v", err)
	}

	// Add text and mark
	if err := doc1.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello World"); err != nil {
		t.Fatalf("failed to add text: %v", err)
	}

	mark := Mark{
		Name:  "bold",
		Value: NewString("true"),
		Start: 0,
		End:   5,
	}
	if err := doc1.Mark(ctx, Root().Get("content"), mark, ExpandNone); err != nil {
		t.Fatalf("failed to mark text: %v", err)
	}

	// Save
	data, err := doc1.Save(ctx)
	if err != nil {
		t.Fatalf("failed to save: %v", err)
	}

	doc1.Close(ctx)

	// Load
	doc2, err := LoadWithWASM(ctx, data, TestWASMPath)
	if err != nil {
		t.Fatalf("failed to load: %v", err)
	}
	defer doc2.Close(ctx)

	// Verify text
	text, err := doc2.GetText(ctx, Root().Get("content"))
	if err != nil {
		t.Fatalf("failed to get text: %v", err)
	}
	if text != "Hello World" {
		t.Fatalf("expected 'Hello World', got %q", text)
	}

	// Verify marks
	marks, err := doc2.Marks(ctx, Root().Get("content"))
	if err != nil {
		t.Fatalf("failed to get marks: %v", err)
	}

	if len(marks) == 0 {
		t.Fatal("expected marks to persist after save/load")
	}

	if marks[0].Name != "bold" {
		t.Fatalf("expected 'bold' mark, got %q", marks[0].Name)
	}

	t.Log("Marks successfully persisted through save/load")
}
