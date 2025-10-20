package wazero

import (
	"context"
	"fmt"
)

// Document Lifecycle - maps to rust/automerge_wasi/src/document.rs

// AmInit initializes a new Automerge document with a Text object at ROOT["content"]
func (r *Runtime) AmInit(ctx context.Context) error {
	results, err := r.callExport(ctx, "am_init")
	if err != nil {
		return err
	}
	return checkErrorCode("am_init", results)
}

// AmSaveLen returns the byte size of the serialized document
func (r *Runtime) AmSaveLen(ctx context.Context) (uint32, error) {
	results, err := r.callExport(ctx, "am_save_len")
	if err != nil {
		return 0, err
	}

	return uint32(results[0]), nil
}

// AmSave serializes the document to binary format
func (r *Runtime) AmSave(ctx context.Context) ([]byte, error) {
	// Get save length
	saveLen, err := r.AmSaveLen(ctx)
	if err != nil {
		return nil, err
	}

	if saveLen == 0 {
		return []byte{}, nil
	}

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, saveLen)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, saveLen)

	// Save
	results, err := r.callExport(ctx, "am_save", uint64(ptr))
	if err != nil {
		return nil, err
	}

	if err := checkErrorCode("am_save", results); err != nil {
		return nil, err
	}

	// Read from memory
	mem := r.Memory()
	data, ok := mem.Read(ptr, saveLen)
	if !ok {
		return nil, fmt.Errorf("failed to read save data from WASM memory")
	}

	// Make a copy since data is backed by WASM memory
	result := make([]byte, len(data))
	copy(result, data)

	return result, nil
}

// AmLoad loads a document from binary format
func (r *Runtime) AmLoad(ctx context.Context, data []byte) error {
	dataLen := uint32(len(data))

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, dataLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, dataLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, data) {
		return fmt.Errorf("failed to write data to WASM memory")
	}

	// Load
	results, err := r.callExport(ctx, "am_load", uint64(ptr), uint64(dataLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_load", results)
}

// AmMerge merges another document into the current document (CRDT magic!)
func (r *Runtime) AmMerge(ctx context.Context, otherDoc []byte) error {
	dataLen := uint32(len(otherDoc))

	// Allocate buffer
	ptr, err := r.AmAlloc(ctx, dataLen)
	if err != nil {
		return fmt.Errorf("failed to allocate memory: %w", err)
	}
	defer r.AmFree(ctx, ptr, dataLen)

	// Write to memory
	mem := r.Memory()
	if !mem.Write(ptr, otherDoc) {
		return fmt.Errorf("failed to write data to WASM memory")
	}

	// Merge
	results, err := r.callExport(ctx, "am_merge", uint64(ptr), uint64(dataLen))
	if err != nil {
		return err
	}

	return checkErrorCode("am_merge", results)
}
