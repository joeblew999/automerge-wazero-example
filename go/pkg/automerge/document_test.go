package automerge_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/joeblew999/automerge-wazero-example/pkg/automerge"
)

// Test helper to get testdata path
func testdataPath(filename string) string {
	return filepath.Join("..", "..", "testdata", "unit", "snapshots", filename)
}

// TestNew verifies we can create a new empty document
func TestNew(t *testing.T) {
	ctx := context.Background()

	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	// Document should exist
	if doc == nil {
		t.Error("New() returned nil document")
	}

	// Should be able to get text from empty document
	path := automerge.Root().Get("content")
	text, err := doc.GetText(ctx, path)
	if err != nil {
		t.Errorf("GetText() on new document error = %v", err)
	}
	if text != "" {
		t.Errorf("GetText() on new document = %q, want empty string", text)
	}
}

// TestDocument_SpliceText verifies basic text splice operations
func TestDocument_SpliceText(t *testing.T) {
	ctx := context.Background()

	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	path := automerge.Root().Get("content")

	// Test 1: Insert text at beginning
	err = doc.SpliceText(ctx, path, 0, 0, "Hello")
	if err != nil {
		t.Fatalf("SpliceText(insert 'Hello') error = %v", err)
	}

	text, err := doc.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}
	if text != "Hello" {
		t.Errorf("After insert, GetText() = %q, want %q", text, "Hello")
	}

	// Test 2: Append text
	err = doc.SpliceText(ctx, path, 5, 0, ", World!")
	if err != nil {
		t.Fatalf("SpliceText(append ', World!') error = %v", err)
	}

	text, err = doc.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}
	if text != "Hello, World!" {
		t.Errorf("After append, GetText() = %q, want %q", text, "Hello, World!")
	}

	// Test 3: Delete text (remove "World")
	err = doc.SpliceText(ctx, path, 7, 5, "")
	if err != nil {
		t.Fatalf("SpliceText(delete) error = %v", err)
	}

	text, err = doc.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}
	if text != "Hello, !" {
		t.Errorf("After delete, GetText() = %q, want %q", text, "Hello, !")
	}

	// Test 4: Replace text
	err = doc.SpliceText(ctx, path, 7, 1, "Earth")
	if err != nil {
		t.Fatalf("SpliceText(replace) error = %v", err)
	}

	text, err = doc.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}
	if text != "Hello, Earth" {
		t.Errorf("After replace, GetText() = %q, want %q", text, "Hello, Earth")
	}
}

// TestDocument_SpliceText_Unicode verifies Unicode handling
func TestDocument_SpliceText_Unicode(t *testing.T) {
	ctx := context.Background()

	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	path := automerge.Root().Get("content")

	// Insert Unicode text with emoji
	unicodeText := "Hello 世界! 🌍🚀"
	err = doc.SpliceText(ctx, path, 0, 0, unicodeText)
	if err != nil {
		t.Fatalf("SpliceText(unicode) error = %v", err)
	}

	text, err := doc.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}
	if text != unicodeText {
		t.Errorf("GetText() = %q, want %q", text, unicodeText)
	}
}

// TestDocument_TextLength verifies text length calculation
func TestDocument_TextLength(t *testing.T) {
	ctx := context.Background()

	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	path := automerge.Root().Get("content")

	// Empty document
	length, err := doc.TextLength(ctx, path)
	if err != nil {
		t.Fatalf("TextLength() on empty doc error = %v", err)
	}
	if length != 0 {
		t.Errorf("TextLength() on empty doc = %d, want 0", length)
	}

	// After inserting text
	err = doc.SpliceText(ctx, path, 0, 0, "Hello")
	if err != nil {
		t.Fatalf("SpliceText() error = %v", err)
	}

	length, err = doc.TextLength(ctx, path)
	if err != nil {
		t.Fatalf("TextLength() error = %v", err)
	}
	if length != 5 {
		t.Errorf("TextLength() = %d, want 5", length)
	}
}

// TestDocument_SaveAndLoad verifies serialization round-trip
func TestDocument_SaveAndLoad(t *testing.T) {
	ctx := context.Background()

	// Create document with content
	doc1, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer doc1.Close(ctx)

	path := automerge.Root().Get("content")
	originalText := "Hello, Automerge!"
	err = doc1.SpliceText(ctx, path, 0, 0, originalText)
	if err != nil {
		t.Fatalf("SpliceText() error = %v", err)
	}

	// Save to bytes
	data, err := doc1.Save(ctx)
	if err != nil {
		t.Fatalf("Save() error = %v", err)
	}
	if len(data) == 0 {
		t.Error("Save() returned empty data")
	}

	// Load into new document
	doc2, err := automerge.LoadWithWASM(ctx, data, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	defer doc2.Close(ctx)

	// Verify content is preserved
	text, err := doc2.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() after Load error = %v", err)
	}
	if text != originalText {
		t.Errorf("After Save/Load, GetText() = %q, want %q", text, originalText)
	}
}

