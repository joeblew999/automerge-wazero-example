package wazero

import (
	"context"
	"fmt"
)

// History Operations - maps to rust/automerge_wasi/src/history.rs

// AmGetHeadsCount returns the number of heads (frontier) in the document
func (r *Runtime) AmGetHeadsCount(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_get_heads_count")
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmGetHeads retrieves the heads (change hashes) of the document
// Each head is a 32-byte hash
func (r *Runtime) AmGetHeads(ctx context.Context) ([][]byte, error) {
	// Get number of heads
	count, err := r.AmGetHeadsCount(ctx)
	if err != nil {
		return nil, err
	}
	if count == 0 {
		return [][]byte{}, nil
	}

	// Allocate buffer for heads (32 bytes per head)
	bufferSize := count * 32
	ptr, err := r.AmAlloc(ctx, bufferSize)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate heads buffer: %w", err)
	}
	defer r.AmFree(ctx, ptr, bufferSize)

	// Get heads
	results, err := r.callExport(ctx, "am_get_heads", uint64(ptr))
	if err != nil {
		return nil, err
	}
	if err := checkErrorCode("am_get_heads", results); err != nil {
		return nil, err
	}

	// Read heads from memory
	mem := r.Memory()
	headsBytes, ok := mem.Read(ptr, bufferSize)
	if !ok {
		return nil, fmt.Errorf("failed to read heads from WASM memory")
	}

	// Split into individual 32-byte hashes
	heads := make([][]byte, count)
	for i := uint32(0); i < count; i++ {
		hash := make([]byte, 32)
		copy(hash, headsBytes[i*32:(i+1)*32])
		heads[i] = hash
	}

	return heads, nil
}

// AmGetChangesCount returns the number of changes since the given heads
func (r *Runtime) AmGetChangesCount(ctx context.Context, haveHeads [][]byte) (uint32, error) {
	var headsPtr uint32
	headsCount := uint32(len(haveHeads))

	if headsCount > 0 {
		// Allocate buffer for heads (32 bytes per head)
		bufferSize := headsCount * 32
		var err error
		headsPtr, err = r.AmAlloc(ctx, bufferSize)
		if err != nil {
			return 0, fmt.Errorf("failed to allocate heads buffer: %w", err)
		}
		defer r.AmFree(ctx, headsPtr, bufferSize)

		// Write heads to memory
		mem := r.Memory()
		for i, head := range haveHeads {
			if len(head) != 32 {
				return 0, fmt.Errorf("invalid head size: expected 32 bytes, got %d", len(head))
			}
			if !mem.Write(headsPtr+uint32(i*32), head) {
				return 0, fmt.Errorf("failed to write head to WASM memory")
			}
		}
	}

	results, err := r.callExport(ctx, "am_get_changes_count", uint64(headsPtr), uint64(headsCount))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmGetChangesLen returns the total byte size of changes since the given heads
func (r *Runtime) AmGetChangesLen(ctx context.Context, haveHeads [][]byte) (uint32, error) {
	var headsPtr uint32
	headsCount := uint32(len(haveHeads))

	if headsCount > 0 {
		bufferSize := headsCount * 32
		var err error
		headsPtr, err = r.AmAlloc(ctx, bufferSize)
		if err != nil {
			return 0, fmt.Errorf("failed to allocate heads buffer: %w", err)
		}
		defer r.AmFree(ctx, bufferSize, headsPtr)

		mem := r.Memory()
		for i, head := range haveHeads {
			if len(head) != 32 {
				return 0, fmt.Errorf("invalid head size: expected 32 bytes, got %d", len(head))
			}
			if !mem.Write(headsPtr+uint32(i*32), head) {
				return 0, fmt.Errorf("failed to write head to WASM memory")
			}
		}
	}

	results, err := r.callExport(ctx, "am_get_changes_len", uint64(headsPtr), uint64(headsCount))
	if err != nil {
		return 0, err
	}
	return uint32(results[0]), nil
}

// AmGetChanges retrieves changes since the given heads
func (r *Runtime) AmGetChanges(ctx context.Context, haveHeads [][]byte) ([]byte, error) {
	// Get total size needed
	changesLen, err := r.AmGetChangesLen(ctx, haveHeads)
	if err != nil {
		return nil, err
	}
	if changesLen == 0 {
		return []byte{}, nil
	}

	var headsPtr uint32
	headsCount := uint32(len(haveHeads))

	if headsCount > 0 {
		bufferSize := headsCount * 32
		headsPtr, err = r.AmAlloc(ctx, bufferSize)
		if err != nil {
			return nil, fmt.Errorf("failed to allocate heads buffer: %w", err)
		}
		defer r.AmFree(ctx, headsPtr, bufferSize)

		mem := r.Memory()
		for i, head := range haveHeads {
			if !mem.Write(headsPtr+uint32(i*32), head) {
				return nil, fmt.Errorf("failed to write head to WASM memory")
			}
		}
	}

	// Allocate buffer for changes
	changesPtr, err := r.AmAlloc(ctx, changesLen)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate changes buffer: %w", err)
	}
	defer r.AmFree(ctx, changesPtr, changesLen)

	// Get changes
	results, err := r.callExport(ctx, "am_get_changes", uint64(headsPtr), uint64(headsCount), uint64(changesPtr))
	if err != nil {
		return nil, err
	}
	if err := checkErrorCode("am_get_changes", results); err != nil {
		return nil, err
	}

	// Read changes from memory
	mem := r.Memory()
	changesBytes, ok := mem.Read(changesPtr, changesLen)
	if !ok {
		return nil, fmt.Errorf("failed to read changes from WASM memory")
	}

	// Make a copy
	result := make([]byte, len(changesBytes))
	copy(result, changesBytes)
	return result, nil
}

// AmApplyChanges applies changes to the document
func (r *Runtime) AmApplyChanges(ctx context.Context, changes []byte) error {
	if len(changes) == 0 {
		return nil
	}

	changesLen := uint32(len(changes))
	changesPtr, err := r.AmAlloc(ctx, changesLen)
	if err != nil {
		return fmt.Errorf("failed to allocate changes buffer: %w", err)
	}
	defer r.AmFree(ctx, changesPtr, changesLen)

	// Write changes to memory
	mem := r.Memory()
	if !mem.Write(changesPtr, changes) {
		return fmt.Errorf("failed to write changes to WASM memory")
	}

	// Apply changes
	results, err := r.callExport(ctx, "am_apply_changes", uint64(changesPtr), uint64(changesLen))
	if err != nil {
		return err
	}
	return checkErrorCode("am_apply_changes", results)
}
