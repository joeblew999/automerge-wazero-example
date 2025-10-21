// ═══════════════════════════════════════════════════════════════
// LAYER 4: Go High-Level CRDT API (Pure Functions)
//
// Responsibilities:
// - Provide idiomatic Go API for Automerge Text CRDT
// - Pure CRDT operations (NO state, NO mutex, NO persistence)
// - Path-based API for document navigation
// - Error handling and validation
//
// Dependencies:
// ⬇️  Calls: go/pkg/wazero/text.go (Layer 3 - FFI wrappers)
// ⬆️  Called by: go/pkg/server/text.go (Layer 5 - adds state + thread safety)
//
// Related Files:
// 🔁 Siblings: map.go, list.go, counter.go, sync.go, richtext.go
// 📝 Tests: text_test.go (CRDT property tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-4-go-api
//
// Design Note:
// This layer is STATELESS - it doesn't own the Document or manage
// concurrency. Layer 5 (server) adds mutex protection and persistence.
// ═══════════════════════════════════════════════════════════════

package automerge

import "context"

// GetText retrieves the text content at the given path.
//
// Currently only supports the root "content" text object created by New().
// Will be extended to support arbitrary paths in M2 (multi-document support).
//
// Status: ✅ Implemented
func (d *Document) GetText(ctx context.Context, path Path) (string, error) {
	// Currently we only have one text object at root["content"]
	if !d.isContentPath(path) {
		return "", ErrInvalidPath
	}

	return d.runtime.AmGetText(ctx)
}

// SpliceText performs a proper CRDT splice operation on text.
//
// This is the CORRECT way to edit text - it maintains fine-grained CRDT history
// which enables better conflict resolution when merging.
//
// Parameters:
//   - path: Path to text object (currently only Root().Get("content"))
//   - pos: Character position (0-indexed)
//   - del: Number of characters to delete (can be 0)
//   - text: Text to insert (can be empty)
//
// Examples:
//   // Insert "Hello" at the beginning
//   doc.SpliceText(ctx, Root().Get("content"), 0, 0, "Hello")
//
//   // Delete 5 characters at position 0
//   doc.SpliceText(ctx, Root().Get("content"), 0, 5, "")
//
//   // Replace 5 characters at position 0 with "Hi"
//   doc.SpliceText(ctx, Root().Get("content"), 0, 5, "Hi")
//
// Status: ✅ WASI export exists, ✅ Go wrapper implemented
func (d *Document) SpliceText(ctx context.Context, path Path, pos uint, del int, text string) error {
	if !d.isContentPath(path) {
		return ErrInvalidPath
	}

	return d.runtime.AmTextSplice(ctx, pos, int64(del), text)
}

// UpdateText replaces all text content.
//
// DEPRECATED: Use SpliceText for better CRDT merging.
// This method deletes all existing text and inserts new text, which destroys
// the fine-grained edit history. When two users concurrently call UpdateText,
// one edit will completely overwrite the other.
//
// Status: ✅ Implemented but deprecated
func (d *Document) UpdateText(ctx context.Context, path Path, newText string) error {
	if !d.isContentPath(path) {
		return ErrInvalidPath
	}

	// Return deprecation warning
	err := &DeprecatedError{
		Method:      "UpdateText",
		Alternative: "SpliceText",
		Reason:      "destroys fine-grained CRDT history",
	}

	// Still execute the operation for backward compatibility
	if execErr := d.runtime.AmSetText(ctx, newText); execErr != nil {
		return execErr
	}

	// Return deprecation warning (non-fatal)
	return err
}

// TextLength returns the number of UTF-8 bytes in the text
//
// Note: This returns byte length, not character count. For character count,
// use len([]rune(doc.GetText(...)))
//
// Status: ✅ Implemented
func (d *Document) TextLength(ctx context.Context, path Path) (uint32, error) {
	if !d.isContentPath(path) {
		return 0, ErrInvalidPath
	}

	return d.runtime.AmGetTextLen(ctx)
}

// isContentPath checks if path points to the root "content" text object
// In M2, this will be generalized to support arbitrary text objects
func (d *Document) isContentPath(path Path) bool {
	// Currently only support Root().Get("content")
	if path.Len() != 1 {
		return false
	}
	return path.Key() == "content"
}