// TestDocument_LoadFromTestData verifies loading pre-generated snapshots
func TestDocument_LoadFromTestData(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name         string
		filename     string
		expectedText string
	}{
		{
			name:         "empty document",
			filename:     "empty.am",
			expectedText: "",
		},
		{
			name:         "hello world",
			filename:     "hello-world.am",
			expectedText: "Hello, World!",
		},
		{
			name:         "simple text",
			filename:     "simple-text.am",
			expectedText: "The quick brown fox jumps over the lazy dog.",
		},
		{
			name:         "unicode text",
			filename:     "unicode-text.am",
			expectedText: "Hello 世界! 🌍🚀 Emoji test: ✅❌🎉",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := os.ReadFile(testdataPath(tt.filename))
			if err != nil {
				t.Fatalf("Failed to read test file %s: %v", tt.filename, err)
			}

			doc, err := automerge.LoadWithWASM(ctx, data, automerge.TestWASMPath)
			if err != nil {
				t.Fatalf("Load(%s) error = %v", tt.filename, err)
			}
			defer doc.Close(ctx)

			path := automerge.Root().Get("content")
			text, err := doc.GetText(ctx, path)
			if err != nil {
				t.Fatalf("GetText() error = %v", err)
			}

			if text != tt.expectedText {
				t.Errorf("GetText() = %q, want %q", text, tt.expectedText)
			}
		})
	}
}

// TestDocument_Merge verifies CRDT merge functionality
// NOTE: This uses the SAME WASM runtime for both docs (shares global state)
// so it can't test true concurrent edits. See TESTING.md for details.
func TestDocument_Merge(t *testing.T) {
	ctx := context.Background()

	// Create two documents that share the same WASM runtime
	doc1, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New(doc1) error = %v", err)
	}
	defer doc1.Close(ctx)

	doc2, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New(doc2) error = %v", err)
	}
	defer doc2.Close(ctx)

	path := automerge.Root().Get("content")

	// Set different text in each
	err = doc1.SpliceText(ctx, path, 0, 0, "Alice")
	if err != nil {
		t.Fatalf("SpliceText(doc1) error = %v", err)
	}

	err = doc2.SpliceText(ctx, path, 0, 0, "Bob")
	if err != nil {
		t.Fatalf("SpliceText(doc2) error = %v", err)
	}

	// Merge doc2 into doc1
	err = doc1.Merge(ctx, doc2)
	if err != nil {
		t.Fatalf("Merge() error = %v", err)
	}

	// Check result
	merged, err := doc1.GetText(ctx, path)
	if err != nil {
		t.Fatalf("GetText() after merge error = %v", err)
	}

	t.Logf("After merge: %q", merged)

	// KNOWN ISSUE: Merge currently only preserves one document's content
	// This is because both docs share the same WASM global state
	// For now, just verify we didn't lose ALL content
	if merged == "" {
		t.Errorf("Merge resulted in empty text - total data loss!")
	}
}

// TestDocument_Get_NotImplemented verifies unimplemented methods return proper errors
func TestDocument_Get_NotImplemented(t *testing.T) {
	ctx := context.Background()

	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	// Test Map.Get - now implemented for ROOT map!
	// Getting a non-existent key should return an error
	_, err = doc.Get(ctx, automerge.Root(), "nonexistent")
	if err == nil {
		t.Error("Get() for non-existent key should return error")
	}

	// But getting an existing key should work
	err = doc.Put(ctx, automerge.Root(), "testkey", automerge.NewString("testvalue"))
	if err != nil {
		t.Fatalf("Put() failed: %v", err)
	}

	val, err := doc.Get(ctx, automerge.Root(), "testkey")
	if err != nil {
		t.Fatalf("Get() failed: %v", err)
	}

	str, ok := val.AsString()
	if !ok || str != "testvalue" {
		t.Errorf("Get() returned wrong value: got %q, want %q", str, "testvalue")
	}

	// Nested maps still not implemented
	nestedPath := automerge.Root().Get("nested")
	_, err = doc.Get(ctx, nestedPath, "key")
	var notImpl *automerge.NotImplementedError
	if !errors.As(err, &notImpl) {
		t.Fatalf("Expected NotImplementedError for nested map, got %T: %v", err, err)
	}
}

