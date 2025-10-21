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
// ⬇️  Calls: WASM functions (am_text_splice, am_get_text, etc)
//           Implemented in: rust/automerge_wasi/src/text.rs (Layer 2)
// ⬆️  Called by: go/pkg/automerge/text.go (Layer 4 - high-level API)
//
// Related Files:
// 🔁 Siblings: map.go, list.go, counter.go, sync.go, richtext.go
// 📝 Tests: text_test.go (FFI boundary tests)
// 🔗 Docs: docs/explanation/architecture.md#layer-3-go-ffi
// ═══════════════════════════════════════════════════════════════

package wazero

import (
	"context"
	"fmt"
)

// Text Operations - maps to rust/automerge_wasi/src/text.rs

// AmTextSplice performs a proper Text CRDT splice operation
func (r *Runtime) AmTextSplice(ctx context.Context, pos uint, del int64, text string) error {
	textBytes := []byte(text)
	textLen := uint32(len(textBytes))

	var textPtr uint32
	var err error

	if textLen > 0 {
		// Allocate memory for text
		textPtr, err = r.AmAlloc(ctx, textLen)
		if err != nil {
			return fmt.Errorf("failed to allocate memory for text: %w", err)
		}
		defer r.AmFree(ctx, textPtr, textLen)

		// Write text to WASM memory
		mem := r.Memory()
		if !mem.Write(textPtr, textBytes) {
			return fmt.Errorf("failed to write text to WASM memory")
		}
	}

	// Call am_text_splice
	results, err := r.callExport(ctx, "am_text_splice",
		uint64(pos),
		uint64(del),
		uint64(textPtr),
		uint64(textLen),
	)
	if err != nil {
		return err
	}

	return checkErrorCode("am_text_splice", results)
}

// AmSetText replaces all text content (DEPRECATED - use AmTextSplice)
func (r *Runtime) AmSetText(ctx context.Context, text string) error {
	textBytes := []byte(text)
	textLen := uint32(len(textBytes))

	// Allocate memory
	ptr, err := r.AmAlloc(ctx, textLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, textLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, textBytes) {
		return fmt.Errorf("failed to write to WASM memory")
	}

	// Call am_set_text
	results, err := r.callExport(ctx, "am_set_text", uint64(ptr), uint64(textLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_set_text", results)
}

// AmGetTextLen returns the byte length of the current text content
func (r *Runtime) AmGetTextLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_get_text_len")
	if err != nil {
		return 0, err
	}

	return uint32(results[0]), nil
}

// AmGetText retrieves the current text content
func (r *Runtime) AmGetText(ctx context.Context) (string, error) {
	// Get text length
	textLen, err := r.AmGetTextLen(ctx)
	if err != nil {
		return "", err
	}

	if textLen == 0 {
		return "", nil
	}

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, textLen)
	if err != nil {
		return "", fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, textLen)

	// Get text
	results, err := r.callExport(ctx, "am_get_text", uint64(ptr))
	if err != nil {
		return "", err
	}

	if err := checkErrorCode("am_get_text", results); err != nil {
		return "", err
	}

	// Read from memory
	mem := r.Memory()
	data, ok := mem.Read(ptr, textLen)
	if !ok {
		return "", fmt.Errorf("failed to read text from WASM memory")
	}

	return string(data), nil
}