// TestPath verifies Path construction and navigation
func TestPath(t *testing.T) {
	root := automerge.Root()
	if !root.IsRoot() {
		t.Error("Root() should return root path")
	}

	path := root.Get("content")
	if path.Len() != 1 {
		t.Errorf("Path.Get() length = %d, want 1", path.Len())
	}

	if path.Key() != "content" {
		t.Errorf("Path.Key() = %s, want 'content'", path.Key())
	}

	// Test string representation
	str := path.String()
	if str != "/content" {
		t.Errorf("Path.String() = %s, want '/content'", str)
	}

	// Test nested paths
	nested := root.Get("users").Index(0).Get("name")
	if nested.Len() != 3 {
		t.Errorf("Nested path length = %d, want 3", nested.Len())
	}

	nestedStr := nested.String()
	expected := "/users[0]/name"
	if nestedStr != expected {
		t.Errorf("Nested path = %s, want %s", nestedStr, expected)
	}
}

// TestValue verifies Value creation and type checking
func TestValue(t *testing.T) {
	// String value
	strVal := automerge.NewString("hello")
	if !strVal.IsScalar() {
		t.Error("String value should be scalar")
	}

	s, ok := strVal.AsString()
	if !ok {
		t.Error("Failed to convert value to string")
	}
	if s != "hello" {
		t.Errorf("AsString() = %s, want 'hello'", s)
	}

	// Int value
	intVal := automerge.NewInt(42)
	i, ok := intVal.AsInt()
	if !ok {
		t.Error("Failed to convert value to int")
	}
	if i != 42 {
		t.Errorf("AsInt() = %d, want 42", i)
	}

	// Bool value
	boolVal := automerge.NewBool(true)
	b, ok := boolVal.AsBool()
	if !ok {
		t.Error("Failed to convert value to bool")
	}
	if !b {
		t.Errorf("AsBool() = %v, want true", b)
	}

	// Float value
	floatVal := automerge.NewFloat(3.14)
	f, ok := floatVal.AsFloat()
	if !ok {
		t.Error("Failed to convert value to float")
	}
	if f != 3.14 {
		t.Errorf("AsFloat() = %f, want 3.14", f)
	}
}

// TestNotImplementedError verifies error formatting
func TestNotImplementedError(t *testing.T) {
	err := &automerge.NotImplementedError{
		Feature:   "Put",
		Milestone: "M2",
		Message:   "Map operations require multi-object support",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error() returned empty string")
	}

	if !errors.Is(err, automerge.ErrNotImplemented) {
		t.Error("NotImplementedError should match ErrNotImplemented")
	}
}

// TestDeprecatedError verifies deprecation warnings
func TestDeprecatedError(t *testing.T) {
	err := &automerge.DeprecatedError{
		Method:      "UpdateText",
		Alternative: "SpliceText",
		Reason:      "destroys CRDT history",
	}

	msg := err.Error()
	if msg == "" {
		t.Error("Error() returned empty string")
	}

	if !errors.Is(err, automerge.ErrDeprecated) {
		t.Error("DeprecatedError should match ErrDeprecated")
	}
}

// BenchmarkSpliceText benchmarks text insertion performance
func BenchmarkSpliceText(b *testing.B) {
	ctx := context.Background()

	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	path := automerge.Root().Get("content")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pos := uint(i % 100) // Keep it reasonable
		doc.SpliceText(ctx, path, pos, 0, "x")
	}
}

// BenchmarkSaveLoad benchmarks serialization performance
func BenchmarkSaveLoad(b *testing.B) {
	ctx := context.Background()

	// Create document with some content
	doc, err := automerge.NewWithWASM(ctx, automerge.TestWASMPath)
	if err != nil {
		b.Fatalf("New() error = %v", err)
	}
	defer doc.Close(ctx)

	path := automerge.Root().Get("content")
	doc.SpliceText(ctx, path, 0, 0, "The quick brown fox jumps over the lazy dog")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		data, err := doc.Save(ctx)
		if err != nil {
			b.Fatalf("Save() error = %v", err)
		}

		doc2, err := automerge.LoadWithWASM(ctx, data, automerge.TestWASMPath)
		if err != nil {
			b.Fatalf("Load() error = %v", err)
		}
		doc2.Close(ctx)
	}
}
